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

// Package grpcreflect enables Connect servers to handle gRPC's server
// reflection API. This lets ad-hoc debugging tools call your Protobuf
// services and print the responses without a copy of the schema.
//
// The exposed reflection API is wire compatible with Google's gRPC
// implementations, so it works with grpcurl, grpcui, BloomRPC, and many other
// tools.
//
// The core Connect package is connectrpc.com/connect/v2. Documentation is
// available at https://connectrpc.com.
package grpcreflect

import (
	"context"
	_ "embed" // required for go:embed directive
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"

	"connectrpc.com/connect/v2"
	reflectionv1 "connectrpc.com/grpcreflect/v2/internal/gen/go/connectext/grpc/reflection/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	// ReflectV1ServiceName is the fully-qualified name of the v1 version of the reflection service.
	ReflectV1ServiceName = "grpc.reflection.v1.ServerReflection"
	// ReflectV1AlphaServiceName is the fully-qualified name of the v1alpha version of the reflection service.
	ReflectV1AlphaServiceName = "grpc.reflection.v1alpha.ServerReflection"

	serviceURLPathV1      = "/" + ReflectV1ServiceName + "/"
	serviceURLPathV1Alpha = "/" + ReflectV1AlphaServiceName + "/"
	methodName            = "ServerReflectionInfo"
)

//nolint:gochecknoglobals
var (
	//go:embed services.bin
	embeddedDescriptors []byte

	globalFiles = resolverHackForConnectext(embeddedDescriptors)
)

// Register registers the gRPC server reflection API on server, serving both
// the v1 and v1alpha versions of the service. The v1alpha version supports
// tools that haven't updated to the v1 API.
//
// By default, reflection describes the services registered on server,
// including the reflection services themselves. Use WithNamer to expose a
// different set of services, for example when proxying reflection requests
// to other backends.
//
// Keep in mind that by default, reflection serves every protobuf type and
// extension compiled into your binary. Think twice before registering it on
// a public API.
//
// Note that because the reflection API requires bidirectional streaming,
// HTTP transports serve it only over HTTP/2.
//
// For more information, see
// https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md,
// https://github.com/grpc/grpc/blob/master/doc/server-reflection.md, and
// https://github.com/fullstorydev/grpcurl.
func Register(server *connect.Server, options ...Option) {
	reflector := &reflector{
		namer:              &serverNamer{server: server},
		extensionResolver:  protoregistry.GlobalTypes,
		descriptorResolver: globalFiles,
	}
	for _, option := range options {
		option.apply(reflector)
	}
	// v1 is binary-compatible with v1alpha, so we only need to change paths.
	for _, servicePath := range []string{serviceURLPathV1, serviceURLPathV1Alpha} {
		server.Register(connect.Method{
			Spec: connect.Spec{
				StreamType: connect.StreamTypeBidi,
				Procedure:  servicePath + methodName,
			},
			Handler: reflector.serverReflectionInfo,
		})
	}
}

// reflector implements the underlying logic for gRPC's protobuf server
// reflection.
type reflector struct {
	namer              Namer
	extensionResolver  ExtensionResolver
	descriptorResolver protodesc.Resolver
}

// serverReflectionInfo implements the gRPC server reflection API.
func (r *reflector) serverReflectionInfo(
	_ context.Context,
	_ connect.Spec,
	stream connect.ServerStream,
) error {
	fileDescriptorsSent := &fileDescriptorNameSet{}
	for {
		var request reflectionv1.ServerReflectionRequest
		if err := stream.Receive(&request); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		// The server reflection API sends file descriptors as uncompressed
		// Protobuf-serialized bytes.
		response := &reflectionv1.ServerReflectionResponse{
			ValidHost:       request.Host,
			OriginalRequest: &request,
		}
		switch messageRequest := request.MessageRequest.(type) {
		case *reflectionv1.ServerReflectionRequest_FileByFilename:
			data, err := r.getFileByFilename(messageRequest.FileByFilename, fileDescriptorsSent)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectionv1.FileDescriptorResponse{FileDescriptorProto: data},
				}
			}
		case *reflectionv1.ServerReflectionRequest_FileContainingSymbol:
			data, err := r.getFileContainingSymbol(
				messageRequest.FileContainingSymbol,
				fileDescriptorsSent,
			)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectionv1.FileDescriptorResponse{FileDescriptorProto: data},
				}
			}
		case *reflectionv1.ServerReflectionRequest_FileContainingExtension:
			msgFQN := messageRequest.FileContainingExtension.ContainingType
			extNumber := messageRequest.FileContainingExtension.ExtensionNumber
			data, err := r.getFileContainingExtension(msgFQN, extNumber, fileDescriptorsSent)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectionv1.FileDescriptorResponse{FileDescriptorProto: data},
				}
			}
		case *reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType:
			nums, err := r.getAllExtensionNumbersOfType(messageRequest.AllExtensionNumbersOfType)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_AllExtensionNumbersResponse{
					AllExtensionNumbersResponse: &reflectionv1.ExtensionNumberResponse{
						BaseTypeName:    messageRequest.AllExtensionNumbersOfType,
						ExtensionNumber: nums,
					},
				}
			}
		case *reflectionv1.ServerReflectionRequest_ListServices:
			services := r.namer.Names()
			serviceResponses := make([]*reflectionv1.ServiceResponse, len(services))
			for i, name := range services {
				serviceResponses[i] = &reflectionv1.ServiceResponse{Name: name}
			}
			response.MessageResponse = &reflectionv1.ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &reflectionv1.ListServiceResponse{Service: serviceResponses},
			}
		default:
			return connect.Errorf(connect.CodeInvalidArgument,
				"invalid MessageRequest: %v",
				request.MessageRequest,
			)
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (r *reflector) getFileByFilename(fname string, sent *fileDescriptorNameSet) ([][]byte, error) {
	fd, err := r.descriptorResolver.FindFileByPath(fname)
	if err != nil {
		return nil, err
	}
	return fileDescriptorWithDependencies(fd, sent)
}

