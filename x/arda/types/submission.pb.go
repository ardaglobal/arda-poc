// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ardapoc/arda/submission.proto

package types

import (
	fmt "fmt"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type Submission struct {
	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Creator string `protobuf:"bytes,2,opt,name=creator,proto3" json:"creator,omitempty"`
	Region  string `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty"`
	Hash    string `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	Valid   string `protobuf:"bytes,5,opt,name=valid,proto3" json:"valid,omitempty"`
}

func (m *Submission) Reset()         { *m = Submission{} }
func (m *Submission) String() string { return proto.CompactTextString(m) }
func (*Submission) ProtoMessage()    {}
func (*Submission) Descriptor() ([]byte, []int) {
	return fileDescriptor_276a3f0e17cb302d, []int{0}
}
func (m *Submission) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Submission) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Submission.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Submission) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Submission.Merge(m, src)
}
func (m *Submission) XXX_Size() int {
	return m.Size()
}
func (m *Submission) XXX_DiscardUnknown() {
	xxx_messageInfo_Submission.DiscardUnknown(m)
}

var xxx_messageInfo_Submission proto.InternalMessageInfo

func (m *Submission) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Submission) GetCreator() string {
	if m != nil {
		return m.Creator
	}
	return ""
}

func (m *Submission) GetRegion() string {
	if m != nil {
		return m.Region
	}
	return ""
}

func (m *Submission) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *Submission) GetValid() string {
	if m != nil {
		return m.Valid
	}
	return ""
}

func init() {
	proto.RegisterType((*Submission)(nil), "ardapoc.arda.Submission")
}

func init() { proto.RegisterFile("ardapoc/arda/submission.proto", fileDescriptor_276a3f0e17cb302d) }

var fileDescriptor_276a3f0e17cb302d = []byte{
	// 209 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4d, 0x2c, 0x4a, 0x49,
	0x2c, 0xc8, 0x4f, 0xd6, 0x07, 0xd1, 0xfa, 0xc5, 0xa5, 0x49, 0xb9, 0x99, 0xc5, 0xc5, 0x99, 0xf9,
	0x79, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0x3c, 0x50, 0x69, 0x3d, 0x10, 0xad, 0x54, 0xc1,
	0xc5, 0x15, 0x0c, 0x57, 0x21, 0xc4, 0xc7, 0xc5, 0x94, 0x99, 0x22, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1,
	0x19, 0xc4, 0x94, 0x99, 0x22, 0x24, 0xc1, 0xc5, 0x9e, 0x5c, 0x94, 0x9a, 0x58, 0x92, 0x5f, 0x24,
	0xc1, 0x04, 0x16, 0x84, 0x71, 0x85, 0xc4, 0xb8, 0xd8, 0x8a, 0x52, 0xd3, 0x33, 0xf3, 0xf3, 0x24,
	0x98, 0xc1, 0x12, 0x50, 0x9e, 0x90, 0x10, 0x17, 0x4b, 0x46, 0x62, 0x71, 0x86, 0x04, 0x0b, 0x58,
	0x14, 0xcc, 0x16, 0x12, 0xe1, 0x62, 0x2d, 0x4b, 0xcc, 0xc9, 0x4c, 0x91, 0x60, 0x05, 0x0b, 0x42,
	0x38, 0x4e, 0xae, 0x27, 0x1e, 0xc9, 0x31, 0x5e, 0x78, 0x24, 0xc7, 0xf8, 0xe0, 0x91, 0x1c, 0xe3,
	0x84, 0xc7, 0x72, 0x0c, 0x17, 0x1e, 0xcb, 0x31, 0xdc, 0x78, 0x2c, 0xc7, 0x10, 0xa5, 0x9d, 0x9e,
	0x59, 0x92, 0x51, 0x9a, 0xa4, 0x97, 0x9c, 0x9f, 0x0b, 0xf6, 0x43, 0x7a, 0x4e, 0x7e, 0x52, 0x62,
	0x0e, 0x98, 0xa9, 0x0b, 0xf2, 0x57, 0x05, 0xc4, 0x67, 0x25, 0x95, 0x05, 0xa9, 0xc5, 0x49, 0x6c,
	0x60, 0x5f, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0xf1, 0xee, 0x58, 0x17, 0xf6, 0x00, 0x00,
	0x00,
}

func (m *Submission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Submission) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Submission) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Valid) > 0 {
		i -= len(m.Valid)
		copy(dAtA[i:], m.Valid)
		i = encodeVarintSubmission(dAtA, i, uint64(len(m.Valid)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Hash) > 0 {
		i -= len(m.Hash)
		copy(dAtA[i:], m.Hash)
		i = encodeVarintSubmission(dAtA, i, uint64(len(m.Hash)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Region) > 0 {
		i -= len(m.Region)
		copy(dAtA[i:], m.Region)
		i = encodeVarintSubmission(dAtA, i, uint64(len(m.Region)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Creator) > 0 {
		i -= len(m.Creator)
		copy(dAtA[i:], m.Creator)
		i = encodeVarintSubmission(dAtA, i, uint64(len(m.Creator)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintSubmission(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintSubmission(dAtA []byte, offset int, v uint64) int {
	offset -= sovSubmission(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Submission) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovSubmission(uint64(l))
	}
	l = len(m.Creator)
	if l > 0 {
		n += 1 + l + sovSubmission(uint64(l))
	}
	l = len(m.Region)
	if l > 0 {
		n += 1 + l + sovSubmission(uint64(l))
	}
	l = len(m.Hash)
	if l > 0 {
		n += 1 + l + sovSubmission(uint64(l))
	}
	l = len(m.Valid)
	if l > 0 {
		n += 1 + l + sovSubmission(uint64(l))
	}
	return n
}

func sovSubmission(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozSubmission(x uint64) (n int) {
	return sovSubmission(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Submission) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSubmission
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Submission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Submission: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSubmission
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSubmission
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Creator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSubmission
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSubmission
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Creator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Region", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSubmission
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSubmission
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Region = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSubmission
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSubmission
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Valid", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSubmission
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthSubmission
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Valid = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSubmission(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthSubmission
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipSubmission(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSubmission
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSubmission
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthSubmission
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupSubmission
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthSubmission
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthSubmission        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSubmission          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupSubmission = fmt.Errorf("proto: unexpected end of group")
)
