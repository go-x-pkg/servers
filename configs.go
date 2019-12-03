package servers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-x-pkg/dumpctx"
	"github.com/go-x-pkg/log"
)

type Configs []*Config

func (cfgs *Configs) PushINETIfNotExists(host string, port int) {
	for _, c := range *cfgs {
		if c.kind == KindINET {
			return
		}
	}

	that := &Config{
		kind: KindINET,
		inner: &ConfigINET{
			Host: host,
			Port: port,
		},
	}

	*cfgs = append(*cfgs, that)
}

func (cfgs *Configs) PushUnixIfNotExists(addr string) {
	for _, c := range *cfgs {
		if c.kind == KindUNIX {
			return
		}
	}

	that := &Config{
		kind: KindUNIX,
		inner: &ConfigUNIX{
			Addr: addr,
		},
	}

	*cfgs = append(*cfgs, that)
}

func (cfgs *Configs) SetPortToFirstINET(port int) {
	for _, c := range *cfgs {
		if configSetPort(c, port) {
			return
		}
	}
}

func (cfgs Configs) Defaultize(inetHost string, inetPort int, unixAddr string) error {
	doneUNIX := false
	doneINET := false

	for _, c := range cfgs {
		if e := c.inner.Defaultize(); e != nil {
			return e
		}

		if c.kind == KindUNIX && !doneUNIX {
			that := c.inner.(*ConfigUNIX)
			if that.Addr == "" {
				that.Addr = unixAddr
			}

			doneUNIX = true
		}

		if c.kind == KindINET && !doneINET {
			that := c.inner.(*ConfigINET)
			if that.Host == "" {
				that.Host = inetHost
			}

			if that.Port == 0 {
				that.Port = inetPort
			}

			doneINET = true
		}
	}

	return nil
}

func (cfgs Configs) Validate() error {
	for _, c := range cfgs {
		if e := c.Validate(); e != nil {
			return e
		}
	}

	return nil
}

func (cfgs Configs) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	if len(cfgs) != 0 {
		fmt.Fprintf(w, " (x%d):\n", len(cfgs))

		for _, r := range cfgs {
			ctx.WrapList(func() { r.Dump(ctx, w) })
		}
	} else {
		fmt.Fprint(w, ": ~\n")
	}
}

type ListenAndServeConfig struct {
	fnShutdownTimeout func() time.Duration
	fnLog             log.FnT
	ctx               context.Context
}

func (cfg *ListenAndServeConfig) defaultize() {
	cfg.fnShutdownTimeout = defaultFnShutdownTimeout
	cfg.fnLog = defaultFnLog
	cfg.ctx = context.TODO()
}

type Arg func(*ListenAndServeConfig)

func FnShutdownTimeout(v func() time.Duration) Arg {
	return func(cfg *ListenAndServeConfig) { cfg.fnShutdownTimeout = v }
}

func FnLog(v log.FnT) Arg {
	return func(cfg *ListenAndServeConfig) { cfg.fnLog = v }
}

func Context(v context.Context) Arg {
	return func(cfg *ListenAndServeConfig) { cfg.ctx = v }
}

func (cfgs Configs) ListenAndServe(handler http.Handler, fnArgs ...Arg) (chan struct{}, chan error) {
	cfg := ListenAndServeConfig{}
	cfg.defaultize()

	for _, fn := range fnArgs {
		fn(&cfg)
	}

	ctx := cfg.ctx
	fnLog := cfg.fnLog

	errs := make(chan error, len(cfgs))
	servers := make(chan *http.Server, len(cfgs))
	done := make(chan struct{}, 1)

	fnOnErr := func(err error) {
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}

	go func() {
		<-ctx.Done()

		defer func() { done <- struct{}{} }()

		for {
			select {
			case server := <-servers:
				ctxTimeout, cancel := context.WithTimeout(context.Background(), cfg.fnShutdownTimeout())
				defer cancel()

				if err := server.Shutdown(ctxTimeout); err != nil {
					fnLog(log.Info, "server shutdown failed: %s", err)
				} else {
					fnLog(log.Info, "server (:addr %s) shutdown OK", server.Addr)
				}
			default: // no more servers to shutdown - all done
				return
			}
		}
	}()

	for _, c := range cfgs {
		go func(c *Config) {
			fnLog(log.Info, "%s server starting on %s", c.RunLogPrefix(), c.inner.GetAddr())

			switch c.kind {
			case KindINET:
				that := c.inner.(*ConfigINET)
				addr := that.GetAddr()

				server := &http.Server{
					Addr:    addr,
					Handler: handler,
				}

				servers <- server

				if that.TLS.Enable {
					if err := server.ListenAndServeTLS(that.TLS.CertFile, that.TLS.KeyFile); err != nil {
						fnOnErr(fmt.Errorf("starting https (%s) server failed: %w", addr, err))
						return
					}
				} else {
					if err := server.ListenAndServe(); err != nil {
						fnOnErr(fmt.Errorf("starting http (%s) server failed: %w", addr, err))
						return
					}
				}

			case KindUNIX:
				that := c.inner.(*ConfigUNIX)
				addr := that.GetAddr()
				mode := that.SocketFileMode

				os.Remove(addr)
				listener, err := net.Listen("unix", addr)
				if err != nil {
					fnOnErr(fmt.Errorf("listen unix (%s) server failed: %w", addr, err))
					return
				}
				defer listener.Close()
				if err := os.Chmod(addr, 0777); err != nil {
					fnOnErr(fmt.Errorf("chmod on unix socket (%s) failed: %w", addr, err))
					return
				}

				go func() {
					<-time.After(1 * time.Second)
					if err := os.Chmod(addr, mode); err != nil {
						fnOnErr(fmt.Errorf("changing permissions to unix socket %s failed: %w", addr, err))
					} else {
						fnLog(log.Info, `{"status": "chmod OK", "perms": "%03o | %s", "addr": %q, "cmd": "chmod %o %s"}`,
							mode.Perm(), mode, addr, mode.Perm(), addr)
					}
				}()

				server := &http.Server{
					Addr:    addr,
					Handler: handler,
				}

				servers <- server

				if err := server.Serve(listener); err != nil {
					fnOnErr(fmt.Errorf("serve unix (%s) failed: %w", addr, err))
					return
				}
			}
		}(c)
	}

	return done, errs
}
