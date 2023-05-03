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
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/bufbuild/connect-go"
	_ "github.com/bufbuild/connect-grpcreflect-go/internal/gen/go/connect/reflecttest/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestClient(t *testing.T) {
	t.Parallel()
	t.Run("v1", func(t *testing.T) {
		t.Parallel()
		testClient(t, func(mux *http.ServeMux) {
			mux.Handle(NewHandlerV1(NewStaticReflector(actualServiceName)))
		})
	})
	t.Run("v1alpha", func(t *testing.T) {
		t.Parallel()
		testClient(t, func(mux *http.ServeMux) {
			mux.Handle(NewHandlerV1Alpha(NewStaticReflector(actualServiceName)))
		})
	})
}

func testClient(t *testing.T, register func(server *http.ServeMux)) {
	t.Helper()
	mux := http.NewServeMux()
	register(mux)
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	ctx := context.Background()
	client := NewClient(server.Client(), server.URL, connect.WithGRPC())
	stream := client.NewStream(ctx)
	t.Cleanup(func() {
		trailers, err := stream.Close()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		// We used gRPC, which always sends back trailers
		if len(trailers) == 0 {
			t.Fatal("expected trailers, got none")
		}
	})

	expectConnectNotFoundError := func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Fatal("expected error but got none")
		}
		if IsReflectionStreamBroken(err) {
			t.Fatalf("error should not be a stream error but is: %v", err)
		}
		var connectError *connect.Error
		if !errors.As(err, &connectError) {
			t.Fatalf("error should be a connect error but is not: %v", err)
		}
		if connectError.Code() != connect.CodeNotFound {
			t.Fatalf("unexpected code: want %v , got %v", connect.CodeNotFound, connectError.Code())
		}
	}

	expectFileDescriptorsContaining := func(t *testing.T, files []*descriptorpb.FileDescriptorProto, err error, path string) {
		t.Helper()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		found := false
		fileNames := make([]string, len(files))
		for i, file := range files {
			fileNames[i] = file.GetName()
			if file.GetName() == path {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected response to include descriptor for %q, but it did not: %v", path, fileNames)
		}
	}

	t.Run("list_services", func(t *testing.T) {
		t.Parallel()
		serviceNames, err := stream.ListServices()
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		expected := []protoreflect.FullName{actualServiceName}
		if !reflect.DeepEqual(expected, serviceNames) {
			t.Fatalf("unexpected service names: want %v ; got %v", expected, serviceNames)
		}
	})

	t.Run("file_by_filename", func(t *testing.T) {
		t.Parallel()
		files, err := stream.FileByFilename("connectext/grpc/reflection/v1/reflection.proto")
		expectFileDescriptorsContaining(t, files, err, "connectext/grpc/reflection/v1/reflection.proto")
	})

	t.Run("file_by_filename_missing", func(t *testing.T) {
		t.Parallel()
		_, err := stream.FileByFilename("foo/bar/baz.proto")
		expectConnectNotFoundError(t, err)
	})

	t.Run("file_containing_symbol", func(t *testing.T) {
		t.Parallel()
		files, err := stream.FileContainingSymbol(actualServiceName)
		expectFileDescriptorsContaining(t, files, err, "connectext/grpc/reflection/v1/reflection.proto")
	})

	t.Run("file_containing_symbol_missing", func(t *testing.T) {
		t.Parallel()
		_, err := stream.FileContainingSymbol("foo.bar.baz.Bedazzle")
		expectConnectNotFoundError(t, err)
	})

	t.Run("file_containing_extension", func(t *testing.T) {
		t.Parallel()
		files, err := stream.FileContainingExtension("connect.reflecttest.v1.Extendable", 10)
		expectFileDescriptorsContaining(t, files, err, "connect/reflecttest/v1/reflecttest_ext.proto")
	})

	t.Run("file_containing_extension_missing", func(t *testing.T) {
		t.Parallel()
		_, err := stream.FileContainingExtension("foo.bar.baz.Bedazzle", 12345)
		expectConnectNotFoundError(t, err)
	})

	t.Run("all_extensions_for_message", func(t *testing.T) {
		t.Parallel()
		exts, err := stream.AllExtensionNumbers("connect.reflecttest.v1.Extendable")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		expected := []protoreflect.FieldNumber{10, 11}
		if !reflect.DeepEqual(expected, exts) {
			t.Fatalf("unexpected extension numbers: want %v ; got %v", expected, exts)
		}
	})

	t.Run("all_extensions_for_message_missing", func(t *testing.T) {
		t.Parallel()
		_, err := stream.AllExtensionNumbers("foo.bar.baz.Bedazzle")
		expectConnectNotFoundError(t, err)
	})
}
