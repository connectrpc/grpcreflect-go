// Copyright 2022-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package grpcreflect enables any net/http server, including those built with
// Connect, to handle gRPC's server reflection API. This lets ad-hoc debugging
// tools call your Protobuf services and print the responses without a copy of
// the schema.
//
// The exposed reflection API is wire compatible with Google's gRPC
// implementations, so it works with grpcurl, grpcui, BloomRPC, and many other
// tools.
//
// The core Connect package is github.com/bufbuild/connect-go. Documentation is
// available at https://connect.build.
package grpcreflect

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"google.golang.org/protobuf/reflect/protodesc"
)

const (
	// ReflectV1ServiceName is the fully-qualified name of the v1 version of the reflection service.
	ReflectV1ServiceName = grpcreflect.ReflectV1ServiceName
	// ReflectV1AlphaServiceName is the fully-qualified name of the v1alpha version of the reflection service.
	ReflectV1AlphaServiceName = grpcreflect.ReflectV1AlphaServiceName
)

// NewHandlerV1 constructs an implementation of v1 of the gRPC server reflection
// API. It returns an HTTP handler and the path on which to mount it.
//
// Note that because the reflection API requires bidirectional streaming, the
// returned handler doesn't support HTTP/1.1. If your server must also support
// older tools that use the v1alpha server reflection API, see NewHandlerV1Alpha.
func NewHandlerV1(reflector *Reflector, options ...connect.HandlerOption) (string, http.Handler) {
	return grpcreflect.NewHandlerV1(reflector, options...)
}

// NewHandlerV1Alpha constructs an implementation of v1alpha of the gRPC server
// reflection API, which is useful to support tools that haven't updated to the
// v1 API. It returns an HTTP handler and the path on which to mount it.
//
// Like NewHandlerV1, the returned handler doesn't support HTTP/1.1.
func NewHandlerV1Alpha(reflector *Reflector, options ...connect.HandlerOption) (string, http.Handler) {
	return grpcreflect.NewHandlerV1Alpha(reflector, options...)
}

// Reflector implements the underlying logic for gRPC's protobuf server
// reflection. They're configurable, so they can support straightforward
// process-local reflection or more complex proxying.
//
// Keep in mind that by default, Reflectors expose every protobuf type and
// extension compiled into your binary. Think twice before including the
// default Reflector in a public API.
//
// For more information, see
// https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md,
// https://github.com/grpc/grpc/blob/master/doc/server-reflection.md, and
// https://github.com/fullstorydev/grpcurl.
type Reflector = grpcreflect.Reflector

// NewReflector constructs a highly configurable Reflector: it can serve a
// dynamic list of services, proxy reflection requests to other backends, or
// use an alternate source of extension information.
//
// To build a simpler Reflector that supports a static list of services using
// information from the package-global Protobuf registry, use
// NewStaticReflector.
func NewReflector(namer Namer, options ...Option) *Reflector {
	return grpcreflect.NewReflector(namer, options...)
}

// NewStaticReflector constructs a simple Reflector that supports a static list
// of services using information from the package-global Protobuf registry. For
// a more configurable Reflector, use NewReflector.
//
// The supplied strings should be fully-qualified Protobuf service names (for
// example, "acme.user.v1.UserService"). Generated Connect service files
// have this declared as a constant.
func NewStaticReflector(services ...string) *Reflector {
	return grpcreflect.NewStaticReflector(services...)
}

// A Namer lists the fully-qualified Protobuf service names available for
// reflection (for example, "acme.user.v1.UserService"). Namers must be safe to
// call concurrently.
type Namer = grpcreflect.Namer

// An Option configures a Reflector.
type Option = grpcreflect.Option

// WithExtensionResolver sets the resolver used to find Protobuf extensions. By
// default, Reflectors use protoregistry.GlobalTypes.
func WithExtensionResolver(resolver ExtensionResolver) Option {
	return grpcreflect.WithExtensionResolver(resolver)
}

// WithDescriptorResolver sets the resolver used to find Protobuf type
// information (typically called a "descriptor"). By default, Reflectors use
// protoregistry.GlobalFiles.
func WithDescriptorResolver(resolver protodesc.Resolver) Option {
	return grpcreflect.WithDescriptorResolver(resolver)
}

// An ExtensionResolver lets server reflection implementations query details
// about the registered Protobuf extensions. protoregistry.GlobalTypes
// implements ExtensionResolver.
//
// ExtensionResolvers must be safe to call concurrently.
type ExtensionResolver = grpcreflect.ExtensionResolver
