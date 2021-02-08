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
  "fmt"
  "http"
  "log"

  "github.com/go-x-pkg/servers"
)

func main() {
  var s servers.Servers

  config := `- kind: [unix, http]
  addr: /run/acme/acme.sock
- kind: inet
  host: 0.0.0.0
  port: 8000
- kind: inet
  host: 0.0.0.0
  port: 8443`

  if err := yaml.Unmarshal([]byte(config), &ss); err != nil {
    log.Fatalf("error unmarshal: %s", err)
  }

  listeners, err := ss.IntoIter().Listen()
  if len(errs) != 0 {
    listeners.IntoIter().Close()

    log.Fatalf("error listen: %#v", errs)
  }

  http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
    log.Printf("[<] %#v %#v", w, r)
  })

  done, errs := listeners.IntoIter().FilterHTTP().ServeHTTP(
    http.DefaultServerMux,
    servers.Context(context.Background()),
  )

  select {
  case <-done:
  case err := <-errsChan:
    log.Printf("error listen: %#v", errs)
  }
}
```

[godev-image]: https://img.shields.io/badge/go.dev-reference-5272B4?logo=go&logoColor=white
[godev-url]: https://pkg.go.dev/github.com/go-x-pkg/servers

[build-image]: https://travis-ci.com/go-x-pkg/servers.svg?branch=master
[build-url]: https://travis-ci.com/go-x-pkg/servers

[coverage-image]: https://coveralls.io/repos/github/go-x-pkg/servers/badge.svg?branch=master
[coverage-url]: https://coveralls.io/github/go-x-pkg/servers?branch=master

[goreport-image]: https://goreportcard.com/badge/github.com/go-x-pkg/servers
[goreport-url]: https://goreportcard.com/report/github.com/go-x-pkg/servers

[deepsource-io-image]: https://static.deepsource.io/deepsource-badge-light-mini.svg
[deepsource-io-url]: https://deepsource.io/gh/go-x-pkg/servers/?ref=repository-badge
