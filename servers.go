package servers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-x-pkg/dumpctx"
)

type (
	iterCb   func(Server) bool
	iterator func(iterCb) bool

	// see https://doc.rust-lang.org/std/iter/trait.IntoIterator.html.
	intoIterator interface {
		IntoIter() iterator
	}
)

func (it iterator) IntoIter() iterator { return it }

func (it iterator) Len() (length int) {
	it(func(Server) bool {
		length++
		return true
	})

	return
}

func (it iterator) First() (srv Server) {
	it(func(s Server) bool {
		srv = s
		return false
	})

	return srv
}

func (it iterator) Filter(take func(Server) bool) iterator {
	return iterator(func(cb iterCb) bool {
		return it(func(s Server) bool {
			if take(s) {
				if !cb(s) {
					return false
				}
			}
			return true
		})
	})
}

func takeUnix(s Server) bool { return s.Kind().Has(KindUNIX) }
func takeInet(s Server) bool { return s.Kind().Has(KindINET) }
func takeHTTP(s Server) bool { return s.Kind().Has(KindHTTP) }

func (it iterator) FilterUnix() iterator { return it.Filter(takeUnix) }
func (it iterator) FilterInet() iterator { return it.Filter(takeInet) }
func (it iterator) FilterHTTP() iterator { return it.Filter(takeHTTP) }
func (it iterator) FilterListener() iterator {
	return it.Filter(func(s Server) bool {
		_, ok := s.(*ServerListener)
		return ok
	})
}

func (it iterator) Defaultize(inetHost string, inetPort int, unixAddr string) (err error) {
	it(func(s Server) bool {
		if defaultizer, ok := s.(serverDefaultizer); ok {
			if e := defaultizer.defaultize(); e != nil {
				err = e
				return false
			}
		}

		return true
	})

	if err != nil {
		return err
	}

	if s := it.FilterUnix().First(); s != nil {
		unix := s.(*ServerUNIX)

		if unix.Addr() == "" {
			unix.Address = unixAddr
		}
	}

	if s := it.FilterInet().First(); s != nil {
		inet := s.(*ServerINET)

		if inet.Host == "" {
			inet.Host = inetHost
		}

		if inet.Port == 0 {
			inet.Port = inetPort
		}

	}

	return err
}

func (it iterator) Validate() (err error) {
	it(func(s Server) bool {
		if validator, ok := s.(serverValidator); ok {
			if e := validator.validate(); e != nil {
				err = e
				return false
			}
		}

		return true
	})

	return err
}

func (it iterator) Dump(ctx *dumpctx.Ctx, w io.Writer) {
	if ln := it.Len(); ln != 0 {
		fmt.Fprintf(w, " (x%d):\n", ln)

		it(func(s Server) bool {
			ctx.WrapList(func() {
				ctx.EmitPrefix(w)

				fmt.Fprintf(w, "kind: %s\n", s.Kind())

				ctx.Enter()
				defer ctx.Leave()

				if dumper, ok := s.(serverDumper); ok {
					dumper.Dump(ctx, w)
				}
			})

			return true
		})
	} else {
		fmt.Fprint(w, ": ~\n")
	}
}

type Servers []*ServerWrapped

func (ss Servers) IntoIter() iterator { return ss.ForEach }

func (ss Servers) ForEach(cb iterCb) bool {
	for _, s := range ss {
		if !cb(s.Server) {
			return false
		}
	}

	return true
}

func (ss Servers) Defaultize(inetHost string, inetPort int, unixAddr string) (err error) {
	return ss.IntoIter().Defaultize(inetHost, inetPort, unixAddr)
}

func (ss Servers) Dump(ctx *dumpctx.Ctx, w io.Writer) { ss.IntoIter().Dump(ctx, w) }

func (ss Servers) Validate() error { return ss.IntoIter().Validate() }

func (ss *Servers) PushINETIfNotExists(host string, port int) {
	if ss.IntoIter().FilterInet().Len() != 0 {
		return
	}

	s := &ServerINET{
		ServerBase: ServerBase{
			WithKind: WithKind{
				Knd: KindINET,
			},
		},

		Host: host,
		Port: port,
	}

	*ss = append(*ss, serverEnsureWrapped(s))
}

func (ss *Servers) PushUnixIfNotExists(addr string) {
	if ss.IntoIter().FilterUnix().Len() != 0 {
		return
	}

	s := &ServerUNIX{
		ServerBase: ServerBase{
			WithKind: WithKind{
				Knd: KindUNIX,
			},
		},

		Address: addr,
	}

	*ss = append(*ss, serverEnsureWrapped(s))
}

func (ss *Servers) SetPortToFirstINET(port int) {
	ss.
		IntoIter().
		FilterInet().
		First().(*ServerINET).
		setPort(port)
}

func (ss *Servers) Listen(fnArgs ...arg) (Servers, []error) {
	return ss.IntoIter().Listen(fnArgs...)
}

func (ss *Servers) ServeHTTP(handler http.Handler, fnArgs ...arg) (chan struct{}, chan error) {
	return ss.IntoIter().ServeHTTP(handler, fnArgs...)
}

func (ss *Servers) Close() []error { return ss.IntoIter().Close() }
