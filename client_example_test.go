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
