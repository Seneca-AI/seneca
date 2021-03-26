// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.6
// source: raw.proto

package types

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

// key(user_id, id)
type RawVideo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId string `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Id     string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	// In the form <USER_ID>.<CREATE_TIMESTAMP_MS>.RAW_VIDEO.mp4
	CloudStorageFileName string `protobuf:"bytes,3,opt,name=cloud_storage_file_name,json=cloudStorageFileName,proto3" json:"cloud_storage_file_name,omitempty"`
	// When the video begins in ms.
	CreateTimeMs int64    `protobuf:"varint,4,opt,name=create_time_ms,json=createTimeMs,proto3" json:"create_time_ms,omitempty"`
	CutVideoId   []string `protobuf:"bytes,5,rep,name=cut_video_id,json=cutVideoId,proto3" json:"cut_video_id,omitempty"`
}

func (x *RawVideo) Reset() {
	*x = RawVideo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_raw_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RawVideo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RawVideo) ProtoMessage() {}

func (x *RawVideo) ProtoReflect() protoreflect.Message {
	mi := &file_raw_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RawVideo.ProtoReflect.Descriptor instead.
func (*RawVideo) Descriptor() ([]byte, []int) {
	return file_raw_proto_rawDescGZIP(), []int{0}
}

func (x *RawVideo) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *RawVideo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RawVideo) GetCloudStorageFileName() string {
	if x != nil {
		return x.CloudStorageFileName
	}
	return ""
}

func (x *RawVideo) GetCreateTimeMs() int64 {
	if x != nil {
		return x.CreateTimeMs
	}
	return 0
}

func (x *RawVideo) GetCutVideoId() []string {
	if x != nil {
		return x.CutVideoId
	}
	return nil
}

var File_raw_proto protoreflect.FileDescriptor

var file_raw_proto_rawDesc = []byte{
	0x0a, 0x09, 0x72, 0x61, 0x77, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x61, 0x70, 0x69,
	0x22, 0xb2, 0x01, 0x0a, 0x08, 0x52, 0x61, 0x77, 0x56, 0x69, 0x64, 0x65, 0x6f, 0x12, 0x17, 0x0a,
	0x07, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x75, 0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x35, 0x0a, 0x17, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x5f,
	0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x53, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a,
	0x0e, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x6d, 0x73, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d,
	0x65, 0x4d, 0x73, 0x12, 0x20, 0x0a, 0x0c, 0x63, 0x75, 0x74, 0x5f, 0x76, 0x69, 0x64, 0x65, 0x6f,
	0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x75, 0x74, 0x56, 0x69,
	0x64, 0x65, 0x6f, 0x49, 0x64, 0x42, 0x19, 0x5a, 0x17, 0x73, 0x65, 0x6e, 0x65, 0x63, 0x61, 0x2f,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_raw_proto_rawDescOnce sync.Once
	file_raw_proto_rawDescData = file_raw_proto_rawDesc
)

func file_raw_proto_rawDescGZIP() []byte {
	file_raw_proto_rawDescOnce.Do(func() {
		file_raw_proto_rawDescData = protoimpl.X.CompressGZIP(file_raw_proto_rawDescData)
	})
	return file_raw_proto_rawDescData
}

var file_raw_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_raw_proto_goTypes = []interface{}{
	(*RawVideo)(nil), // 0: api.RawVideo
}
var file_raw_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_raw_proto_init() }
func file_raw_proto_init() {
	if File_raw_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_raw_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RawVideo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_raw_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_raw_proto_goTypes,
		DependencyIndexes: file_raw_proto_depIdxs,
		MessageInfos:      file_raw_proto_msgTypes,
	}.Build()
	File_raw_proto = out.File
	file_raw_proto_rawDesc = nil
	file_raw_proto_goTypes = nil
	file_raw_proto_depIdxs = nil
}
