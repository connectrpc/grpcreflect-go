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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: connect/reflecttest/v1/reflecttest.proto

package reflecttestv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Extendable struct {
	state           protoimpl.MessageState
	sizeCache       protoimpl.SizeCache
	unknownFields   protoimpl.UnknownFields
	extensionFields protoimpl.ExtensionFields

	Number *int32 `protobuf:"varint,1,req,name=number" json:"number,omitempty"`
}

func (x *Extendable) Reset() {
	*x = Extendable{}
	if protoimpl.UnsafeEnabled {
		mi := &file_connect_reflecttest_v1_reflecttest_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Extendable) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Extendable) ProtoMessage() {}

func (x *Extendable) ProtoReflect() protoreflect.Message {
	mi := &file_connect_reflecttest_v1_reflecttest_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Extendable.ProtoReflect.Descriptor instead.
func (*Extendable) Descriptor() ([]byte, []int) {
	return file_connect_reflecttest_v1_reflecttest_proto_rawDescGZIP(), []int{0}
}

func (x *Extendable) GetNumber() int32 {
	if x != nil && x.Number != nil {
		return *x.Number
	}
	return 0
}

var File_connect_reflecttest_v1_reflecttest_proto protoreflect.FileDescriptor

var file_connect_reflecttest_v1_reflecttest_proto_rawDesc = []byte{
	0x0a, 0x28, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2f, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63,
	0x74, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74,
	0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x2e, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2e,
	0x76, 0x31, 0x22, 0x2a, 0x0a, 0x0a, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x64, 0x61, 0x62, 0x6c, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x02, 0x28, 0x05,
	0x52, 0x06, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x2a, 0x04, 0x08, 0x0a, 0x10, 0x1f, 0x42, 0x89,
	0x02, 0x0a, 0x1a, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x72,
	0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x42, 0x10, 0x52,
	0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x5f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x75,
	0x66, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2d, 0x67,
	0x72, 0x70, 0x63, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x2d, 0x67, 0x6f, 0x2f, 0x69, 0x6e,
	0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67, 0x6f, 0x2f, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2f, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73,
	0x74, 0x2f, 0x76, 0x31, 0x3b, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74,
	0x76, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x52, 0x58, 0xaa, 0x02, 0x16, 0x43, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x2e, 0x52, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x56,
	0x31, 0xca, 0x02, 0x16, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5c, 0x52, 0x65, 0x66, 0x6c,
	0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x22, 0x43, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x5c, 0x52, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x18, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x3a, 0x3a, 0x52, 0x65, 0x66, 0x6c, 0x65,
	0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x3a, 0x3a, 0x56, 0x31,
}

var (
	file_connect_reflecttest_v1_reflecttest_proto_rawDescOnce sync.Once
	file_connect_reflecttest_v1_reflecttest_proto_rawDescData = file_connect_reflecttest_v1_reflecttest_proto_rawDesc
)

func file_connect_reflecttest_v1_reflecttest_proto_rawDescGZIP() []byte {
	file_connect_reflecttest_v1_reflecttest_proto_rawDescOnce.Do(func() {
		file_connect_reflecttest_v1_reflecttest_proto_rawDescData = protoimpl.X.CompressGZIP(file_connect_reflecttest_v1_reflecttest_proto_rawDescData)
	})
	return file_connect_reflecttest_v1_reflecttest_proto_rawDescData
}

var file_connect_reflecttest_v1_reflecttest_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_connect_reflecttest_v1_reflecttest_proto_goTypes = []interface{}{
	(*Extendable)(nil), // 0: connect.reflecttest.v1.Extendable
}
var file_connect_reflecttest_v1_reflecttest_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_connect_reflecttest_v1_reflecttest_proto_init() }
func file_connect_reflecttest_v1_reflecttest_proto_init() {
	if File_connect_reflecttest_v1_reflecttest_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_connect_reflecttest_v1_reflecttest_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Extendable); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			case 3:
				return &v.extensionFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connect_reflecttest_v1_reflecttest_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_connect_reflecttest_v1_reflecttest_proto_goTypes,
		DependencyIndexes: file_connect_reflecttest_v1_reflecttest_proto_depIdxs,
		MessageInfos:      file_connect_reflecttest_v1_reflecttest_proto_msgTypes,
	}.Build()
	File_connect_reflecttest_v1_reflecttest_proto = out.File
	file_connect_reflecttest_v1_reflecttest_proto_rawDesc = nil
	file_connect_reflecttest_v1_reflecttest_proto_goTypes = nil
	file_connect_reflecttest_v1_reflecttest_proto_depIdxs = nil
}
