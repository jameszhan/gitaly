// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.7
// source: namespace.proto

package gitalypb

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

// This comment is left unintentionally blank.
type AddNamespaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	StorageName string `protobuf:"bytes,1,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
	// This comment is left unintentionally blank.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *AddNamespaceRequest) Reset() {
	*x = AddNamespaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddNamespaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddNamespaceRequest) ProtoMessage() {}

func (x *AddNamespaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddNamespaceRequest.ProtoReflect.Descriptor instead.
func (*AddNamespaceRequest) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{0}
}

func (x *AddNamespaceRequest) GetStorageName() string {
	if x != nil {
		return x.StorageName
	}
	return ""
}

func (x *AddNamespaceRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// This comment is left unintentionally blank.
type RemoveNamespaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	StorageName string `protobuf:"bytes,1,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
	// This comment is left unintentionally blank.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *RemoveNamespaceRequest) Reset() {
	*x = RemoveNamespaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveNamespaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveNamespaceRequest) ProtoMessage() {}

func (x *RemoveNamespaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveNamespaceRequest.ProtoReflect.Descriptor instead.
func (*RemoveNamespaceRequest) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{1}
}

func (x *RemoveNamespaceRequest) GetStorageName() string {
	if x != nil {
		return x.StorageName
	}
	return ""
}

func (x *RemoveNamespaceRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// This comment is left unintentionally blank.
type RenameNamespaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	StorageName string `protobuf:"bytes,1,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
	// This comment is left unintentionally blank.
	From string `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	// This comment is left unintentionally blank.
	To string `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
}

func (x *RenameNamespaceRequest) Reset() {
	*x = RenameNamespaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RenameNamespaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RenameNamespaceRequest) ProtoMessage() {}

func (x *RenameNamespaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RenameNamespaceRequest.ProtoReflect.Descriptor instead.
func (*RenameNamespaceRequest) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{2}
}

func (x *RenameNamespaceRequest) GetStorageName() string {
	if x != nil {
		return x.StorageName
	}
	return ""
}

func (x *RenameNamespaceRequest) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *RenameNamespaceRequest) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

// This comment is left unintentionally blank.
type NamespaceExistsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	StorageName string `protobuf:"bytes,1,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
	// This comment is left unintentionally blank.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *NamespaceExistsRequest) Reset() {
	*x = NamespaceExistsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NamespaceExistsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NamespaceExistsRequest) ProtoMessage() {}

func (x *NamespaceExistsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NamespaceExistsRequest.ProtoReflect.Descriptor instead.
func (*NamespaceExistsRequest) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{3}
}

func (x *NamespaceExistsRequest) GetStorageName() string {
	if x != nil {
		return x.StorageName
	}
	return ""
}

func (x *NamespaceExistsRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

// This comment is left unintentionally blank.
type NamespaceExistsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	Exists bool `protobuf:"varint,1,opt,name=exists,proto3" json:"exists,omitempty"`
}

func (x *NamespaceExistsResponse) Reset() {
	*x = NamespaceExistsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NamespaceExistsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NamespaceExistsResponse) ProtoMessage() {}

func (x *NamespaceExistsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NamespaceExistsResponse.ProtoReflect.Descriptor instead.
func (*NamespaceExistsResponse) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{4}
}

func (x *NamespaceExistsResponse) GetExists() bool {
	if x != nil {
		return x.Exists
	}
	return false
}

// This comment is left unintentionally blank.
type AddNamespaceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AddNamespaceResponse) Reset() {
	*x = AddNamespaceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddNamespaceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddNamespaceResponse) ProtoMessage() {}

func (x *AddNamespaceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddNamespaceResponse.ProtoReflect.Descriptor instead.
func (*AddNamespaceResponse) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{5}
}

// This comment is left unintentionally blank.
type RemoveNamespaceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoveNamespaceResponse) Reset() {
	*x = RemoveNamespaceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveNamespaceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveNamespaceResponse) ProtoMessage() {}

func (x *RemoveNamespaceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveNamespaceResponse.ProtoReflect.Descriptor instead.
func (*RemoveNamespaceResponse) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{6}
}

// This comment is left unintentionally blank.
type RenameNamespaceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RenameNamespaceResponse) Reset() {
	*x = RenameNamespaceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_namespace_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RenameNamespaceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RenameNamespaceResponse) ProtoMessage() {}