func (r *reflector) getFileContainingSymbol(fqn string, sent *fileDescriptorNameSet) ([][]byte, error) {
	desc, err := r.descriptorResolver.FindDescriptorByName(protoreflect.FullName(fqn))
	if err != nil {
		return nil, err
	}
	fd := desc.ParentFile()
	if fd == nil {
		return nil, fmt.Errorf("no file for symbol %s", fqn)
	}
	return fileDescriptorWithDependencies(fd, sent)
}

func (r *reflector) getFileContainingExtension(
	msgFQN string,
	extNumber int32,
	sent *fileDescriptorNameSet,
) ([][]byte, error) {
	extension, err := r.extensionResolver.FindExtensionByNumber(
		protoreflect.FullName(msgFQN),
		protoreflect.FieldNumber(extNumber),
	)
	if err != nil {
		return nil, err
	}
	fd := extension.TypeDescriptor().ParentFile()
	if fd == nil {
		return nil, fmt.Errorf("no file for extension %d of message %s", extNumber, msgFQN)
	}
	return fileDescriptorWithDependencies(fd, sent)
}

func (r *reflector) getAllExtensionNumbersOfType(fqn string) ([]int32, error) {
	nums := []int32{}
	name := protoreflect.FullName(fqn)
	r.extensionResolver.RangeExtensionsByMessage(name, func(ext protoreflect.ExtensionType) bool {
		num := int32(ext.TypeDescriptor().Number())
		nums = append(nums, num)
		return true
	})
	if len(nums) == 0 {
		if _, err := r.descriptorResolver.FindDescriptorByName(name); err != nil {
			return nil, err
		}
	}
	slices.Sort(nums)
	return nums, nil
}

// A Namer lists the fully-qualified Protobuf service names available for
// reflection (for example, "acme.user.v1.UserService"). Namers must be safe to
// call concurrently.
type Namer interface {
	Names() []string
}

// NamerFunc is an adapter to allow the use of an ordinary function as a Namer.
// Example:
//
//	grpcreflect.Register(server, grpcreflect.WithNamer(grpcreflect.NamerFunc(
//		func() []string { return s.names },
//	)))
type NamerFunc func() []string

// Names returns the service names, implements the Namer interface.
func (f NamerFunc) Names() []string {
	return f()
}

// An Option configures the reflection services registered by Register.
type Option interface {
	apply(*reflector)
}

// WithNamer sets the Namer that lists the services exposed by reflection.
// By default, reflection describes the services registered on the server
// it's registered with.
func WithNamer(namer Namer) Option {
	return &namerOption{namer: namer}
}

// WithExtensionResolver sets the resolver used to find Protobuf extensions. By
// default, reflection uses protoregistry.GlobalTypes.
func WithExtensionResolver(resolver ExtensionResolver) Option {
	return &extensionResolverOption{resolver: resolver}
}

// WithDescriptorResolver sets the resolver used to find Protobuf type
// information (typically called a "descriptor"). By default, reflection uses
// protoregistry.GlobalFiles.
func WithDescriptorResolver(resolver protodesc.Resolver) Option {
	return &descriptorResolverOption{resolver: resolver}
}

// An ExtensionResolver lets server reflection implementations query details
// about the registered Protobuf extensions. protoregistry.GlobalTypes
// implements ExtensionResolver.
//
// ExtensionResolvers must be safe to call concurrently.
type ExtensionResolver interface {
	protoregistry.ExtensionTypeResolver

	RangeExtensionsByMessage(protoreflect.FullName, func(protoreflect.ExtensionType) bool)
}

type fileDescriptorNameSet struct {
	names map[string]struct{}
}

