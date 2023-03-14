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

package grpcreflect_test

import (
	"context"
	"fmt"
	"github.com/bufbuild/connect-grpcreflect-go"
	"log"
	"net/http"
)

func ExampleNewClient() {
	// Create a client to the Connect demo server.
	client := grpcreflect.NewClient(http.DefaultClient, "https://demo.connect.build")
	// Create a new reflection stream.
	stream := client.NewStream(context.Background())
	// Ask the server for its services and for the file descriptor that contains the first one.
	names, err := stream.ListServices()
	if err != nil {
		log.Printf("error listing services: %v", err)
		return
	}
	fmt.Printf("services: %v\n", names)
	files, err := stream.FileContainingSymbol(names[0])
	if err != nil {
		log.Printf("error getting file that contains %q: %v", names[0], err)
		return
	}
	fmt.Printf("file descriptor for %q\n", files[len(files)-1].GetName())
	// Output:
	// services: [buf.connect.demo.eliza.v1.ElizaService]
	// file descriptor for "buf/connect/demo/eliza/v1/eliza.proto"
}
