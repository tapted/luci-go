// Code generated by protoc-gen-go. DO NOT EDIT.
// source: go.chromium.org/luci/cq/api/bigquery/attempt.proto

package bigquery

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Whether the attempt is a dry run or full run. If constituent CLs have
// different modes, then the mode is disagreement.
type Attempt_Mode int32

const (
	Attempt_DISAGREEMENT Attempt_Mode = 0
	Attempt_DRY_RUN      Attempt_Mode = 1
	Attempt_FULL_RUN     Attempt_Mode = 2
)

var Attempt_Mode_name = map[int32]string{
	0: "DISAGREEMENT",
	1: "DRY_RUN",
	2: "FULL_RUN",
}

var Attempt_Mode_value = map[string]int32{
	"DISAGREEMENT": 0,
	"DRY_RUN":      1,
	"FULL_RUN":     2,
}

func (x Attempt_Mode) String() string {
	return proto.EnumName(Attempt_Mode_name, int32(x))
}

func (Attempt_Mode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8792fe122a6ce934, []int{0, 0}
}

// Pass or fail state of the attempt. Pending attempts shouldn't be included
// in the completed attempts table.
type Attempt_SubmitState int32

const (
	Attempt_QUEUED    Attempt_SubmitState = 0
	Attempt_FAILED    Attempt_SubmitState = 1
	Attempt_SUBMITTED Attempt_SubmitState = 2
)

var Attempt_SubmitState_name = map[int32]string{
	0: "QUEUED",
	1: "FAILED",
	2: "SUBMITTED",
}

var Attempt_SubmitState_value = map[string]int32{
	"QUEUED":    0,
	"FAILED":    1,
	"SUBMITTED": 2,
}

func (x Attempt_SubmitState) String() string {
	return proto.EnumName(Attempt_SubmitState_name, int32(x))
}

func (Attempt_SubmitState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8792fe122a6ce934, []int{0, 1}
}

