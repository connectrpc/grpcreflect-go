// Copyright 2022-2024 The Connect Authors
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

package resolvertest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	grpchealth "connectrpc.com/grpchealth"
	. "connectrpc.com/grpcreflect"
	_ "connectrpc.com/grpcreflect/internal/gen/go/connect/reflecttest/v1"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestResolverHack(t *testing.T) {
	t.Parallel()
	advertisedNames := []string{grpchealth.HealthV1ServiceName, ReflectV1ServiceName, ReflectV1AlphaServiceName}
	reflector := NewStaticReflector(advertisedNames...)
	mux := http.NewServeMux()
	mux.Handle(NewHandlerV1(reflector))
	mux.Handle(NewHandlerV1Alpha(reflector))
	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client := NewClient(server.Client(), server.URL)
	stream := client.NewStream(context.Background())
	defer func() {
		_, _ = stream.Close()
	}()
	names, err := stream.ListServices()
	if err != nil {
		t.Fatal(err.Error())
	}

	namesAsStrings := make([]string, len(names))
	for i := range names {
		namesAsStrings[i] = string(names[i])
	}
	if diff := cmp.Diff(advertisedNames, namesAsStrings); diff != "" {
		t.Fatal(diff)
	}

	for _, serviceName := range names {
		files, err := stream.FileContainingSymbol(serviceName)
		if err != nil {
			t.Errorf("could not get file containing %q: %v", serviceName, err)
			continue
		}
		serviceSimpleName := string(serviceName.Name())
		var fileWithService *descriptorpb.FileDescriptorProto
		for _, file := range files {
			for _, svc := range file.Service {
				if svc.GetName() == serviceSimpleName {
					fileWithService = file
					break
				}
			}
		}
		if fileWithService == nil {
			t.Errorf("returned files did not contain service %q", serviceName)
			continue
		}
		if fileWithService.GetPackage() != string(serviceName.Parent()) {
			t.Errorf("file had unexpected package: want %q, got %q", serviceName.Parent(), fileWithService.GetPackage())
			continue
		}
		if strings.HasPrefix(fileWithService.GetName(), "connectext") {
			t.Errorf("file path contains unwanted 'connectext' prefix: %q", fileWithService.GetName())
			continue
		}
	}

	// Follow-up with a request that comes from the global registry instead
	// of from the hacked overrides
	file, err := stream.FileContainingSymbol("connect.reflecttest.v1.TestService")
	if err != nil {
		t.Fatalf("could not get file containing 'connect.reflecttest.v1.TestService': %v", err)
	}
	if file[0].Service[0].GetName() != "TestService" {
		t.Fatalf("file did not contain service 'connect.reflecttest.v1.TestService'")
	}
}
