// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.6
// source: common.proto

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

type Tag int32

const (
	Tag_UNKNOWN Tag = 0
	// Values: [0, 1000]
	Tag_TRAFFIC Tag = 1
	// Values: [0.0, 1000.0] hours
	Tag_LONG_DURATION Tag = 2
	// Values: [0, 1000] mph
	Tag_SPEED Tag = 3
	/// Values: [-1000, 1000] feet/s/s
	Tag_ACCELERATION Tag = 4
	// Values: [0, 1000] feet
	Tag_FOLLOWING_DISTANCE Tag = 5
	// Values: [0, 1000]
	Tag_NIGHT Tag = 6
	// Values: true, false
	Tag_WEEKEND Tag = 7
	// Values: [0, 1000]
	Tag_SNOW Tag = 8
	// Values: [0, 1000]
	Tag_RAIN Tag = 9
	// Values: [0, 1000]
	Tag_ICE Tag = 10
	// Values: [0, 1000]
	Tag_CITY Tag = 11
	// Values: [0, 1000]
	Tag_HIGH_WAY Tag = 12
)

// Enum value maps for Tag.
var (
	Tag_name = map[int32]string{
		0:  "UNKNOWN",
		1:  "TRAFFIC",
		2:  "LONG_DURATION",
		3:  "SPEED",
		4:  "ACCELERATION",
		5:  "FOLLOWING_DISTANCE",
		6:  "NIGHT",
		7:  "WEEKEND",
		8:  "SNOW",
		9:  "RAIN",
		10: "ICE",
		11: "CITY",
		12: "HIGH_WAY",
	}
	Tag_value = map[string]int32{
		"UNKNOWN":            0,
		"TRAFFIC":            1,
		"LONG_DURATION":      2,
		"SPEED":              3,
		"ACCELERATION":       4,
		"FOLLOWING_DISTANCE": 5,
		"NIGHT":              6,
		"WEEKEND":            7,
		"SNOW":               8,
		"RAIN":               9,
		"ICE":                10,
		"CITY":               11,
		"HIGH_WAY":           12,
	}
)

func (x Tag) Enum() *Tag {
	p := new(Tag)
	*p = x
	return p
}

func (x Tag) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Tag) Descriptor() protoreflect.EnumDescriptor {
	return file_common_proto_enumTypes[0].Descriptor()
}

func (Tag) Type() protoreflect.EnumType {
	return &file_common_proto_enumTypes[0]
}

func (x Tag) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Tag.Descriptor instead.
func (Tag) EnumDescriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{0}
}

type TimePeriod struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StartTimeMs int64 `protobuf:"varint,1,opt,name=start_time_ms,json=startTimeMs,proto3" json:"start_time_ms,omitempty"`
	EndTimeMs   int64 `protobuf:"varint,2,opt,name=end_time_ms,json=endTimeMs,proto3" json:"end_time_ms,omitempty"`
}

func (x *TimePeriod) Reset() {
	*x = TimePeriod{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimePeriod) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimePeriod) ProtoMessage() {}

func (x *TimePeriod) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimePeriod.ProtoReflect.Descriptor instead.
func (*TimePeriod) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{0}
}

func (x *TimePeriod) GetStartTimeMs() int64 {
	if x != nil {
		return x.StartTimeMs
	}
	return 0
}

func (x *TimePeriod) GetEndTimeMs() int64 {
	if x != nil {
		return x.EndTimeMs
	}
	return 0
}

