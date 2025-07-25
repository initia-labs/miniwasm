// Code generated by protoc-gen-go-pulsar. DO NOT EDIT.
package tokenfactoryv1

import (
	fmt "fmt"
	io "io"
	reflect "reflect"
	sync "sync"

	runtime "github.com/cosmos/cosmos-proto/runtime"
	_ "github.com/cosmos/gogoproto/gogoproto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

var (
	md_DenomAuthorityMetadata       protoreflect.MessageDescriptor
	fd_DenomAuthorityMetadata_admin protoreflect.FieldDescriptor
)

func init() {
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_init()
	md_DenomAuthorityMetadata = File_miniwasm_tokenfactory_v1_authority_metadata_proto.Messages().ByName("DenomAuthorityMetadata")
	fd_DenomAuthorityMetadata_admin = md_DenomAuthorityMetadata.Fields().ByName("admin")
}

var _ protoreflect.Message = (*fastReflection_DenomAuthorityMetadata)(nil)

type fastReflection_DenomAuthorityMetadata DenomAuthorityMetadata

func (x *DenomAuthorityMetadata) ProtoReflect() protoreflect.Message {
	return (*fastReflection_DenomAuthorityMetadata)(x)
}

func (x *DenomAuthorityMetadata) slowProtoReflect() protoreflect.Message {
	mi := &file_miniwasm_tokenfactory_v1_authority_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_DenomAuthorityMetadata_messageType fastReflection_DenomAuthorityMetadata_messageType
var _ protoreflect.MessageType = fastReflection_DenomAuthorityMetadata_messageType{}

type fastReflection_DenomAuthorityMetadata_messageType struct{}

func (x fastReflection_DenomAuthorityMetadata_messageType) Zero() protoreflect.Message {
	return (*fastReflection_DenomAuthorityMetadata)(nil)
}
func (x fastReflection_DenomAuthorityMetadata_messageType) New() protoreflect.Message {
	return new(fastReflection_DenomAuthorityMetadata)
}
func (x fastReflection_DenomAuthorityMetadata_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_DenomAuthorityMetadata
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_DenomAuthorityMetadata) Descriptor() protoreflect.MessageDescriptor {
	return md_DenomAuthorityMetadata
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_DenomAuthorityMetadata) Type() protoreflect.MessageType {
	return _fastReflection_DenomAuthorityMetadata_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_DenomAuthorityMetadata) New() protoreflect.Message {
	return new(fastReflection_DenomAuthorityMetadata)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_DenomAuthorityMetadata) Interface() protoreflect.ProtoMessage {
	return (*DenomAuthorityMetadata)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_DenomAuthorityMetadata) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_DenomAuthorityMetadata_admin, value) {
			return
		}
	}
}

// Has reports whether a field is populated.
//
// Some fields have the property of nullability where it is possible to
// distinguish between the default value of a field and whether the field
// was explicitly populated with the default value. Singular message fields,
// member fields of a oneof, and proto2 scalar fields are nullable. Such
// fields are populated only if explicitly set.
//
// In other cases (aside from the nullable cases above),
// a proto3 scalar field is populated if it contains a non-zero value, and
// a repeated field is populated if it is non-empty.
func (x *fastReflection_DenomAuthorityMetadata) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "miniwasm.tokenfactory.v1.DenomAuthorityMetadata.admin":
		return x.Admin != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: miniwasm.tokenfactory.v1.DenomAuthorityMetadata"))
		}
		panic(fmt.Errorf("message miniwasm.tokenfactory.v1.DenomAuthorityMetadata does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_DenomAuthorityMetadata) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "miniwasm.tokenfactory.v1.DenomAuthorityMetadata.admin":
		x.Admin = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: miniwasm.tokenfactory.v1.DenomAuthorityMetadata"))
		}
		panic(fmt.Errorf("message miniwasm.tokenfactory.v1.DenomAuthorityMetadata does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_DenomAuthorityMetadata) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "miniwasm.tokenfactory.v1.DenomAuthorityMetadata.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: miniwasm.tokenfactory.v1.DenomAuthorityMetadata"))
		}
		panic(fmt.Errorf("message miniwasm.tokenfactory.v1.DenomAuthorityMetadata does not contain field %s", descriptor.FullName()))
	}
}

// Set stores the value for a field.
//
// For a field belonging to a oneof, it implicitly clears any other field
// that may be currently set within the same oneof.
// For extension fields, it implicitly stores the provided ExtensionType.
// When setting a composite type, it is unspecified whether the stored value
// aliases the source's memory in any way. If the composite value is an
// empty, read-only value, then it panics.
//
// Set is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_DenomAuthorityMetadata) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "miniwasm.tokenfactory.v1.DenomAuthorityMetadata.admin":
		x.Admin = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: miniwasm.tokenfactory.v1.DenomAuthorityMetadata"))
		}
		panic(fmt.Errorf("message miniwasm.tokenfactory.v1.DenomAuthorityMetadata does not contain field %s", fd.FullName()))
	}
}

