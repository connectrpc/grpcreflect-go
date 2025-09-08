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
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"connectrpc.com/connect"
	_ "connectrpc.com/grpcreflect/internal/gen/go/connect/reflecttest/v1"
	reflectionv1 "connectrpc.com/grpcreflect/internal/gen/go/connectext/grpc/reflection/v1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/descriptorpb"
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
		res, err := client.CallUnary(t.Context(), connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}

	assertFileDescriptorResponseContains := func(
		tb testing.TB,
		req *reflectionv1.ServerReflectionRequest,
		numFiles int,
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
		if len(fds.FileDescriptorProto) != numFiles {
			tb.Fatalf("got %d FileDescriptorProtos, expected %d", len(fds.FileDescriptorProto), numFiles)
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
		assertFileDescriptorResponseContains(t, req, 1, reflectionRequestFQN)
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
		assertFileDescriptorResponseContains(t, req, 1, "reflection.proto")
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
		// We expect two files here: both reflecttest_ext.proto and its dependency, reflecttest.proto
		assertFileDescriptorResponseContains(t, req, 2, "reflecttest_ext.proto")
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

func TestFileDescriptorWithDependencies(t *testing.T) {
	t.Parallel()

	depFile, err := protodesc.NewFile(
		&descriptorpb.FileDescriptorProto{
			Name: proto.String("dep.proto"),
		}, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	deps := &protoregistry.Files{}
	if err := deps.RegisterFile(depFile); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	rootFileProto := &descriptorpb.FileDescriptorProto{
		Name: proto.String("root.proto"),
		Dependency: []string{
			"google/protobuf/descriptor.proto",
			"connect/reflecttest/v1/reflecttest_ext.proto",
			"dep.proto",
		},
	}

	// dep.proto is in deps; the other imports come from protoregistry.GlobalFiles
	resolver := &combinedResolver{first: protoregistry.GlobalFiles, second: deps}
	rootFile, err := protodesc.NewFile(rootFileProto, resolver)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Create a file hierarchy that contains a placeholder for dep.proto
	placeholderDep := placeholderFile{depFile}
	placeholderDeps := &protoregistry.Files{}
	if err := placeholderDeps.RegisterFile(placeholderDep); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	resolver = &combinedResolver{first: protoregistry.GlobalFiles, second: placeholderDeps}

	rootFileHasPlaceholderDep, err := protodesc.NewFile(rootFileProto, resolver)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	rootFileIsPlaceholder := placeholderFile{rootFile}

	// Full transitive dependency graph of root.proto includes five files:
	// - root.proto
	//   - google/protobuf/descriptor.proto
	//   - connect/reflecttest/v1/reflecttest_ext.proto
	//     - connect/reflecttest/v1/reflecttest.proto
	//   - dep.proto

	testCases := []struct {
		name   string
		sent   []string
		root   protoreflect.FileDescriptor
		expect []string
	}{
		{
			name: "send_all",
			root: rootFile,
			// expect full transitive closure
			expect: []string{
				"root.proto",
				"google/protobuf/descriptor.proto",
				"connect/reflecttest/v1/reflecttest_ext.proto",
				"connect/reflecttest/v1/reflecttest.proto",
				"dep.proto",
			},
		},
		{
			name: "already_sent",
			sent: []string{
				"root.proto",
				"google/protobuf/descriptor.proto",
				"connect/reflecttest/v1/reflecttest_ext.proto",
				"connect/reflecttest/v1/reflecttest.proto",
				"dep.proto",
			},
			root: rootFile,
			// expect only the root to be re-sent
			expect: []string{"root.proto"},
		},
		{
			name: "some_already_sent",
			sent: []string{
				"connect/reflecttest/v1/reflecttest_ext.proto",
				"connect/reflecttest/v1/reflecttest.proto",
			},
			root: rootFile,
			expect: []string{
				"root.proto",
				"google/protobuf/descriptor.proto",
				"dep.proto",
			},
		},
		{
			name: "root_is_placeholder",
			root: rootFileIsPlaceholder,
			// expect error, no files
		},
		{
			name: "placeholder_skipped",
			root: rootFileHasPlaceholderDep,
			// dep.proto is a placeholder so is skipped
			expect: []string{
				"root.proto",
				"google/protobuf/descriptor.proto",
				"connect/reflecttest/v1/reflecttest_ext.proto",
				"connect/reflecttest/v1/reflecttest.proto",
			},
		},
		{
			name: "placeholder_skipped_and_some_sent",
			sent: []string{
				"connect/reflecttest/v1/reflecttest_ext.proto",
				"connect/reflecttest/v1/reflecttest.proto",
			},
			root: rootFileHasPlaceholderDep,
			expect: []string{
				"root.proto",
				"google/protobuf/descriptor.proto",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sent := &fileDescriptorNameSet{}
			for _, path := range testCase.sent {
				sent.Insert(dummyFile{path: path})
			}

			descriptors, err := fileDescriptorWithDependencies(testCase.root, sent)
			if len(testCase.expect) == 0 {
				// if we're not expecting any files then we're expecting an error
				if err == nil {
					t.Fatalf("expecting an error; instead got %d files", len(descriptors))
				}
				return
			}

			checkDescriptorResults(t, descriptors, testCase.expect)
		})
	}
}

func checkDescriptorResults(t *testing.T, descriptors [][]byte, expect []string) {
	t.Helper()
	if len(descriptors) != len(expect) {
		t.Errorf("expected result to contain %d descriptor(s); instead got %d", len(expect), len(descriptors))
	}
	names := map[string]struct{}{}
	for i, desc := range descriptors {
		var descProto descriptorpb.FileDescriptorProto
		if err := proto.Unmarshal(desc, &descProto); err != nil {
			t.Fatalf("could not unmarshal descriptor result #%d", i+1)
		}
		names[descProto.GetName()] = struct{}{}
	}
	actual := make([]string, 0, len(names))
	for name := range names {
		actual = append(actual, name)
	}
	sort.Strings(actual)
	sort.Strings(expect)
	if !reflect.DeepEqual(actual, expect) {
		t.Fatalf("expected file descriptors for %v; instead got %v", expect, actual)
	}
}

type placeholderFile struct {
	protoreflect.FileDescriptor
}

func (placeholderFile) IsPlaceholder() bool {
	return true
}

type dummyFile struct {
	protoreflect.FileDescriptor
	path string
}

func (f dummyFile) Path() string {
	return f.path
}
