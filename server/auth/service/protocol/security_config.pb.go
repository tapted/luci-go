// Code generated by protoc-gen-go. DO NOT EDIT.
// source: go.chromium.org/luci/server/auth/service/protocol/security_config.proto

package protocol

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

// SecurityConfig is read from 'security.cfg' by Auth Service and distributed to
// all linked services (in its serialized form) as part of AuthDB proto.
//
// See AuthDB.security_config in replication.proto.
type SecurityConfig struct {
	// A list of regular expressions matching hostnames that should be recognized
	// as being a part of single LUCI deployment.
	//
	// Different microservices within a single LUCI deployment may trust each
	// other. This setting (coupled with the TLS certificate check) allows
	// a service to recognize that a target of an RPC is another internal service
	// belonging to the same LUCI deployment.
	//
	// '^' and '$' are implied. The regexp language is intersection of Python and
	// Golang regexp languages and thus should use only very standard features
	// common to both.
	//
	// Example: "(.*-dot-)?chromium-swarm\.appspot\.com".
	InternalServiceRegexp []string `protobuf:"bytes,1,rep,name=internal_service_regexp,json=internalServiceRegexp,proto3" json:"internal_service_regexp,omitempty"`
	XXX_NoUnkeyedLiteral  struct{} `json:"-"`
	XXX_unrecognized      []byte   `json:"-"`
	XXX_sizecache         int32    `json:"-"`
}

func (m *SecurityConfig) Reset()         { *m = SecurityConfig{} }
func (m *SecurityConfig) String() string { return proto.CompactTextString(m) }
func (*SecurityConfig) ProtoMessage()    {}
func (*SecurityConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb8e278d7923eeac, []int{0}
}

func (m *SecurityConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SecurityConfig.Unmarshal(m, b)
}
func (m *SecurityConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SecurityConfig.Marshal(b, m, deterministic)
}
func (m *SecurityConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SecurityConfig.Merge(m, src)
}
func (m *SecurityConfig) XXX_Size() int {
	return xxx_messageInfo_SecurityConfig.Size(m)
}
func (m *SecurityConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_SecurityConfig.DiscardUnknown(m)
}

var xxx_messageInfo_SecurityConfig proto.InternalMessageInfo

func (m *SecurityConfig) GetInternalServiceRegexp() []string {
	if m != nil {
		return m.InternalServiceRegexp
	}
	return nil
}

func init() {
	proto.RegisterType((*SecurityConfig)(nil), "components.auth.SecurityConfig")
}

func init() {
	proto.RegisterFile("go.chromium.org/luci/server/auth/service/protocol/security_config.proto", fileDescriptor_bb8e278d7923eeac)
}

var fileDescriptor_bb8e278d7923eeac = []byte{
	// 168 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x72, 0x4f, 0xcf, 0xd7, 0x4b,
	0xce, 0x28, 0xca, 0xcf, 0xcd, 0x2c, 0xcd, 0xd5, 0xcb, 0x2f, 0x4a, 0xd7, 0xcf, 0x29, 0x4d, 0xce,
	0xd4, 0x2f, 0x4e, 0x2d, 0x2a, 0x4b, 0x2d, 0xd2, 0x4f, 0x2c, 0x2d, 0xc9, 0x00, 0xb3, 0x33, 0x93,
	0x53, 0xf5, 0x0b, 0x8a, 0xf2, 0x4b, 0xf2, 0x93, 0xf3, 0x73, 0xf4, 0x8b, 0x53, 0x93, 0x4b, 0x8b,
	0x32, 0x4b, 0x2a, 0xe3, 0x93, 0xf3, 0xf3, 0xd2, 0x32, 0xd3, 0xf5, 0xc0, 0x12, 0x42, 0xfc, 0xc9,
	0xf9, 0xb9, 0x05, 0xf9, 0x79, 0xa9, 0x79, 0x25, 0xc5, 0x7a, 0x20, 0x7d, 0x4a, 0x1e, 0x5c, 0x7c,
	0xc1, 0x50, 0x95, 0xce, 0x60, 0x85, 0x42, 0x66, 0x5c, 0xe2, 0x99, 0x79, 0x25, 0xa9, 0x45, 0x79,
	0x89, 0x39, 0xf1, 0x50, 0x53, 0xe3, 0x8b, 0x52, 0xd3, 0x53, 0x2b, 0x0a, 0x24, 0x18, 0x15, 0x98,
	0x35, 0x38, 0x83, 0x44, 0x61, 0xd2, 0xc1, 0x10, 0xd9, 0x20, 0xb0, 0xa4, 0x93, 0x4d, 0x94, 0x15,
	0xc9, 0xae, 0xb4, 0x86, 0x31, 0x92, 0xd8, 0xc0, 0x2c, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x77, 0x66, 0xfc, 0x5f, 0xea, 0x00, 0x00, 0x00,
}