// Attempt includes the state CQ attempt.
// For thecompleted
type Attempt struct {
	// Time when the attempt started and stopped (TODO: define more precisely)
	Start                *timestamp.Timestamp `protobuf:"bytes,1,opt,name=start,proto3" json:"start,omitempty"`
	Stop                 *timestamp.Timestamp `protobuf:"bytes,2,opt,name=stop,proto3" json:"stop,omitempty"`
	Mode                 Attempt_Mode         `protobuf:"varint,3,opt,name=mode,proto3,enum=bigquery.Attempt_Mode" json:"mode,omitempty"`
	State                Attempt_SubmitState  `protobuf:"varint,4,opt,name=state,proto3,enum=bigquery.Attempt_SubmitState" json:"state,omitempty"`
	AttemptKey           string               `protobuf:"bytes,5,opt,name=attempt_key,json=attemptKey,proto3" json:"attempt_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *Attempt) Reset()         { *m = Attempt{} }
func (m *Attempt) String() string { return proto.CompactTextString(m) }
func (*Attempt) ProtoMessage()    {}
func (*Attempt) Descriptor() ([]byte, []int) {
	return fileDescriptor_8792fe122a6ce934, []int{0}
}

func (m *Attempt) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Attempt.Unmarshal(m, b)
}
func (m *Attempt) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Attempt.Marshal(b, m, deterministic)
}
func (m *Attempt) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Attempt.Merge(m, src)
}
func (m *Attempt) XXX_Size() int {
	return xxx_messageInfo_Attempt.Size(m)
}
func (m *Attempt) XXX_DiscardUnknown() {
	xxx_messageInfo_Attempt.DiscardUnknown(m)
}

var xxx_messageInfo_Attempt proto.InternalMessageInfo

func (m *Attempt) GetStart() *timestamp.Timestamp {
	if m != nil {
		return m.Start
	}
	return nil
}

func (m *Attempt) GetStop() *timestamp.Timestamp {
	if m != nil {
		return m.Stop
	}
	return nil
}

func (m *Attempt) GetMode() Attempt_Mode {
	if m != nil {
		return m.Mode
	}
	return Attempt_DISAGREEMENT
}

func (m *Attempt) GetState() Attempt_SubmitState {
	if m != nil {
		return m.State
	}
	return Attempt_QUEUED
}

func (m *Attempt) GetAttemptKey() string {
	if m != nil {
		return m.AttemptKey
	}
	return ""
}

func init() {
	proto.RegisterEnum("bigquery.Attempt_Mode", Attempt_Mode_name, Attempt_Mode_value)
	proto.RegisterEnum("bigquery.Attempt_SubmitState", Attempt_SubmitState_name, Attempt_SubmitState_value)
	proto.RegisterType((*Attempt)(nil), "bigquery.Attempt")
}

func init() {
	proto.RegisterFile("go.chromium.org/luci/cq/api/bigquery/attempt.proto", fileDescriptor_8792fe122a6ce934)
}

var fileDescriptor_8792fe122a6ce934 = []byte{
	// 317 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0x4f, 0x6b, 0xfa, 0x30,
	0x1c, 0xc6, 0x6d, 0x7f, 0xf5, 0xdf, 0xb7, 0xfe, 0x46, 0xc9, 0x61, 0x14, 0x61, 0x28, 0x9e, 0x64,
	0x87, 0x64, 0xe8, 0xde, 0x80, 0xa3, 0x71, 0xc8, 0x54, 0x58, 0xda, 0x1e, 0x76, 0x92, 0x56, 0xb3,
	0x2e, 0xcc, 0x90, 0x5a, 0xd3, 0x83, 0xef, 0x74, 0x2f, 0x67, 0x98, 0xb6, 0x30, 0xd8, 0x61, 0xb7,
	0xe4, 0xe1, 0xf3, 0xe1, 0xfb, 0xf0, 0xc0, 0x2c, 0x53, 0x78, 0xff, 0x51, 0x28, 0x29, 0x4a, 0x89,
	0x55, 0x91, 0x91, 0x63, 0xb9, 0x17, 0x64, 0x7f, 0x22, 0x49, 0x2e, 0x48, 0x2a, 0xb2, 0x53, 0xc9,
	0x8b, 0x0b, 0x49, 0xb4, 0xe6, 0x32, 0xd7, 0x38, 0x2f, 0x94, 0x56, 0xa8, 0xd7, 0xe4, 0xc3, 0x51,
	0xa6, 0x54, 0x76, 0xe4, 0xc4, 0xe4, 0x69, 0xf9, 0x4e, 0xb4, 0x90, 0xfc, 0xac, 0x13, 0x99, 0x57,
	0xe8, 0xe4, 0xcb, 0x86, 0xee, 0xa2, 0x92, 0xd1, 0x03, 0xb4, 0xcf, 0x3a, 0x29, 0xb4, 0x6f, 0x8d,
	0xad, 0xa9, 0x3b, 0x1b, 0xe2, 0x4a, 0xc6, 0x8d, 0x8c, 0xa3, 0x46, 0x66, 0x15, 0x88, 0x30, 0x38,
	0x67, 0xad, 0x72, 0xdf, 0xfe, 0x53, 0x30, 0x1c, 0xba, 0x07, 0x47, 0xaa, 0x03, 0xf7, 0xff, 0x8d,
	0xad, 0xe9, 0xcd, 0xec, 0x16, 0x37, 0x3d, 0x71, 0x5d, 0x01, 0x6f, 0xd4, 0x81, 0x33, 0xc3, 0xa0,
	0xb9, 0x69, 0xa3, 0xb9, 0xef, 0x18, 0xf8, 0xee, 0x37, 0x1c, 0x96, 0xa9, 0x14, 0x3a, 0xbc, 0x42,
	0xac, 0x62, 0xd1, 0x08, 0xdc, 0x7a, 0x8a, 0xdd, 0x27, 0xbf, 0xf8, 0xed, 0xb1, 0x35, 0xed, 0x33,
	0xa8, 0xa3, 0x17, 0x7e, 0x99, 0xcc, 0xc1, 0xb9, 0xde, 0x40, 0x1e, 0x0c, 0x82, 0x55, 0xb8, 0x78,
	0x66, 0x94, 0x6e, 0xe8, 0x36, 0xf2, 0x5a, 0xc8, 0x85, 0x6e, 0xc0, 0xde, 0x76, 0x2c, 0xde, 0x7a,
	0x16, 0x1a, 0x40, 0x6f, 0x19, 0xaf, 0xd7, 0xe6, 0x67, 0x4f, 0x1e, 0xc1, 0xfd, 0x71, 0x0b, 0x01,
	0x74, 0x5e, 0x63, 0x1a, 0xd3, 0xc0, 0x6b, 0x5d, 0xdf, 0xcb, 0xc5, 0x6a, 0x4d, 0x03, 0xcf, 0x42,
	0xff, 0xa1, 0x1f, 0xc6, 0x4f, 0x9b, 0x55, 0x14, 0xd1, 0xc0, 0xb3, 0xd3, 0x8e, 0x99, 0x61, 0xfe,
	0x1d, 0x00, 0x00, 0xff, 0xff, 0xfa, 0xc7, 0x3f, 0x5c, 0xc2, 0x01, 0x00, 0x00,
}
