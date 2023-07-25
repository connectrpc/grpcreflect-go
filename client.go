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

package grpcreflect

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
)

// Client is a Connect client for the server reflection service.
type Client = grpcreflect.Client

// NewClient returns a client for interacting with the gRPC server reflection service.
// The given HTTP client, base URL, and options are used to connect to the service.
//
// This client will try "v1" of the service first (grpc.reflection.v1.ServerReflection).
// If this results in a "Not Implemented" error, the client will fall back to "v1alpha"
// of the service (grpc.reflection.v1alpha.ServerReflection).
func NewClient(httpClient connect.HTTPClient, baseURL string, options ...connect.ClientOption) *Client {
	return grpcreflect.NewClient(httpClient, baseURL, options...)
}

// ClientStreamOption is an option that can be provided when calling [Client.NewStream].
type ClientStreamOption = grpcreflect.ClientStreamOption

// WithRequestHeaders is an option that allows the caller to provide the request headers
// that will be sent when a stream is created.
func WithRequestHeaders(headers http.Header) ClientStreamOption {
	return grpcreflect.WithRequestHeaders(headers)
}

// WithReflectionHost is an option that allows the caller to provide the hostname that
// will be included with all requests on the stream. This may be used by the server
// when deciding what source of reflection information to use (which could include
// forwarding the request message to a different host).
func WithReflectionHost(host string) ClientStreamOption {
	return grpcreflect.WithReflectionHost(host)
}

// ClientStream represents a bidirectional stream for downloading reflection information.
// The reflection protocol resembles a sequence of unary RPCs: multiple requests sent on the
// stream, each getting back a corresponding response. However, all such requests and responses
// and sent on a single stream to a single server, to ensure consistency in the information
// downloaded (since different servers could potentially have different versions of reflection
// information).
type ClientStream = grpcreflect.ClientStream

// IsReflectionStreamBroken returns true if the given error was the result of a [ClientStream]
// failing. If the stream returns an error for which this function returns false, only the
// one operation failed; the stream is still intact and may be used for subsequent operations.
func IsReflectionStreamBroken(err error) bool {
	return grpcreflect.IsReflectionStreamBroken(err)
}
