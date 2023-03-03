// Copyright 2022 Buf Technologies, Inc.
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
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/bufbuild/connect-go"
	reflectionv1 "github.com/bufbuild/connect-grpcreflect-go/internal/gen/go/connectext/grpc/reflection/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func NewClientV1(httpClient connect.HTTPClient, baseURL string, options ...connect.ClientOption) *Client {
	return newClient(
		httpClient,
		baseURL+serviceNameV1,
		options...,
	)
}

func NewClientV1Alpha(httpClient connect.HTTPClient, baseURL string, options ...connect.ClientOption) *Client {
	return newClient(
		httpClient,
		baseURL+serviceNameV1Alpha,
		options...,
	)
}

func newClient(httpClient connect.HTTPClient, serviceURL string, options ...connect.ClientOption) *Client {
	client := connect.NewClient[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse](
		httpClient,
		serviceURL+methodName,
		options...,
	)
	return &Client{client: client}
}

type reflectClient = connect.Client[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse]
type reflectStream = connect.BidiStreamForClient[reflectionv1.ServerReflectionRequest, reflectionv1.ServerReflectionResponse]

type Client struct {
	client *reflectClient
}

type ClientStreamOption func(*clientStreamOptions)

func WithRequestHeaders(headers http.Header) ClientStreamOption {
	return func(o *clientStreamOptions) {
		o.headers = headers
	}
}

func WithReflectionHost(host string) ClientStreamOption {
	return func(o *clientStreamOptions) {
		o.host = host
	}
}

type clientStreamOptions struct {
	host    string
	headers http.Header
}

func (c *Client) NewReflectionStream(ctx context.Context, options ...ClientStreamOption) *ClientStream {
	var opts clientStreamOptions
	for _, option := range options {
		option(&opts)
	}
	stream := c.client.CallBidiStream(ctx)
	for k, v := range opts.headers {
		stream.RequestHeader()[k] = v
	}
	// we can eagerly send request headers; we can ignore return
	// value because caller will see any errors when calling any
	// other method on returned stream
	_ = stream.Send(nil)
	return &ClientStream{host: opts.host, stream: stream}
}

func IsReflectionStreamError(err error) bool {
	var streamErr *streamError
	return errors.As(err, &streamErr)
}

type ClientStream struct {
	host string

	mu     sync.Mutex
	stream *reflectStream
}

func (cs *ClientStream) Peer() connect.Peer {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.stream.Peer()
}

func (cs *ClientStream) ResponseHeader() http.Header {
	cs.mu.Lock()
	stream := cs.stream
	cs.mu.Unlock()
	// this will block until headers received, so we can't hold lock
	// while calling or else we'll potentially deadlock calls to send
	return stream.ResponseHeader()
}

func (cs *ClientStream) ResponseTrailer() http.Header {
	cs.mu.Lock()
	stream := cs.stream
	cs.mu.Unlock()
	return stream.ResponseTrailer()
}

func (cs *ClientStream) ListServices() ([]protoreflect.FullName, error) {
	resp, err := cs.send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_ListServices{
			ListServices: "",
		},
	})
	if err != nil {
		return nil, err
	}
	respNames := resp.GetListServicesResponse()
	if respNames == nil {
		return nil, errWrongResponseType(resp, "list_services")
	}
	names := make([]protoreflect.FullName, len(respNames.Service))
	for i, svc := range respNames.Service {
		names[i] = protoreflect.FullName(svc.Name)
	}
	return names, nil
}

func errWrongResponseType(resp *reflectionv1.ServerReflectionResponse, operation string) error {
	return fmt.Errorf("protocol error: wrong response type %T in reply to %s", resp.MessageResponse, operation)
}

func (cs *ClientStream) FileByFilename(filename string) ([]*descriptorpb.FileDescriptorProto, error) {
	return cs.getDescriptors("file_by_filename", &reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileByFilename{
			FileByFilename: filename,
		},
	})
}

func (cs *ClientStream) FileContainingSymbol(name protoreflect.FullName) ([]*descriptorpb.FileDescriptorProto, error) {
	return cs.getDescriptors("file_containing_symbol", &reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: string(name),
		},
	})
}

func (cs *ClientStream) FileContainingExtension(messageName protoreflect.FullName, extensionNumber protoreflect.FieldNumber) ([]*descriptorpb.FileDescriptorProto, error) {
	return cs.getDescriptors("file_containing_extension", &reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingExtension{
			FileContainingExtension: &reflectionv1.ExtensionRequest{
				ContainingType:  string(messageName),
				ExtensionNumber: int32(extensionNumber),
			},
		},
	})
}

func (cs *ClientStream) AllExtensionNumbers(messageName protoreflect.FullName) ([]protoreflect.FieldNumber, error) {
	resp, err := cs.send(&reflectionv1.ServerReflectionRequest{
		MessageRequest: &reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType{
			AllExtensionNumbersOfType: string(messageName),
		},
	})
	if err != nil {
		return nil, err
	}
	respExtNumbers := resp.GetAllExtensionNumbersResponse()
	if respExtNumbers == nil {
		return nil, errWrongResponseType(resp, "all_extension_numbers")
	}
	extNumbers := make([]protoreflect.FieldNumber, len(respExtNumbers.ExtensionNumber))
	for i, num := range respExtNumbers.ExtensionNumber {
		extNumbers[i] = protoreflect.FieldNumber(num)
	}
	return extNumbers, nil
}

func (cs *ClientStream) getDescriptors(operation string, req *reflectionv1.ServerReflectionRequest) ([]*descriptorpb.FileDescriptorProto, error) {
	resp, err := cs.send(req)
	if err != nil {
		return nil, err
	}
	respDescriptors := resp.GetFileDescriptorResponse()
	if respDescriptors == nil {
		return nil, errWrongResponseType(resp, operation)
	}
	descriptors := make([]*descriptorpb.FileDescriptorProto, len(respDescriptors.FileDescriptorProto))
	for i, data := range respDescriptors.FileDescriptorProto {
		fileDescriptor := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(data, fileDescriptor); err != nil {
			return nil, fmt.Errorf("reply to %s contained invalid descriptor proto: %w", operation, err)
		}
		descriptors[i] = fileDescriptor
	}
	return descriptors, nil
}

func (cs *ClientStream) send(req *reflectionv1.ServerReflectionRequest) (*reflectionv1.ServerReflectionResponse, error) {
	req.Host = cs.host
	// Sending on a bidi stream is usually thread-safe. But the replies are in the same order
	// as the requests. So to prevent concurrent use from interleaving replies (which would
	// require much more logic here to properly correlate replies with requests), we send and
	// receive while holding the mutex. This means that this API does not support pipelining
	// reflection requests (which could in theory reduce latency, but only when the client
	// knows all of their requests up-front, which is rarely the case since subsequent calls
	// often depend on the data in prior responses.
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if err := cs.stream.Send(req); err != nil {
		return nil, &streamError{err: err}
	}
	resp, err := cs.stream.Receive()
	if err != nil {
		return nil, &streamError{err: err}
	}
	if errResp := resp.GetErrorResponse(); errResp != nil {
		return nil, connect.NewError(connect.Code(errResp.ErrorCode), errors.New(errResp.ErrorMessage))
	}
	return resp, nil
}

func (cs *ClientStream) Close() error {
	cs.mu.Lock()
	stream := cs.stream
	cs.mu.Unlock()

	if err := stream.CloseRequest(); err != nil {
		return err
	}
	return stream.CloseResponse()
}

type streamError struct {
	err error
}

func (e *streamError) Error() string {
	return e.err.Error()
}

func (e *streamError) Unwrap() error {
	return e.err
}
