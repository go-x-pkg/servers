# servers

[![GoDev][godev-image]][godev-url]
[![Build Status][build-image]][build-url]
[![Coverage Status][coverage-image]][coverage-url]
[![Go Report Card][goreport-image]][goreport-url]
[![Deepsource.io Card][deepsource-io-image]][deepsource-io-url]

HTTP, HTTPs, gRPC, UNIX servers with configs and context

## Quick-start

```go
package main

import (
  "bytes"
  "context"
  "fmt"
  "http"
  "log"
  "sync"

  "google.golang.org/grpc"
  "github.com/go-x-pkg/servers"
)

func main() {
  var s servers.Servers

  config := `- kind: [unix, http]
  addr: /run/acme/acme.sock
- kind: inet
  host: 0.0.0.0
  port: 8000
- kind: [inet, grpc]
  host: 0.0.0.0
  port: 8443
  tls:
    enable: true
    cert-file: /etc/acme/tls.cert
    key-file: /etc/acme/tls.key
  grpc:
    # e.g. grpcurl
    reflection: true`

  if err := yaml.Unmarshal([]byte(config), &ss); err != nil {
    log.Fatalf("error unmarshal: %s", err)
  }

  listeners, err := ss.Listen()
  if len(errs) != 0 {
    listeners.Close()

    log.Fatalf("error listen: %#v", errs)
  }

  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
    log.Printf("[<] %#v %#v", w, r)
  })

  ctx := context.Background()
  ctx, cancel := context.WithCancel(ctx)
  defer cancel()

  wg := sync.WaitGroup()

  wg.Add(1)
  go func() {
    defer wg.Done()

    err := listeners.ServeHTTP(
      http.DefaultServerMux,

      servers.Context(ctx),
    )
  }()

  wg.Add(1)
  go func() {
    defer wg.Done()

    err := listeners.ServeGRPC(
      func (*grpc.Server) {
        # register gRPC servers here
      },

      servers.Context(ctx),
    )
  }()

  wg.Wait()
}
```

## Config example

```yaml
- kind: [unix, http]
  addr: /run/acme/acme.sock

- kind: inet
  host: 0.0.0.0
  port: 8000

- kind: [inet, grpc]
  host: 0.0.0.0
  port: 8443
  tls:
    enable: true
    cert-file: /etc/acme/tls.cert
    key-file: /etc/acme/tls.key
  grpc:
    # e.g. grpcurl
    reflection: true
```

[godev-image]: https://img.shields.io/badge/go.dev-reference-5272B4?logo=go&logoColor=white
[godev-url]: https://pkg.go.dev/github.com/go-x-pkg/servers

[build-image]: https://app.travis-ci.com/go-x-pkg/servers.svg?branch=master
[build-url]: https://app.travis-ci.com/github/go-x-pkg/servers

[coverage-image]: https://coveralls.io/repos/github/go-x-pkg/servers/badge.svg?branch=master
[coverage-url]: https://coveralls.io/github/go-x-pkg/servers?branch=master

[goreport-image]: https://goreportcard.com/badge/github.com/go-x-pkg/servers
[goreport-url]: https://goreportcard.com/report/github.com/go-x-pkg/servers

[deepsource-io-image]: https://static.deepsource.io/deepsource-badge-light-mini.svg
[deepsource-io-url]: https://deepsource.io/gh/go-x-pkg/servers/?ref=repository-badge
