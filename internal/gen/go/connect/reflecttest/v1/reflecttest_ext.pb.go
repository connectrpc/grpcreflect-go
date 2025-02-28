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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: connect/reflecttest/v1/reflecttest_ext.proto

package reflecttestv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_connect_reflecttest_v1_reflecttest_ext_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*Extendable)(nil),
		ExtensionType: (*string)(nil),
		Field:         10,
		Name:          "connect.reflecttest.v1.message",
		Tag:           "bytes,10,opt,name=message",
		Filename:      "connect/reflecttest/v1/reflecttest_ext.proto",
	},
	{
		ExtendedType:  (*Extendable)(nil),
		ExtensionType: (*string)(nil),
		Field:         11,
		Name:          "connect.reflecttest.v1.localized_message",
		Tag:           "bytes,11,opt,name=localized_message",
		Filename:      "connect/reflecttest/v1/reflecttest_ext.proto",
	},
}

// Extension fields to Extendable.
var (
	// optional string message = 10;
	E_Message = &file_connect_reflecttest_v1_reflecttest_ext_proto_extTypes[0]
	// optional string localized_message = 11;
	E_LocalizedMessage = &file_connect_reflecttest_v1_reflecttest_ext_proto_extTypes[1]
)

var File_connect_reflecttest_v1_reflecttest_ext_proto protoreflect.FileDescriptor

var file_connect_reflecttest_v1_reflecttest_ext_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2f, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63,
	0x74, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74,
	0x74, 0x65, 0x73, 0x74, 0x5f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74,
	0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x28, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2f,
	0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x72,
	0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x3a, 0x3c, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x22, 0x2e, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73,
	0x74, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x64, 0x61, 0x62, 0x6c, 0x65, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x3a, 0x4f,
	0x0a, 0x11, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x5f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x22, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x72, 0x65,
	0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x78, 0x74,
	0x65, 0x6e, 0x64, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x6c,
	0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x42,
	0xfc, 0x01, 0x0a, 0x1a, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e,
	0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x76, 0x31, 0x42, 0x13,
	0x52, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x45, 0x78, 0x74, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x72, 0x70,
	0x63, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63,
	0x74, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x67,
	0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2f, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63,
	0x74, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x76, 0x31, 0x3b, 0x72, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74,
	0x74, 0x65, 0x73, 0x74, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x52, 0x58, 0xaa, 0x02, 0x16, 0x43,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x52, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65,
	0x73, 0x74, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x16, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5c,
	0x52, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x5c, 0x56, 0x31, 0xe2, 0x02,
	0x22, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5c, 0x52, 0x65, 0x66, 0x6c, 0x65, 0x63, 0x74,
	0x74, 0x65, 0x73, 0x74, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x18, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x3a, 0x3a, 0x52,
	0x65, 0x66, 0x6c, 0x65, 0x63, 0x74, 0x74, 0x65, 0x73, 0x74, 0x3a, 0x3a, 0x56, 0x31,
}

var file_connect_reflecttest_v1_reflecttest_ext_proto_goTypes = []interface{}{
	(*Extendable)(nil), // 0: connect.reflecttest.v1.Extendable
}
var file_connect_reflecttest_v1_reflecttest_ext_proto_depIdxs = []int32{
	0, // 0: connect.reflecttest.v1.message:extendee -> connect.reflecttest.v1.Extendable
	0, // 1: connect.reflecttest.v1.localized_message:extendee -> connect.reflecttest.v1.Extendable
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	0, // [0:2] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_connect_reflecttest_v1_reflecttest_ext_proto_init() }
func file_connect_reflecttest_v1_reflecttest_ext_proto_init() {
	if File_connect_reflecttest_v1_reflecttest_ext_proto != nil {
		return
	}
	file_connect_reflecttest_v1_reflecttest_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_connect_reflecttest_v1_reflecttest_ext_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 2,
			NumServices:   0,
		},
		GoTypes:           file_connect_reflecttest_v1_reflecttest_ext_proto_goTypes,
		DependencyIndexes: file_connect_reflecttest_v1_reflecttest_ext_proto_depIdxs,
		ExtensionInfos:    file_connect_reflecttest_v1_reflecttest_ext_proto_extTypes,
	}.Build()
	File_connect_reflecttest_v1_reflecttest_ext_proto = out.File
	file_connect_reflecttest_v1_reflecttest_ext_proto_rawDesc = nil
	file_connect_reflecttest_v1_reflecttest_ext_proto_goTypes = nil
	file_connect_reflecttest_v1_reflecttest_ext_proto_depIdxs = nil
}
