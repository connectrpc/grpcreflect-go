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
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bufbuild/connect"
	_ "github.com/bufbuild/connect-grpcreflect-go/internal/gen/go/connect/reflecttest/v1"
	reflectionv1alpha1 "github.com/bufbuild/connect-grpcreflect-go/internal/gen/go/connectext/grpc/reflection/v1alpha"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestReflection(t *testing.T) {
	const service = "internal.reflection.v1alpha1.ServerReflection"
	t.Run("static", func(t *testing.T) {
		reflector := NewStaticReflector(service)
		testReflector(t, reflector)
	})
	t.Run("options", func(t *testing.T) {
		reflector := NewReflector(
			&staticNames{names: []string{service}},
			WithExtensionResolver(protoregistry.GlobalTypes),
			WithDescriptorResolver(protoregistry.GlobalFiles),
		)
		testReflector(t, reflector)
	})
}

func testReflector(t *testing.T, reflector *Reflector) {
	mux := http.NewServeMux()
	mux.Handle(NewHandler(reflector))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	reflectionRequestFQN := string((&reflectionv1alpha1.ServerReflectionRequest{}).
		ProtoReflect().
		Descriptor().
		FullName())
	client, err := connect.NewClient[
		reflectionv1alpha1.ServerReflectionRequest,
		reflectionv1alpha1.ServerReflectionResponse,
	](
		server.Client(),
		server.URL+"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		connect.WithGRPC(),
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	call := func(req *reflectionv1alpha1.ServerReflectionRequest) (*reflectionv1alpha1.ServerReflectionResponse, error) {
		res, err := client.CallUnary(context.Background(), connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, err
	}

	assertFileDescriptorResponseContains := func(
		t testing.TB,
		req *reflectionv1alpha1.ServerReflectionRequest,
		substring string,
	) {
		t.Helper()
		res, err := call(req)
		if err != nil {
			t.Fatal(err.Error())
		}
		if res.GetErrorResponse() != nil {
			t.Fatal(res.GetErrorResponse())
		}
		fds := res.GetFileDescriptorResponse()
		if fds == nil {
			t.Fatal("got nil FileDescriptorResponse")
		}
		if len(fds.FileDescriptorProto) != 1 {
			t.Fatalf("got %d FileDescriptorProtos, expected 1", len(fds.FileDescriptorProto))
		}
		if !bytes.Contains(fds.FileDescriptorProto[0], []byte(substring)) {
			t.Fatalf(
				"expected FileDescriptorProto to contain %s, got:\n%v",
				substring,
				fds.FileDescriptorProto[0],
			)
		}
	}

	assertFileDescriptorResponseNotFound := func(
		t testing.TB,
		req *reflectionv1alpha1.ServerReflectionRequest,
	) {
		t.Helper()
		res, netErr := call(req)
		if netErr != nil {
			t.Fatal(err.Error())
		}
		err := res.GetErrorResponse()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.ErrorCode != int32(connect.CodeNotFound) {
			t.Fatalf("got code %v, expected %v", err.ErrorCode, connect.CodeNotFound)
		}
		if err.ErrorMessage == "" {
			t.Fatalf("got empty error message, expected some text")
		}
	}

	t.Run("list_services", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_ListServices{
				ListServices: "ignored per protobuf documentation",
			},
		}
		res, err := call(req)
		if err != nil {
			t.Fatal(err.Error())
		}
		expect := &reflectionv1alpha1.ServerReflectionResponse{
			ValidHost:       req.Host,
			OriginalRequest: req,
			MessageResponse: &reflectionv1alpha1.ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &reflectionv1alpha1.ListServiceResponse{
					Service: []*reflectionv1alpha1.ServiceResponse{
						{Name: "internal.reflection.v1alpha1.ServerReflection"},
					},
				},
			},
		}
		if diff := cmp.Diff(expect, res, protocmp.Transform()); diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("file_by_filename", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_FileByFilename{
				FileByFilename: "connectext/grpc/reflection/v1alpha/reflection.proto",
			},
		}
		assertFileDescriptorResponseContains(t, req, reflectionRequestFQN)
	})
	t.Run("file_by_filename_missing", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_FileByFilename{
				FileByFilename: "foo.proto",
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
	t.Run("file_containing_symbol", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: reflectionRequestFQN,
			},
		}
		assertFileDescriptorResponseContains(t, req, "reflection.proto")
	})
	t.Run("file_containing_symbol_missing", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: "something.Thing",
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
	t.Run("file_containing_extension", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_FileContainingExtension{
				FileContainingExtension: &reflectionv1alpha1.ExtensionRequest{
					ContainingType:  "connect.reflecttest.v1.Extendable",
					ExtensionNumber: 10,
				},
			},
		}
		assertFileDescriptorResponseContains(t, req, "reflecttest_ext.proto")
	})
	t.Run("file_containing_extension_missing", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_FileContainingExtension{
				FileContainingExtension: &reflectionv1alpha1.ExtensionRequest{
					ContainingType:  "connect.reflecttest.v1.Extendable",
					ExtensionNumber: 42,
				},
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
	t.Run("all_extension_numbers_of_type", func(t *testing.T) {
		const extendableFQN = "connect.reflecttest.v1.Extendable"
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_AllExtensionNumbersOfType{
				AllExtensionNumbersOfType: extendableFQN,
			},
		}
		res, err := call(req)
		if err != nil {
			t.Fatal(err.Error())
		}
		expect := &reflectionv1alpha1.ServerReflectionResponse{
			ValidHost:       req.Host,
			OriginalRequest: req,
			MessageResponse: &reflectionv1alpha1.ServerReflectionResponse_AllExtensionNumbersResponse{
				AllExtensionNumbersResponse: &reflectionv1alpha1.ExtensionNumberResponse{
					BaseTypeName:    extendableFQN,
					ExtensionNumber: []int32{10, 11},
				},
			},
		}
		if diff := cmp.Diff(expect, res, protocmp.Transform()); diff != "" {
			t.Fatal(diff)
		}
	})
	t.Run("all_extension_numbers_of_type_missing", func(t *testing.T) {
		req := &reflectionv1alpha1.ServerReflectionRequest{
			Host: "some-host",
			MessageRequest: &reflectionv1alpha1.ServerReflectionRequest_AllExtensionNumbersOfType{
				AllExtensionNumbersOfType: "foobar",
			},
		}
		assertFileDescriptorResponseNotFound(t, req)
	})
}