func (x *RenameNamespaceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_namespace_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RenameNamespaceResponse.ProtoReflect.Descriptor instead.
func (*RenameNamespaceResponse) Descriptor() ([]byte, []int) {
	return file_namespace_proto_rawDescGZIP(), []int{7}
}

var File_namespace_proto protoreflect.FileDescriptor

var file_namespace_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x1a, 0x0a, 0x6c, 0x69, 0x6e, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x52, 0x0a, 0x13, 0x41, 0x64, 0x64, 0x4e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0c,
	0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x04, 0x88, 0xc6, 0x2c, 0x01, 0x52, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x55, 0x0a, 0x16, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0c, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0x88, 0xc6, 0x2c, 0x01, 0x52,
	0x0b, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x22, 0x65, 0x0a, 0x16, 0x52, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70,
	0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x0c, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x42, 0x04, 0x88, 0xc6, 0x2c, 0x01, 0x52, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x22, 0x55, 0x0a, 0x16, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x27, 0x0a, 0x0c, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0x88, 0xc6, 0x2c, 0x01, 0x52, 0x0b, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x31,
	0x0a, 0x17, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x45, 0x78, 0x69, 0x73, 0x74,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x69,
	0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78, 0x69, 0x73, 0x74,
	0x73, 0x22, 0x16, 0x0a, 0x14, 0x41, 0x64, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x19, 0x0a, 0x17, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x19, 0x0a, 0x17, 0x52, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32,
	0x81, 0x03, 0x0a, 0x10, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x53, 0x0a, 0x0c, 0x41, 0x64, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x12, 0x1b, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x41, 0x64,
	0x64, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x1c, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x41, 0x64, 0x64, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x08, 0xfa, 0x97, 0x28, 0x04, 0x08, 0x01, 0x10, 0x02, 0x12, 0x5c, 0x0a, 0x0f, 0x52, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x1e, 0x2e, 0x67,
	0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x67,
	0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x08, 0xfa,
	0x97, 0x28, 0x04, 0x08, 0x01, 0x10, 0x02, 0x12, 0x5c, 0x0a, 0x0f, 0x52, 0x65, 0x6e, 0x61, 0x6d,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x1e, 0x2e, 0x67, 0x69, 0x74,
	0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70,
	0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x67, 0x69, 0x74,
	0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x6e, 0x61, 0x6d, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70,
	0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x08, 0xfa, 0x97, 0x28,
	0x04, 0x08, 0x01, 0x10, 0x02, 0x12, 0x5c, 0x0a, 0x0f, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x45, 0x78, 0x69, 0x73, 0x74, 0x73, 0x12, 0x1e, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c,
	0x79, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x45, 0x78, 0x69, 0x73, 0x74,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c,
	0x79, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x45, 0x78, 0x69, 0x73, 0x74,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x08, 0xfa, 0x97, 0x28, 0x04, 0x08,
	0x02, 0x10, 0x02, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2d, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x69, 0x74,
	0x61, 0x6c, 0x79, 0x2f, 0x76, 0x31, 0x36, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f,
	0x2f, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_namespace_proto_rawDescOnce sync.Once
	file_namespace_proto_rawDescData = file_namespace_proto_rawDesc
)

func file_namespace_proto_rawDescGZIP() []byte {
	file_namespace_proto_rawDescOnce.Do(func() {
		file_namespace_proto_rawDescData = protoimpl.X.CompressGZIP(file_namespace_proto_rawDescData)
	})
	return file_namespace_proto_rawDescData
}

var file_namespace_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_namespace_proto_goTypes = []interface{}{
	(*AddNamespaceRequest)(nil),     // 0: gitaly.AddNamespaceRequest
	(*RemoveNamespaceRequest)(nil),  // 1: gitaly.RemoveNamespaceRequest
	(*RenameNamespaceRequest)(nil),  // 2: gitaly.RenameNamespaceRequest
	(*NamespaceExistsRequest)(nil),  // 3: gitaly.NamespaceExistsRequest
	(*NamespaceExistsResponse)(nil), // 4: gitaly.NamespaceExistsResponse
	(*AddNamespaceResponse)(nil),    // 5: gitaly.AddNamespaceResponse
	(*RemoveNamespaceResponse)(nil), // 6: gitaly.RemoveNamespaceResponse
	(*RenameNamespaceResponse)(nil), // 7: gitaly.RenameNamespaceResponse
}
var file_namespace_proto_depIdxs = []int32{
	0, // 0: gitaly.NamespaceService.AddNamespace:input_type -> gitaly.AddNamespaceRequest
	1, // 1: gitaly.NamespaceService.RemoveNamespace:input_type -> gitaly.RemoveNamespaceRequest
	2, // 2: gitaly.NamespaceService.RenameNamespace:input_type -> gitaly.RenameNamespaceRequest
	3, // 3: gitaly.NamespaceService.NamespaceExists:input_type -> gitaly.NamespaceExistsRequest
	5, // 4: gitaly.NamespaceService.AddNamespace:output_type -> gitaly.AddNamespaceResponse
	6, // 5: gitaly.NamespaceService.RemoveNamespace:output_type -> gitaly.RemoveNamespaceResponse
	7, // 6: gitaly.NamespaceService.RenameNamespace:output_type -> gitaly.RenameNamespaceResponse
	4, // 7: gitaly.NamespaceService.NamespaceExists:output_type -> gitaly.NamespaceExistsResponse
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_namespace_proto_init() }
func file_namespace_proto_init() {
	if File_namespace_proto != nil {
		return
	}
	file_lint_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_namespace_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddNamespaceRequest); i {
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
		file_namespace_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoveNamespaceRequest); i {
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
		file_namespace_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RenameNamespaceRequest); i {
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
		file_namespace_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NamespaceExistsRequest); i {
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
		file_namespace_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NamespaceExistsResponse); i {
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
		file_namespace_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddNamespaceResponse); i {
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
		file_namespace_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoveNamespaceResponse); i {
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
		file_namespace_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RenameNamespaceResponse); i {
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
			RawDescriptor: file_namespace_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_namespace_proto_goTypes,
		DependencyIndexes: file_namespace_proto_depIdxs,
		MessageInfos:      file_namespace_proto_msgTypes,
	}.Build()
	File_namespace_proto = out.File
	file_namespace_proto_rawDesc = nil
	file_namespace_proto_goTypes = nil
	file_namespace_proto_depIdxs = nil
}
