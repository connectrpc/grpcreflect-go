grpcreflect
===========

[![Build](https://github.com/connectrpc/grpcreflect-go/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/connectrpc/grpcreflect-go/actions/workflows/ci.yaml)
[![GoDoc](https://pkg.go.dev/badge/connectrpc.com/grpcreflect/v2.svg)](https://pkg.go.dev/connectrpc.com/grpcreflect/v2)

`connectrpc.com/grpcreflect/v2` adds support for gRPC's server reflection API
to servers built with [Connect][connect]. With server reflection enabled,
ad-hoc debugging tools can call your gRPC-compatible handlers and print the
responses *without* a copy of the schema.

The exposed reflection API is wire compatible with Google's gRPC
implementations, so it works with [grpcurl], [grpcui], and many
other tools.

For more on Connect, see the [announcement blog post][blog], the documentation
on [connectrpc.com][docs] (especially the [Getting Started] guide for Go), the
[Connect][connect] repo, or the [demo service][examples-go].

## Example

```go
package main

import (
  "log"
  "net/http"

  "connectrpc.com/connect/v2"
  "connectrpc.com/connect/v2/connecthttp"
  "connectrpc.com/grpcreflect/v2"
)

func main() {
  server := connect.NewServer()
  // Register your Connect services on the server, then register reflection.
  // By default, reflection describes the services registered on the server,
  // and serves both the v1 and v1alpha versions of the reflection API.
  grpcreflect.Register(server)
  mux := http.NewServeMux()
  connecthttp.Mount(mux, server)
  p := new(http.Protocols)
  p.SetHTTP1(true)
  // The reflection API requires bidirectional streaming, so it's only served
  // over HTTP/2. Supporting HTTP/2 without TLS is convenient for gRPC tools.
  p.SetUnencryptedHTTP2(true)
  s := &http.Server{
    Addr:      ":8080",
    Handler:   mux,
    Protocols: p,
  }
  if err := s.ListenAndServe(); err != nil {
    log.Fatalf("listen failed: %v", err)
  }
}
```

## Status: Unstable

This module is unstable while connect-go v2 is in alpha. Expect breaking
changes as we iterate toward a stable v2 release.

It supports:

* The two most recent major releases of Go. Keep in mind that [only the last
  two releases receive security patches][go-support-policy].
* [APIv2] of Protocol Buffers in Go (`google.golang.org/protobuf`).

## Legal

Offered under the [Apache 2 license][license].

[APIv2]: https://blog.golang.org/protobuf-apiv2
[Getting Started]: https://connectrpc.com/go/getting-started
[blog]: https://buf.build/blog/connect-a-better-grpc
[connect]: https://github.com/connectrpc/connect-go
[examples-go]: https://github.com/connectrpc/examples-go
[docs]: https://connectrpc.com
[go-support-policy]: https://golang.org/doc/devel/release#policy
[grpcui]: https://github.com/fullstorydev/grpcui
[grpcurl]: https://github.com/fullstorydev/grpcurl
[license]: https://github.com/connectrpc/grpcreflect-go/blob/main/LICENSE.txt
