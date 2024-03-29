package servers_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/go-x-pkg/dumpctx"
	"github.com/go-x-pkg/servers"
	"gopkg.in/yaml.v2"
)

func TestServers(t *testing.T) {
	tests := []struct {
		raw string
	}{
		{`- host: 0.0.0.0
  port: 8000`},

		{`- kind: inet
  host: 0.0.0.0
  port: 8000`},

		{`- kind: inet
  host: 0.0.0.0
  port: 8443
  tls:
    enable: false
    certFile: "/etc/acme/tls.cert"
    keyFile: "/etc/acme/tls.key"`},

		{`- kind: inet
  host: 0.0.0.0
  port: 443
  tls:
    enable: true
    certFile: "/etc/acme/tls.cert"
    keyFile: "/etc/acme/tls.key"`},

		{`- kind: inet
  host: 0.0.0.0
  port: 443
  tls:
    enable: true
    certFile: "/etc/acme/tls.cert"
    keyFile: "/etc/acme/tls.key"
    preferServerCipherSuites: false`},

		{`- kind: inet
  host: 0.0.0.0
  port: 443
  tls:w
    enable: true
    certFile: "/etc/acme/tls.cert"
    keyFile: "/etc/acme/tls.key"
  clientAuth:
    tls:
      enable: true
      authType: require-and-verify-client-cert`},

		{`- kind: unix
  addr: /run/acme/acme.sock`},

		{`- kind: [unix, http]
  addr: /run/acme/acme.sock`},

		{`- kind: [unix, http]
  addr: /run/acme/acme.sock
- kind: inet
  host: 0.0.0.0
  port: 8000
- kind: inet
  host: 0.0.0.0
  port: 8443
  http:
    readHeaderTimeout: 10s
  tls:
    enable: false
    certFile: "/etc/acme/tls.cert"
    keyFile: "/etc/acme/tls.key"`},

		{`- kind: [inet, grpc]
  host: 0.0.0.0
  port: 8000
  tls:
    enable: false
    certFile: "/etc/acme/tls.cert"
    keyFile: "/etc/acme/tls.key"`},

		{`- kind: unix
- kind: inet`},
	}

	for i, tt := range tests {
		func() {
			var ss servers.Servers

			if err := yaml.Unmarshal([]byte(tt.raw), &ss); err != nil {
				t.Errorf("%d: err unmarshal yaml: %s", i, err)
				return
			}

			ss.Defaultize("0.0.0.0", 80, "/run/foo/bar.sock")

			w := bytes.Buffer{}

			dctx := dumpctx.Ctx{}
			dctx.Init()

			fmt.Fprintf(&w, "%sservers", dctx.Indent())
			ss.Dump(&dctx, &w)

			t.Logf("%d:\n%s\n", i, w.String())

			if err := ss.Validate(); err != nil {
				t.Logf("%d: invalid server configuration: %s", i, err)
				return
			}

			ss.SetPortToFirstINET(9000)
			ss.PushINETIfNotExists("10.1.1.2", 9001, servers.KindGRPC)
			ss.PushUnixIfNotExists("/run/bar/buz.sock", servers.KindHTTP)
		}()
	}
}
