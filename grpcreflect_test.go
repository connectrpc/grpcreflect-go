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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect-go"
	_ "github.com/bufbuild/connect-grpcreflect-go/internal/gen/go/connect/reflecttest/v1"
	reflectionv1 "github.com/bufbuild/connect-grpcreflect-go/internal/gen/go/connectext/grpc/reflection/v1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"
)

const actualServiceName = "connectext.grpc.reflection.v1.ServerReflection"

func TestReflection(t *testing.T) {
	t.Parallel()
	t.Run("static", func(t *testing.T) {
		t.Parallel()
		reflector := NewStaticReflector(actualServiceName)
		testReflector(t, reflector, serviceURLPathV1)
	})
	t.Run("v1alpha1", func(t *testing.T) {
		t.Parallel()
		reflector := NewStaticReflector(actualServiceName)
		testReflector(t, reflector, serviceURLPathV1Alpha)
	})
	t.Run("options", func(t *testing.T) {
		t.Parallel()
		reflector := NewReflector(
			&staticNames{names: []string{actualServiceName}},
			WithExtensionResolver(protoregistry.GlobalTypes),
			WithDescriptorResolver(protoregistry.GlobalFiles),
		)
		testReflector(t, reflector, serviceURLPathV1)
	})
}

func testReflector(t *testing.T, reflector *Reflector, servicePath string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(NewHandlerV1(reflector))
	mux.Handle(NewHandlerV1Alpha(reflector))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	reflectionRequestFQN := string((&reflectionv1.ServerReflectionRequest{}).
		ProtoReflect().
		Descriptor().
		FullName())
	client := connect.NewClient[
		reflectionv1.ServerReflectionRequest,
		reflectionv1.ServerReflectionResponse,
	](
		server.Client(),
		server.URL+servicePath+methodName,
		connect.WithGRPC(),
	)
	call := func(req *reflectionv1.ServerReflectionRequest) (*reflectionv1.ServerReflectionResponse, error) {
		res, err := client.CallUnary(context.Background(), connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}

	assertFileDescriptorResponseContains := func(
		tb testing.TB,
		req *reflectionv1.ServerReflectionRequest,
		substring string,
	) {
		tb.Helper()
		res, err := call(req)
		if err != nil {
			tb.Fatal(err.Error())
		}
		if res.GetErrorResponse() != nil {
			tb.Fatal(res.GetErrorResponse())
		}
		fds := res.GetFileDescriptorResponse()
		if fds == nil {
			tb.Fatal("got nil FileDescriptorResponse")
			return // convinces staticcheck that remaining code is unreachable
		}
		if len(fds.FileDescriptorProto) != 1 {
			tb.Fatalf("got %d FileDescriptorProtos, expected 1", len(fds.FileDescriptorProto))
		}
		if !bytes.Contains(fds.FileDescriptorProto[0], []byte(substring)) {
			tb.Fatalf(
				"expected FileDescriptorProto to contain %s, got:\n%v",
				substring,
				fds.FileDescriptorProto[0],
			)
		}
	}

	assertFileDescriptorResponseNotFound := func(
		tb testing.TB,
		req *reflectionv1.ServerReflectionRequest,
	) {
		tb.Helper()
		res, netErr := call(req)
		if netErr != nil {
			tb.Fatal(netErr)
		}
		err := res.GetErrorResponse()
		if err == nil {
			tb.Fatal("expected error, got nil")
			return // convinces staticcheck that remaining code is unreachable
		}
		if err.ErrorCode != int32(connect.CodeNotFound) {
			tb.Fatalf("got code %v, expected %v", err.ErrorCode, connect.CodeNotFound)
		}
		if err.ErrorMessage == "" {
			tb.Fatalf("got empty error message, expected some text")
		}
	}

	t.Run("list_services", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_ListServices{
				ListServices: "ignored per protobuf documentation",
			},
		}
		res, err := call(req)
		if err != nil {
			t.Fatal(err)
		}
		expect := &reflectionv1.ServerReflectionResponse{
			ValidHost:       req.Host,
			OriginalRequest: req,
			MessageResponse: &reflectionv1.ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &reflectionv1.ListServiceResponse{
					Service: []*reflectionv1.ServiceResponse{
						{Name: "connectext.grpc.reflection.v1.ServerReflection"},
					},
				},
			},
		}
		if diff := cmp.Diff(expect, res, protocmp.Transform()); diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("file_by_filename", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileByFilename{
				FileByFilename: "connectext/grpc/reflection/v1/reflection.proto",
			},
		}
		assertFileDescriptorResponseContains(t, req, reflectionRequestFQN)
	})
	t.Run("file_by_filename_missing", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileByFilename{
				FileByFilename: "foo.proto",
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
	t.Run("file_containing_symbol", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: reflectionRequestFQN,
			},
		}
		assertFileDescriptorResponseContains(t, req, "reflection.proto")
	})
	t.Run("file_containing_symbol_missing", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: "something.Thing",
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
	t.Run("file_containing_extension", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingExtension{
				FileContainingExtension: &reflectionv1.ExtensionRequest{
					ContainingType:  "connect.reflecttest.v1.Extendable",
					ExtensionNumber: 10,
				},
			},
		}
		assertFileDescriptorResponseContains(t, req, "reflecttest_ext.proto")
	})
	t.Run("file_containing_extension_missing", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_FileContainingExtension{
				FileContainingExtension: &reflectionv1.ExtensionRequest{
					ContainingType:  "connect.reflecttest.v1.Extendable",
					ExtensionNumber: 42,
				},
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
	t.Run("all_extension_numbers_of_type", func(t *testing.T) {
		t.Parallel()
		const extendableFQN = "connect.reflecttest.v1.Extendable"
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType{
				AllExtensionNumbersOfType: extendableFQN,
			},
		}
		res, err := call(req)
		if err != nil {
			t.Fatal(err.Error())
		}
		expect := &reflectionv1.ServerReflectionResponse{
			ValidHost:       req.Host,
			OriginalRequest: req,
			MessageResponse: &reflectionv1.ServerReflectionResponse_AllExtensionNumbersResponse{
				AllExtensionNumbersResponse: &reflectionv1.ExtensionNumberResponse{
					BaseTypeName:    extendableFQN,
					ExtensionNumber: []int32{10, 11},
				},
			},
		}
		if diff := cmp.Diff(expect, res, protocmp.Transform()); diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("all_extension_numbers_of_type_find_descriptor_by_name", func(t *testing.T) {
		const extendableFQN = "connect.reflecttest.v1.DoRequest"
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType{
				AllExtensionNumbersOfType: extendableFQN,
			},
		}
		res, err := call(req)
		if err != nil {
			t.Fatal(err.Error())
		}
		expect := &reflectionv1.ServerReflectionResponse{
			ValidHost:       req.Host,
			OriginalRequest: req,
			MessageResponse: &reflectionv1.ServerReflectionResponse_AllExtensionNumbersResponse{
				AllExtensionNumbersResponse: &reflectionv1.ExtensionNumberResponse{
					BaseTypeName:    extendableFQN,
					ExtensionNumber: []int32{},
				},
			},
		}
		if diff := cmp.Diff(expect, res, protocmp.Transform()); diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("all_extension_numbers_of_type_missing", func(t *testing.T) {
		t.Parallel()
		req := &reflectionv1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType{
				AllExtensionNumbersOfType: "foobar",
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
}