// Mutable returns a mutable reference to a composite type.
//
// If the field is unpopulated, it may allocate a composite value.
// For a field belonging to a oneof, it implicitly clears any other field
// that may be currently set within the same oneof.
// For extension fields, it implicitly stores the provided ExtensionType
// if not already stored.
// It panics if the field does not contain a composite type.
//
// Mutable is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_DenomAuthorityMetadata) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "miniwasm.tokenfactory.v1.DenomAuthorityMetadata.admin":
		panic(fmt.Errorf("field admin of message miniwasm.tokenfactory.v1.DenomAuthorityMetadata is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: miniwasm.tokenfactory.v1.DenomAuthorityMetadata"))
		}
		panic(fmt.Errorf("message miniwasm.tokenfactory.v1.DenomAuthorityMetadata does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_DenomAuthorityMetadata) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "miniwasm.tokenfactory.v1.DenomAuthorityMetadata.admin":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: miniwasm.tokenfactory.v1.DenomAuthorityMetadata"))
		}
		panic(fmt.Errorf("message miniwasm.tokenfactory.v1.DenomAuthorityMetadata does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_DenomAuthorityMetadata) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in miniwasm.tokenfactory.v1.DenomAuthorityMetadata", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_DenomAuthorityMetadata) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_DenomAuthorityMetadata) SetUnknown(fields protoreflect.RawFields) {
	x.unknownFields = fields
}

// IsValid reports whether the message is valid.
//
// An invalid message is an empty, read-only value.
//
// An invalid message often corresponds to a nil pointer of the concrete
// message type, but the details are implementation dependent.
// Validity is not part of the protobuf data model, and may not
// be preserved in marshaling or other operations.
func (x *fastReflection_DenomAuthorityMetadata) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_DenomAuthorityMetadata) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*DenomAuthorityMetadata)
		if x == nil {
			return protoiface.SizeOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Size:              0,
			}
		}
		options := runtime.SizeInputToOptions(input)
		_ = options
		var n int
		var l int
		_ = l
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*DenomAuthorityMetadata)
		if x == nil {
			return protoiface.MarshalOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Buf:               input.Buf,
			}, nil
		}
		options := runtime.MarshalInputToOptions(input)
		_ = options
		size := options.Size(x)
		dAtA := make([]byte, size)
		i := len(dAtA)
		_ = i
		var l int
		_ = l
		if x.unknownFields != nil {
			i -= len(x.unknownFields)
			copy(dAtA[i:], x.unknownFields)
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
			i--
			dAtA[i] = 0xa
		}
		if input.Buf != nil {
			input.Buf = append(input.Buf, dAtA...)
		} else {
			input.Buf = dAtA
		}
		return protoiface.MarshalOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Buf:               input.Buf,
		}, nil
	}
	unmarshal := func(input protoiface.UnmarshalInput) (protoiface.UnmarshalOutput, error) {
		x := input.Message.Interface().(*DenomAuthorityMetadata)
		if x == nil {
			return protoiface.UnmarshalOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Flags:             input.Flags,
			}, nil
		}
		options := runtime.UnmarshalInputToOptions(input)
		_ = options
		dAtA := input.Buf
		l := len(dAtA)
		iNdEx := 0
		for iNdEx < l {
			preIndex := iNdEx
			var wire uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
				}
				if iNdEx >= l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: DenomAuthorityMetadata: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: DenomAuthorityMetadata: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
				}
				var stringLen uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
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
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + intStringLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			default:
				iNdEx = preIndex
				skippy, err := runtime.Skip(dAtA[iNdEx:])
				if err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				if (skippy < 0) || (iNdEx+skippy) < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if (iNdEx + skippy) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				if !options.DiscardUnknown {
					x.unknownFields = append(x.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
				}
				iNdEx += skippy
			}
		}

		if iNdEx > l {
			return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
		}
		return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, nil
	}
	return &protoiface.Methods{
		NoUnkeyedLiterals: struct{}{},
		Flags:             protoiface.SupportMarshalDeterministic | protoiface.SupportUnmarshalDiscardUnknown,
		Size:              size,
		Marshal:           marshal,
		Unmarshal:         unmarshal,
		Merge:             nil,
		CheckInitialized:  nil,
	}
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.0
// 	protoc        (unknown)
// source: miniwasm/tokenfactory/v1/authority_metadata.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// DenomAuthorityMetadata specifies metadata for addresses that have specific
// capabilities over a token factory denom. Right now there is only one Admin
// permission, but is planned to be extended to the future.
type DenomAuthorityMetadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Can be empty for no admin, or a valid address
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
}

