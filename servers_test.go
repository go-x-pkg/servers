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
    cert-file: "/etc/acme/tls.cert"
    key-file: "/etc/acme/tls.key"`},

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
  tls:
    enable: false
    cert-file: "/etc/acme/tls.cert"
    key-file: "/etc/acme/tls.key"`},

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
			ss.PushINETIfNotExists("10.1.1.2", 9001)
			ss.PushUnixIfNotExists("/run/bar/buz.sock")
		}()
	}
}
