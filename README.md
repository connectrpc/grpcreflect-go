connect-grpcreflect-go
======================

[![Build](https://github.com/bufbuild/connect-grpcreflect-go/actions/workflows/ci.yaml/badge.svg?event=push?branch=main)](https://github.com/bufbuild/connect-grpcreflect-go/actions/workflows/ci.yaml)
[![Report Card](https://goreportcard.com/badge/github.com/bufbuild/connect-grpcreflect-go)](https://goreportcard.com/report/github.com/bufbuild/connect-grpcreflect-go)
[![GoDoc](https://pkg.go.dev/badge/github.com/bufbuild/connect-grpcreflect-go.svg)](https://pkg.go.dev/github.com/bufbuild/connect-grpcreflect-go)

`connect-grpcreflect-go` adds support for gRPC's server reflection API to any
`net/http` server&mdash;including those built with [Connect][docs]! The server
reflection API lets ad-hoc debugging tools call your Protobuf services and
print the responses *without* a copy of the schema.

The exposed reflection API is wire compatible with Google's gRPC
implementations, so it works with [grpcurl], [grpcui], [BloomRPC], and many
other tools.

## Example

```go
package main

import (
  "net/http"

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
  mux.Handle(grpcreflect.NewHandler(reflector))
  http.ListenAndServeTLS(":8081", "server.crt", "server.key", mux)
}
```

## Status

Like [Connect][] itself, `connect-grpcreflect-go` is in _beta_. We plan to tag a
release candidate in July 2022 and stable v1 soon after the Go 1.19 release.

## Support and Versioning

`connect-grpcreflect-go` supports:

* The [two most recent major releases][go-support-policy] of Go, with a minimum
  of Go 1.18.
* [APIv2][] of protocol buffers in Go (`google.golang.org/protobuf`).

Within those parameters, it follows semantic versioning.

## Legal

Offered under the [Apache 2 license][license].

[APIv2]: https://blog.golang.org/protobuf-apiv2
[BloomRPC]: https://github.com/bloomrpc/bloomrpc
[connect]: https://github.com/bufbuild/connect
[docs]: https://bufconnect.com
[go-support-policy]: https://golang.org/doc/devel/release#policy
[grpcui]: https://github.com/fullstorydev/grpcui
[grpcurl]: https://github.com/fullstorydev/grpcurl
[license]: https://github.com/bufbuild/connect-grpcreflect-go/blob/main/LICENSE.txt