type Location struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Latitude    float64 `protobuf:"fixed64,1,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude   float64 `protobuf:"fixed64,2,opt,name=longitude,proto3" json:"longitude,omitempty"`
	TimestampMs int64   `protobuf:"varint,3,opt,name=timestamp_ms,json=timestampMs,proto3" json:"timestamp_ms,omitempty"`
}

func (x *Location) Reset() {
	*x = Location{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Location) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Location) ProtoMessage() {}

func (x *Location) ProtoReflect() protoreflect.Message {
	mi := &file_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Location.ProtoReflect.Descriptor instead.
func (*Location) Descriptor() ([]byte, []int) {
	return file_common_proto_rawDescGZIP(), []int{1}
}

func (x *Location) GetLatitude() float64 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *Location) GetLongitude() float64 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *Location) GetTimestampMs() int64 {
	if x != nil {
		return x.TimestampMs
	}
	return 0
}

var File_common_proto protoreflect.FileDescriptor

var file_common_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03,
	0x61, 0x70, 0x69, 0x22, 0x50, 0x0a, 0x0a, 0x54, 0x69, 0x6d, 0x65, 0x50, 0x65, 0x72, 0x69, 0x6f,
	0x64, 0x12, 0x22, 0x0a, 0x0d, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f,
	0x6d, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54,
	0x69, 0x6d, 0x65, 0x4d, 0x73, 0x12, 0x1e, 0x0a, 0x0b, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x5f, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x65, 0x6e, 0x64, 0x54,
	0x69, 0x6d, 0x65, 0x4d, 0x73, 0x22, 0x67, 0x0a, 0x08, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x5f, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x0b, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x4d, 0x73, 0x2a, 0xb4,
	0x01, 0x0a, 0x03, 0x54, 0x61, 0x67, 0x12, 0x0b, 0x0a, 0x07, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x54, 0x52, 0x41, 0x46, 0x46, 0x49, 0x43, 0x10, 0x01,
	0x12, 0x11, 0x0a, 0x0d, 0x4c, 0x4f, 0x4e, 0x47, 0x5f, 0x44, 0x55, 0x52, 0x41, 0x54, 0x49, 0x4f,
	0x4e, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x53, 0x50, 0x45, 0x45, 0x44, 0x10, 0x03, 0x12, 0x10,
	0x0a, 0x0c, 0x41, 0x43, 0x43, 0x45, 0x4c, 0x45, 0x52, 0x41, 0x54, 0x49, 0x4f, 0x4e, 0x10, 0x04,
	0x12, 0x16, 0x0a, 0x12, 0x46, 0x4f, 0x4c, 0x4c, 0x4f, 0x57, 0x49, 0x4e, 0x47, 0x5f, 0x44, 0x49,
	0x53, 0x54, 0x41, 0x4e, 0x43, 0x45, 0x10, 0x05, 0x12, 0x09, 0x0a, 0x05, 0x4e, 0x49, 0x47, 0x48,
	0x54, 0x10, 0x06, 0x12, 0x0b, 0x0a, 0x07, 0x57, 0x45, 0x45, 0x4b, 0x45, 0x4e, 0x44, 0x10, 0x07,
	0x12, 0x08, 0x0a, 0x04, 0x53, 0x4e, 0x4f, 0x57, 0x10, 0x08, 0x12, 0x08, 0x0a, 0x04, 0x52, 0x41,
	0x49, 0x4e, 0x10, 0x09, 0x12, 0x07, 0x0a, 0x03, 0x49, 0x43, 0x45, 0x10, 0x0a, 0x12, 0x08, 0x0a,
	0x04, 0x43, 0x49, 0x54, 0x59, 0x10, 0x0b, 0x12, 0x0c, 0x0a, 0x08, 0x48, 0x49, 0x47, 0x48, 0x5f,
	0x57, 0x41, 0x59, 0x10, 0x0c, 0x42, 0x12, 0x5a, 0x10, 0x73, 0x65, 0x6e, 0x65, 0x63, 0x61, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_common_proto_rawDescOnce sync.Once
	file_common_proto_rawDescData = file_common_proto_rawDesc
)

func file_common_proto_rawDescGZIP() []byte {
	file_common_proto_rawDescOnce.Do(func() {
		file_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_proto_rawDescData)
	})
	return file_common_proto_rawDescData
}

var file_common_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_common_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_common_proto_goTypes = []interface{}{
	(Tag)(0),           // 0: api.Tag
	(*TimePeriod)(nil), // 1: api.TimePeriod
	(*Location)(nil),   // 2: api.Location
}
var file_common_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_proto_init() }
func file_common_proto_init() {
	if File_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimePeriod); i {
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
		file_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Location); i {
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
			RawDescriptor: file_common_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_common_proto_goTypes,
		DependencyIndexes: file_common_proto_depIdxs,
		EnumInfos:         file_common_proto_enumTypes,
		MessageInfos:      file_common_proto_msgTypes,
	}.Build()
	File_common_proto = out.File
	file_common_proto_rawDesc = nil
	file_common_proto_goTypes = nil
	file_common_proto_depIdxs = nil
}
