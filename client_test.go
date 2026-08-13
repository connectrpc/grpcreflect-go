// Copyright 2022-2025 The Connect Authors
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
	"sync/atomic"
	"testing"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	_ "connectrpc.com/grpcreflect/v2/internal/gen/go/connect/reflecttest/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestClient(t *testing.T) {
	t.Parallel()
	t.Run("v1", func(t *testing.T) {
		t.Parallel()
		testClient(t, func(mux *http.ServeMux) {
			mountReflectorPath(mux, serviceURLPathV1)
		})
	})
	t.Run("v1alpha", func(t *testing.T) {
		t.Parallel()
		testClient(t, func(mux *http.ServeMux) {
			mountReflectorPath(mux, serviceURLPathV1Alpha)
		})
	})
}

func TestClientCallInfo(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mountReflectorPath(mux, serviceURLPathV1)
	var gotRequestHeader atomic.Value
	gotRequestHeader.Store("")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestHeader.Store(r.Header.Get("Test-Request-Header"))
		mux.ServeHTTP(w, r)
	})
	server := httptest.NewUnstartedServer(handler)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	client := NewClient(connect.NewClient(connecthttp.NewTransport(
		server.Client(), server.URL,
	)))

	ctx, info := connect.NewClientContext(t.Context())
	info.RequestHeader().Set("Test-Request-Header", "request-value")
	stream := client.NewStream(ctx)

	if _, err := stream.ListServices(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got, _ := gotRequestHeader.Load().(string); got != "request-value" {
		t.Errorf("request header: got %q, want %q", got, "request-value")
	}
	if got := info.ResponseHeader().Get("Content-Type"); got == "" {
		t.Error("expected response headers, got none")
	}
	if info.Protocol == "" {
		t.Error("expected Protocol to be set")
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

// mountReflectorPath serves reflection on a single service path, unlike
// Register, so tests can exercise the client's v1 to v1alpha fallback.
func mountReflectorPath(mux *http.ServeMux, servicePath string) {
	ref := &reflector{
		namer:              NamerFunc(func() []string { return []string{actualServiceName} }),
		extensionResolver:  protoregistry.GlobalTypes,
		descriptorResolver: globalFiles,
	}
	server := connect.NewServer()
	server.Register(connect.Method{
		Spec: connect.Spec{
			StreamType: connect.StreamTypeBidi,
			Procedure:  servicePath + methodName,
		},
		Handler: ref.serverReflectionInfo,
	})
	connecthttp.Mount(mux, server)
}

func testClient(t *testing.T, register func(server *http.ServeMux)) {
	t.Helper()
	mux := http.NewServeMux()
	register(mux)
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	t.Cleanup(server.Close)

	// We want to clean up ourselves below, to test stream.Close().
	// So we don't want the context and thus the stream) to have
	// already been canceled. So we don't use t.Context().
	ctx := context.Background()
	client := NewClient(connect.NewClient(connecthttp.NewTransport(
		server.Client(),
		server.URL,
		connecthttp.WithGRPC(),
	)))
	stream := client.NewStream(ctx)
	t.Cleanup(func() {
		if err := stream.Close(); err != nil {
			t.Fatalf("unexpected err: %v", err)
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