func (s *fileDescriptorNameSet) Insert(fd protoreflect.FileDescriptor) {
	if s.names == nil {
		s.names = make(map[string]struct{}, 1)
	}
	s.names[fd.Path()] = struct{}{}
}

func (s *fileDescriptorNameSet) Contains(fd protoreflect.FileDescriptor) bool {
	_, ok := s.names[fd.Path()]
	return ok
}

func fileDescriptorWithDependencies(rootFile protoreflect.FileDescriptor, sent *fileDescriptorNameSet) ([][]byte, error) {
	if rootFile.IsPlaceholder() {
		// A placeholder is used when a dependency is missing. If a placeholder is all we have
		// then we don't actually have anything.
		return nil, protoregistry.NotFound
	}
	results := make([][]byte, 0, 1)
	queue := []protoreflect.FileDescriptor{rootFile}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if curr.IsPlaceholder() {
			continue // don't bother serializing placeholders
		}
		if len(results) == 0 || !sent.Contains(curr) { // always send root fd
			// Mark as sent immediately. If we hit an error marshaling below, there's
			// no point trying again later.
			sent.Insert(curr)
			encoded, err := proto.Marshal(protodesc.ToFileDescriptorProto(curr))
			if err != nil {
				return nil, err
			}
			results = append(results, encoded)
		}
		imports := curr.Imports()
		for i := range imports.Len() {
			queue = append(queue, imports.Get(i).FileDescriptor)
		}
	}
	return results, nil
}

func newNotFoundResponse(err error) *reflectionv1.ServerReflectionResponse_ErrorResponse {
	return &reflectionv1.ServerReflectionResponse_ErrorResponse{
		ErrorResponse: &reflectionv1.ErrorResponse{
			ErrorCode:    int32(connect.CodeNotFound),
			ErrorMessage: err.Error(),
		},
	}
}

type namerOption struct {
	namer Namer
}

func (o *namerOption) apply(reflector *reflector) {
	reflector.namer = o.namer
}

type extensionResolverOption struct {
	resolver ExtensionResolver
}

func (o *extensionResolverOption) apply(reflector *reflector) {
	reflector.extensionResolver = o.resolver
}

type descriptorResolverOption struct {
	resolver protodesc.Resolver
}

func (o *descriptorResolverOption) apply(reflector *reflector) {
	reflector.descriptorResolver = o.resolver
}

// serverNamer lists the services registered on a connect.Server. Names are
// derived lazily from the server's specs, so services registered after
// reflection still appear. Methods without a protobuf schema are skipped.
type serverNamer struct {
	server *connect.Server
}

func (n *serverNamer) Names() []string {
	seen := make(map[string]struct{})
	for spec := range n.server.Specs() {
		method, ok := spec.Schema.(protoreflect.MethodDescriptor)
		if !ok {
			continue
		}
		seen[string(method.Parent().FullName())] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for service := range seen {
		names = append(names, service)
	}
	sort.Strings(names)
	return names
}

// resolverHackForConnectext returns a resolver that can successfully resolve the descriptors
// for the gRPC health and reflection services. We need a work-around since this repo (and the
// connect-grpchealth-go repo) use "hacked" services that have a "connectext." package prefix.
// We don't use the "authoritative" packages for these descriptors because they depend on the
// gRPC runtime (ew!). We add a special prefix to the packages to avoid an init-time panic from
// duplicate registrations, in the event that the calling application _also_ imports the gRPC
// versions.
//
// This works by serving embedded descriptors (from "services.bin") for items not found in
// protoregistry.GlobalFiles. The only thing in the embedded descriptors are for the health
// and reflection services.
func resolverHackForConnectext(data []byte) protodesc.Resolver {
	var backupResolver protodesc.Resolver
	var fileSet descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &fileSet); err != nil {
		backupResolver = &errResolver{err}
	} else if files, err := protodesc.NewFiles(&fileSet); err != nil {
		backupResolver = &errResolver{err}
	} else {
		backupResolver = files
	}

	return &combinedResolver{
		first:  protoregistry.GlobalFiles,
		second: backupResolver,
	}
}

type combinedResolver struct {
	first, second protodesc.Resolver
}

func (r *combinedResolver) FindFileByPath(s string) (protoreflect.FileDescriptor, error) {
	file, err := r.first.FindFileByPath(s)
	if errors.Is(err, protoregistry.NotFound) {
		file, err = r.second.FindFileByPath(s)
	}
	return file, err
}

func (r *combinedResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	desc, err := r.first.FindDescriptorByName(name)
	if errors.Is(err, protoregistry.NotFound) {
		desc, err = r.second.FindDescriptorByName(name)
	}
	return desc, err
}

type errResolver struct {
	err error
}

func (r *errResolver) FindFileByPath(_ string) (protoreflect.FileDescriptor, error) {
	return nil, r.err
}

func (r *errResolver) FindDescriptorByName(_ protoreflect.FullName) (protoreflect.Descriptor, error) {
	return nil, r.err
}