func (x *DenomAuthorityMetadata) Reset() {
	*x = DenomAuthorityMetadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_miniwasm_tokenfactory_v1_authority_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DenomAuthorityMetadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DenomAuthorityMetadata) ProtoMessage() {}

// Deprecated: Use DenomAuthorityMetadata.ProtoReflect.Descriptor instead.
func (*DenomAuthorityMetadata) Descriptor() ([]byte, []int) {
	return file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *DenomAuthorityMetadata) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

var File_miniwasm_tokenfactory_v1_authority_metadata_proto protoreflect.FileDescriptor

var file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDesc = []byte{
	0x0a, 0x31, 0x6d, 0x69, 0x6e, 0x69, 0x77, 0x61, 0x73, 0x6d, 0x2f, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x69, 0x74, 0x79, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x18, 0x6d, 0x69, 0x6e, 0x69, 0x77, 0x61, 0x73, 0x6d, 0x2e, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x1a, 0x14, 0x67,
	0x6f, 0x67, 0x6f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x67, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x46, 0x0a, 0x16, 0x44, 0x65, 0x6e, 0x6f, 0x6d, 0x41, 0x75, 0x74, 0x68,
	0x6f, 0x72, 0x69, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x26, 0x0a,
	0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x10, 0xf2, 0xde,
	0x1f, 0x0c, 0x79, 0x61, 0x6d, 0x6c, 0x3a, 0x22, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x22, 0x52, 0x05,
	0x61, 0x64, 0x6d, 0x69, 0x6e, 0x3a, 0x04, 0xe8, 0xa0, 0x1f, 0x01, 0x42, 0x85, 0x02, 0x0a, 0x1c,
	0x63, 0x6f, 0x6d, 0x2e, 0x6d, 0x69, 0x6e, 0x69, 0x77, 0x61, 0x73, 0x6d, 0x2e, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x76, 0x31, 0x42, 0x16, 0x41, 0x75,
	0x74, 0x68, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x2d, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x6d,
	0x69, 0x6e, 0x69, 0x77, 0x61, 0x73, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x69, 0x6e, 0x69,
	0x77, 0x61, 0x73, 0x6d, 0x2f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x79, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x79, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x4d, 0x54, 0x58, 0xaa, 0x02, 0x18, 0x4d, 0x69, 0x6e, 0x69,
	0x77, 0x61, 0x73, 0x6d, 0x2e, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72,
	0x79, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x18, 0x4d, 0x69, 0x6e, 0x69, 0x77, 0x61, 0x73, 0x6d, 0x5c,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x5c, 0x56, 0x31, 0xe2,
	0x02, 0x24, 0x4d, 0x69, 0x6e, 0x69, 0x77, 0x61, 0x73, 0x6d, 0x5c, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1a, 0x4d, 0x69, 0x6e, 0x69, 0x77, 0x61, 0x73,
	0x6d, 0x3a, 0x3a, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x66, 0x61, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x3a,
	0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescOnce sync.Once
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescData = file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDesc
)

func file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescGZIP() []byte {
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescOnce.Do(func() {
		file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescData)
	})
	return file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDescData
}

var file_miniwasm_tokenfactory_v1_authority_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_miniwasm_tokenfactory_v1_authority_metadata_proto_goTypes = []interface{}{
	(*DenomAuthorityMetadata)(nil), // 0: miniwasm.tokenfactory.v1.DenomAuthorityMetadata
}
var file_miniwasm_tokenfactory_v1_authority_metadata_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_miniwasm_tokenfactory_v1_authority_metadata_proto_init() }
func file_miniwasm_tokenfactory_v1_authority_metadata_proto_init() {
	if File_miniwasm_tokenfactory_v1_authority_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_miniwasm_tokenfactory_v1_authority_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DenomAuthorityMetadata); i {
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
			RawDescriptor: file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_miniwasm_tokenfactory_v1_authority_metadata_proto_goTypes,
		DependencyIndexes: file_miniwasm_tokenfactory_v1_authority_metadata_proto_depIdxs,
		MessageInfos:      file_miniwasm_tokenfactory_v1_authority_metadata_proto_msgTypes,
	}.Build()
	File_miniwasm_tokenfactory_v1_authority_metadata_proto = out.File
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_rawDesc = nil
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_goTypes = nil
	file_miniwasm_tokenfactory_v1_authority_metadata_proto_depIdxs = nil
}
