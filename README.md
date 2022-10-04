connect-grpcreflect-go
======================

[![Build](https://github.com/bufbuild/connect-grpcreflect-go/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/bufbuild/connect-grpcreflect-go/actions/workflows/ci.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/bufbuild/connect-grpcreflect-go)](https://goreportcard.com/report/github.com/bufbuild/connect-grpcreflect-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/bufbuild/connect-grpcreflect-go.svg)](https://pkg.go.dev/github.com/bufbuild/connect-grpcreflect-go)

`connect-grpcreflect-go` adds support for gRPC's server reflection API to any
`net/http` server &mdash; including those built with [Connect][connect-go]. With
server reflection enabled, ad-hoc debugging tools can call your gRPC-compatible
handlers and print the responses *without* a copy of the schema.

The exposed reflection API is wire compatible with Google's gRPC
implementations, so it works with [grpcurl], [grpcui], [BloomRPC], and many
other tools.

For more on Connect, see the [announcement blog post][blog], the documentation
on [connect.build][docs] (especially the [Getting Started] guide for Go), the
[`connect-go`][connect-go] repo, or the [demo service][demo].

## Example

```go
package main

import (
  "net/http"

  "golang.org/x/net/http2"
  "golang.org/x/net/http2/h2c"
  grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
)

func main() {
  mux := http.NewServeMux()
  reflector := grpcreflect.NewStaticReflector(
    "acme.user.v1.UserService",
    "acme.group.v1.GroupService",
    // protoc-gen-connect-go generates package-level constants
    // for these fully-qualified protobuf service names, so you'd more likely
    // reference userv1.UserServiceName and groupv1.GroupServiceName.
  )
  mux.Handle(grpcreflect.NewHandlerV1(reflector))
  // Many tools still expect the older version of the server reflection API, so
  // most servers should mount both handlers.
  mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
  // If you don't need to support HTTP/2 without TLS (h2c), you can drop
  // x/net/http2 and use http.ListenAndServeTLS instead.
  http.ListenAndServe(
    ":8080",
    h2c.NewHandler(mux, &http2.Server{}),
  )
}
```

## Status: Stable

This module is stable. It supports:

* The [two most recent major releases][go-support-policy] of Go.
* [APIv2] of Protocol Buffers in Go (`google.golang.org/protobuf`).

Within those parameters, `connect-grpcreflect-go` follows semantic versioning.
We will _not_ make breaking changes in the 1.x series of releases.

## Legal

Offered under the [Apache 2 license][license].

[APIv2]: https://blog.golang.org/protobuf-apiv2
[BloomRPC]: https://github.com/bloomrpc/bloomrpc
[Getting Started]: https://connect.build/go/getting-started
[blog]: https://buf.build/blog/connect-a-better-grpc
[connect-go]: https://github.com/bufbuild/connect-go
[demo]: https://github.com/bufbuild/connect-demo
[docs]: https://connect.build
[go-support-policy]: https://golang.org/doc/devel/release#policy
[grpcui]: https://github.com/fullstorydev/grpcui
[grpcurl]: https://github.com/fullstorydev/grpcurl
[license]: https://github.com/bufbuild/connect-grpcreflect-go/blob/main/LICENSE.txt
