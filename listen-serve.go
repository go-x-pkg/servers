package servers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-x-pkg/log"
)

type ServerListener struct {
	Server
	net.Listener
}

func (sl *ServerListener) Addr() string { return sl.Server.Addr() }

func (it iterator) Listen(fnArgs ...arg) (ss Servers, errs []error) {
	cfg := args{}
	cfg.defaultize()

	for _, fn := range fnArgs {
		fn(&cfg)
	}

	fnLog := cfg.fnLog

	wg := sync.WaitGroup{}
	wg.Add(it.Len())

	errsChan := make(chan error, it.Len())
	serversChan := make(chan Server, it.Len())

	it(func(s Server) bool {
		go func(s Server) {
			defer wg.Done()

			addr := s.Addr()

			if s.Kind().Has(KindUNIX) {
				unix := s.(*ServerUNIX)
				mode := unix.SocketFileMode

				os.Remove(addr)
				listener, err := net.Listen("unix", addr)
				if err != nil {
					errsChan <- fmt.Errorf("listen unix (%s) server failed: %w", addr, err)
					return
				}

				if err := os.Chmod(addr, 0o777); err != nil {
					errsChan <- fmt.Errorf("chmod on unix socket (%s) failed: %w", addr, err)

					listener.Close()

					return
				}

				<-time.After(1 * time.Second)

				if err := os.Chmod(addr, mode); err != nil {
					errsChan <- fmt.Errorf("changing permissions to unix socket %s failed: %w", addr, err)

					listener.Close()

					return
				}

				fnLog(log.Info, `{"status": "chmod OK", "perms": "%03o | %s", "addr": %q, "cmd": "chmod %o %s"}`,
					mode.Perm(), mode, addr, mode.Perm(), addr)

				serversChan <- &ServerListener{Server: s, Listener: listener}
			} else {
				listener, err := net.Listen("tcp", addr)
				if err != nil {
					errsChan <- fmt.Errorf("listen tcp (%s) server failed: %w", addr, err)
					return
				}

				serversChan <- &ServerListener{Server: s, Listener: listener}
			}
		}(s)

		return true
	})

	go func() { wg.Wait(); close(errsChan); close(serversChan) }()

	for s := range serversChan {
		ss = append(ss, serverEnsureWrapped(s))
	}

	for err := range errsChan {
		errs = append(errs, err)
	}

	return ss, errs
}

func (it iterator) ServeHTTP(handler http.Handler, fnArgs ...arg) (chan struct{}, chan error) {
	it = it.FilterListener()

	cfg := args{}
	cfg.defaultize()

	for _, fn := range fnArgs {
		fn(&cfg)
	}

	ctx := cfg.ctx
	fnLog := cfg.fnLog

	errs := make(chan error, it.Len())
	servers := make(chan *http.Server, it.Len())

	// all servers shutdown
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

	it.FilterListener()(func(s Server) bool {
		l := s.(*ServerListener)

		go func(l *ServerListener) {
			addr := l.Addr()

			fnLog(log.Info, "%s HTTP server starting on %s", runLogPrefix(l), addr)

			// create multiple servers with addr
			// instead single server for better shutdown
			// logging
			// 1 server == 1 listener
			server := &http.Server{
				Addr:    addr,
				Handler: handler,
			}

			servers <- server

			if s.Kind().Has(KindUNIX) {
				if err := server.Serve(l.Listener); err != nil {
					fnOnErr(fmt.Errorf("serve unix (%s) failed: %w", addr, err))
					return
				}
			} else {
				inet := l.Server.(*ServerINET)

				if inet.TLS.Enable {
					if err := server.ServeTLS(l.Listener, inet.TLS.CertFile, inet.TLS.KeyFile); err != nil {
						fnOnErr(fmt.Errorf("starting https (%s) server failed: %w", addr, err))
						return
					}
				} else {
					if err := server.Serve(l.Listener); err != nil {
						fnOnErr(fmt.Errorf("starting http (%s) server failed: %w", addr, err))
						return
					}
				}
			}
		}(l)

		return true
	})

	return done, errs
}

func (it iterator) Close() (errs []error) {
	it = it.FilterListener()

	wg := sync.WaitGroup{}
	wg.Add(it.Len())

	errsChan := make(chan error, it.Len())

	it(func(s Server) bool {
		l := s.(*ServerListener)

		go func(l *ServerListener) {
			defer wg.Done()

			errsChan <- l.Close()
		}(l)

		return true
	})

	go func() { wg.Wait(); close(errsChan) }()

	for err := range errsChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
