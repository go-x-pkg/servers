package servers

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	xlog "github.com/go-x-pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

type fnLogHTTPError struct {
	logger *xlog.FnT
}

func (fw *fnLogHTTPError) Write(raw []byte) (n int, err error) {
	msg := strings.TrimSuffix(string(raw), "\n")
	fw.logger.Log(xlog.Error, "gin server", zap.Error(errors.New(msg)))
	return len(raw), nil
}

type ServerListener struct {
	Server
	net.Listener
}

func (sl *ServerListener) Addr() string { return sl.Server.Addr() }

func (it iterator) Listen(fnArgs ...Arg) (ss Servers, errs []error) {
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
			network := s.Network()

			if s.Kind().Has(KindUNIX) {
				unix := s.(*ServerUNIX)
				mode := unix.SocketFileMode

				os.Remove(addr)
				listener, err := net.Listen(network, addr)
				if err != nil {
					errsChan <- fmt.Errorf("listen %s (%s) server failed: %w", network, addr, err)
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

				fnLog(xlog.Info, `{"status": "chmod OK", "perms": "%03o | %s", "addr": %q, "cmd": "chmod %o %s"}`,
					mode.Perm(), mode, addr, mode.Perm(), addr)

				serversChan <- &ServerListener{Server: s, Listener: listener}
			} else {
				listener, err := net.Listen(network, addr)
				if err != nil {
					errsChan <- fmt.Errorf("listen %s (%s) server failed: %w", network, addr, err)
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

func (it iterator) ServeHTTP(fnNewHandler func(Server) http.Handler, fnArgs ...Arg) error {
	it = it.FilterListener()

	cfg := args{}
	cfg.defaultize()

	for _, fn := range fnArgs {
		fn(&cfg)
	}

	ctx := cfg.ctx
	if ctx == nil {
		ctx = context.TODO()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fnLog := cfg.fnLog

	errChan := make(chan error, it.Len())

	wg := sync.WaitGroup{}
	wg.Add(it.Len())

	fnOnErr := func(err error) {
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}
	it(func(s Server) bool {
		l := s.(*ServerListener)

		go func(l *ServerListener) {
			defer wg.Done()

			addr := l.Addr()

			fnLog(xlog.Info, "%s HTTP server starting on %s", runLogPrefix(l), addr)

			server := &http.Server{
				Addr:     addr,
				Handler:  fnNewHandler(l.Server),
				ErrorLog: log.New(&fnLogHTTPError{&cfg.fnLogHTTPError}, "", 0),

				// see: Potential slowloris attack GO-S2112
				ReadHeaderTimeout: s.Base().HTTP.ReadHeaderTimeout,
			}

			go func() {
				<-ctx.Done()

				ctxTimeout, cancel := context.WithTimeout(context.Background(), cfg.fnShutdownTimeout())
				defer cancel()

				if err := server.Shutdown(ctxTimeout); err != nil {
					fnLog(xlog.Info, "server (:addr %s) shutdown failed: %s", addr, err)

					return
				}

				fnLog(xlog.Info, "%s HTTP server (:addr %s) shutdown OK", runLogPrefix(l), addr)
			}()

			if s.Kind().Has(KindUNIX) {
				if err := server.Serve(l.Listener); err != nil {
					fnOnErr(fmt.Errorf("serve unix (%s) failed: %w", addr, err))
				}
			} else {
				inet := l.Server.(*ServerINET)

				tlsConfig, err := inet.newTLSConfig()
				if err != nil {
					fnOnErr(err)
					return
				}

				if tlsConfig != nil {
					l.Listener = tls.NewListener(l.Listener, tlsConfig)
					server.TLSConfig = tlsConfig
				}

				if err := server.Serve(l.Listener); err != nil {
					serverType := "http"
					if inet.TLS.Enable {
						serverType = "https"
					}
					fnOnErr(fmt.Errorf("starting %s (%s) server failed: %w", serverType, addr, err))
				}
			}
		}(l)

		return true
	})

	done := make(chan struct{})

	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		return nil

	case err := <-errChan:
		cancel()

		deadline := time.NewTimer(cfg.fnShutdownTimeout())
		defer func() {
			if !deadline.Stop() {
				select {
				case <-deadline.C:
				default:
				}
			}
		}()

		select {
		case <-done:
		case <-deadline.C:
		}

		return err
	}
}

func (it iterator) ServeGRPC(fnNewServer func(s Server, opts ...grpc.ServerOption) *grpc.Server, fnArgs ...Arg) error {
	it = it.FilterListener()

	cfg := args{}
	cfg.defaultize()

	for _, fn := range fnArgs {
		fn(&cfg)
	}

	ctx := cfg.ctx
	if ctx == nil {
		ctx = context.TODO()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fnLog := cfg.fnLog

	errChan := make(chan error, it.Len())

	wg := sync.WaitGroup{}
	wg.Add(it.Len())

	fnOnErr := func(err error) {
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}

	it(func(s Server) bool {
		l := s.(*ServerListener)

		go func(l *ServerListener) {
			defer wg.Done()

			inet := l.Server.(*ServerINET)

			addr := l.Addr()

			var opts []grpc.ServerOption

			fnLog(xlog.Info, "%s gRPC server starting on %s", runLogPrefix(l), addr)

			tlsConfig, err := inet.newTLSConfig()
			if err != nil {
				fnOnErr(err)
				return
			}

			if tlsConfig != nil {
				opt := grpc.Creds(credentials.NewTLS(tlsConfig))
				opts = append(opts, opt)
			}

			grpclog.SetLoggerV2(newXGRPCLogger(cfg.logGRPC))

			server := fnNewServer(l.Server, opts...)

			if inet.GRPC.Reflection {
				reflection.Register(server)
			}

			go func() {
				<-ctx.Done()

				server.GracefulStop()
				fnLog(xlog.Info, "%s gRPC server (:addr %s) shutdown OK", runLogPrefix(l), addr)
			}()

			if err := server.Serve(l.Listener); err != nil {
				fnOnErr(fmt.Errorf("starting gRPC (%s) server failed: %w", addr, err))
			}
		}(l)

		return true
	})

	done := make(chan struct{})

	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		return nil

	case err := <-errChan:
		cancel()

		deadline := time.NewTimer(cfg.fnShutdownTimeout())
		defer func() {
			if !deadline.Stop() {
				select {
				case <-deadline.C:
				default:
				}
			}
		}()

		select {
		case <-done:
		case <-deadline.C:
		}

		return err
	}
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
