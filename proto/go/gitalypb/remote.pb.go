// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.21.7
// source: remote.proto

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
type UpdateRemoteMirrorRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repository is the repository whose mirror repository to update.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// Remote contains parameters required to connect to the remote repository.
	// This allows Gitaly to use an in-memory remote and does not require any
	// on-disk remote configuration.
	Remote *UpdateRemoteMirrorRequest_Remote `protobuf:"bytes,7,opt,name=remote,proto3" json:"remote,omitempty"`
	// OnlyBranchesMatching contains patterns to match branches against. Only the
	// matched brances are updated in the remote mirror. If no patterns are
	// specified, all branches are updated. The patterns should only contain the
	// branch name without the 'refs/heads/' prefix. "*" can be used as a
	// wildcard to match anything. only_branches_matching can be streamed to the
	// server over multiple messages. Optional.
	OnlyBranchesMatching [][]byte `protobuf:"bytes,3,rep,name=only_branches_matching,json=onlyBranchesMatching,proto3" json:"only_branches_matching,omitempty"` // protolint:disable:this REPEATED_FIELD_NAMES_PLURALIZED
	// SshKey is the SSH key to use for accessing to the mirror repository.
	// Optional.
	SshKey string `protobuf:"bytes,4,opt,name=ssh_key,json=sshKey,proto3" json:"ssh_key,omitempty"`
	// KnownHosts specifies the identities used for strict host key checking.
	// Optional.
	KnownHosts string `protobuf:"bytes,5,opt,name=known_hosts,json=knownHosts,proto3" json:"known_hosts,omitempty"`
	// KeepDivergentRefs specifies whether or not to update diverged references
	// in the mirror repository.
	KeepDivergentRefs bool `protobuf:"varint,6,opt,name=keep_divergent_refs,json=keepDivergentRefs,proto3" json:"keep_divergent_refs,omitempty"`
}

func (x *UpdateRemoteMirrorRequest) Reset() {
	*x = UpdateRemoteMirrorRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRemoteMirrorRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRemoteMirrorRequest) ProtoMessage() {}

func (x *UpdateRemoteMirrorRequest) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRemoteMirrorRequest.ProtoReflect.Descriptor instead.
func (*UpdateRemoteMirrorRequest) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{0}
}

func (x *UpdateRemoteMirrorRequest) GetRepository() *Repository {
	if x != nil {
		return x.Repository
	}
	return nil
}

func (x *UpdateRemoteMirrorRequest) GetRemote() *UpdateRemoteMirrorRequest_Remote {
	if x != nil {
		return x.Remote
	}
	return nil
}

func (x *UpdateRemoteMirrorRequest) GetOnlyBranchesMatching() [][]byte {
	if x != nil {
		return x.OnlyBranchesMatching
	}
	return nil
}

func (x *UpdateRemoteMirrorRequest) GetSshKey() string {
	if x != nil {
		return x.SshKey
	}
	return ""
}

func (x *UpdateRemoteMirrorRequest) GetKnownHosts() string {
	if x != nil {
		return x.KnownHosts
	}
	return ""
}

func (x *UpdateRemoteMirrorRequest) GetKeepDivergentRefs() bool {
	if x != nil {
		return x.KeepDivergentRefs
	}
	return false
}

// This comment is left unintentionally blank.
type UpdateRemoteMirrorResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// DivergentRefs contains a list of references that had diverged in the
	// mirror from the source repository.
	DivergentRefs [][]byte `protobuf:"bytes,1,rep,name=divergent_refs,json=divergentRefs,proto3" json:"divergent_refs,omitempty"`
}

func (x *UpdateRemoteMirrorResponse) Reset() {
	*x = UpdateRemoteMirrorResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRemoteMirrorResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRemoteMirrorResponse) ProtoMessage() {}

func (x *UpdateRemoteMirrorResponse) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRemoteMirrorResponse.ProtoReflect.Descriptor instead.
func (*UpdateRemoteMirrorResponse) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateRemoteMirrorResponse) GetDivergentRefs() [][]byte {
	if x != nil {
		return x.DivergentRefs
	}
	return nil
}

// This comment is left unintentionally blank.
type FindRemoteRepositoryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	Remote string `protobuf:"bytes,1,opt,name=remote,proto3" json:"remote,omitempty"`
	// This field is used to redirect request to proper storage where it can be handled.
	// As of now it doesn't matter what storage will be used, but it still must be a valid.
	// For more details: https://gitlab.com/gitlab-org/gitaly/-/issues/2442
	StorageName string `protobuf:"bytes,2,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
}

func (x *FindRemoteRepositoryRequest) Reset() {
	*x = FindRemoteRepositoryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindRemoteRepositoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindRemoteRepositoryRequest) ProtoMessage() {}

func (x *FindRemoteRepositoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindRemoteRepositoryRequest.ProtoReflect.Descriptor instead.
func (*FindRemoteRepositoryRequest) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{2}
}

func (x *FindRemoteRepositoryRequest) GetRemote() string {
	if x != nil {
		return x.Remote
	}
	return ""
}

func (x *FindRemoteRepositoryRequest) GetStorageName() string {
	if x != nil {
		return x.StorageName
	}
	return ""
}

// This migth throw a GRPC Unavailable code, to signal the request failure
// is transient.
type FindRemoteRepositoryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// This comment is left unintentionally blank.
	Exists bool `protobuf:"varint,1,opt,name=exists,proto3" json:"exists,omitempty"`
}

func (x *FindRemoteRepositoryResponse) Reset() {
	*x = FindRemoteRepositoryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindRemoteRepositoryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindRemoteRepositoryResponse) ProtoMessage() {}

func (x *FindRemoteRepositoryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindRemoteRepositoryResponse.ProtoReflect.Descriptor instead.
func (*FindRemoteRepositoryResponse) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{3}
}

func (x *FindRemoteRepositoryResponse) GetExists() bool {
	if x != nil {
		return x.Exists
	}
	return false
}

// FindRemoteRootRefRequest represents a request for the FindRemoteRootRef RPC.
type FindRemoteRootRefRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Repository is the repository in which the request shall be executed in. If
	// a remote name is given, then this is the repository in which the remote
	// will be looked up.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// RemoteUrl specifies the remote repository URL which should be fetched from.
	RemoteUrl string `protobuf:"bytes,3,opt,name=remote_url,json=remoteUrl,proto3" json:"remote_url,omitempty"`
	// HttpAuthorizationHeader is the HTTP header which should be added to the
	// request in order to authenticate against the repository.
	HttpAuthorizationHeader string `protobuf:"bytes,4,opt,name=http_authorization_header,json=httpAuthorizationHeader,proto3" json:"http_authorization_header,omitempty"`
	// ResolvedAddress holds the resolved IP address of the remote_url. This is
	// used to avoid DNS rebinding by mapping the url to the resolved address.
	// Only IPv4 dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"), or IPv4-mapped
	// IPv6 ("::ffff:192.0.2.1") forms are supported.
	// Works with HTTP/HTTPS/Git/SSH protocols.
	// Optional.
	ResolvedAddress string `protobuf:"bytes,6,opt,name=resolved_address,json=resolvedAddress,proto3" json:"resolved_address,omitempty"`
}

func (x *FindRemoteRootRefRequest) Reset() {
	*x = FindRemoteRootRefRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindRemoteRootRefRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindRemoteRootRefRequest) ProtoMessage() {}

func (x *FindRemoteRootRefRequest) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindRemoteRootRefRequest.ProtoReflect.Descriptor instead.
func (*FindRemoteRootRefRequest) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{4}
}

func (x *FindRemoteRootRefRequest) GetRepository() *Repository {
	if x != nil {
		return x.Repository
	}
	return nil
}

func (x *FindRemoteRootRefRequest) GetRemoteUrl() string {
	if x != nil {
		return x.RemoteUrl
	}
	return ""
}

func (x *FindRemoteRootRefRequest) GetHttpAuthorizationHeader() string {
	if x != nil {
		return x.HttpAuthorizationHeader
	}
	return ""
}

func (x *FindRemoteRootRefRequest) GetResolvedAddress() string {
	if x != nil {
		return x.ResolvedAddress
	}
	return ""
}

// FindRemoteRootRefResponse represents the response for the FindRemoteRootRef
// request.
type FindRemoteRootRefResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Ref is the name of the remote root reference.
	Ref string `protobuf:"bytes,1,opt,name=ref,proto3" json:"ref,omitempty"`
}

func (x *FindRemoteRootRefResponse) Reset() {
	*x = FindRemoteRootRefResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FindRemoteRootRefResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FindRemoteRootRefResponse) ProtoMessage() {}

func (x *FindRemoteRootRefResponse) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FindRemoteRootRefResponse.ProtoReflect.Descriptor instead.
func (*FindRemoteRootRefResponse) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{5}
}

func (x *FindRemoteRootRefResponse) GetRef() string {
	if x != nil {
		return x.Ref
	}
	return ""
}

// This comment is left unintentionally blank.
type UpdateRemoteMirrorRequest_Remote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// URL is the URL of the remote repository.
	Url string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	// HTTPAuthorizationHeader is an optional HTTP header used for
	// authenticating against the remote repository.
	HttpAuthorizationHeader string `protobuf:"bytes,2,opt,name=http_authorization_header,json=httpAuthorizationHeader,proto3" json:"http_authorization_header,omitempty"`
	// ResolvedAddress holds the resolved IP address of the remote_url. This is
	// used to avoid DNS rebinding by mapping the url to the resolved address.
	// Only IPv4 dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"), or IPv4-mapped
	// IPv6 ("::ffff:192.0.2.1") forms are supported.
	// Works with HTTP/HTTPS/Git/SSH protocols.
	// Optional.
	ResolvedAddress string `protobuf:"bytes,4,opt,name=resolved_address,json=resolvedAddress,proto3" json:"resolved_address,omitempty"`
}

func (x *UpdateRemoteMirrorRequest_Remote) Reset() {
	*x = UpdateRemoteMirrorRequest_Remote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_remote_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRemoteMirrorRequest_Remote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRemoteMirrorRequest_Remote) ProtoMessage() {}

func (x *UpdateRemoteMirrorRequest_Remote) ProtoReflect() protoreflect.Message {
	mi := &file_remote_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRemoteMirrorRequest_Remote.ProtoReflect.Descriptor instead.
func (*UpdateRemoteMirrorRequest_Remote) Descriptor() ([]byte, []int) {
	return file_remote_proto_rawDescGZIP(), []int{0, 0}
}

func (x *UpdateRemoteMirrorRequest_Remote) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *UpdateRemoteMirrorRequest_Remote) GetHttpAuthorizationHeader() string {
	if x != nil {
		return x.HttpAuthorizationHeader
	}
	return ""
}

func (x *UpdateRemoteMirrorRequest_Remote) GetResolvedAddress() string {
	if x != nil {
		return x.ResolvedAddress
	}
	return ""
}

var File_remote_proto protoreflect.FileDescriptor

var file_remote_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x1a, 0x0a, 0x6c, 0x69, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xdc, 0x03, 0x0a, 0x19, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x6d, 0x6f, 0x74,
	0x65, 0x4d, 0x69, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x38,
	0x0a, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x70, 0x6f,
	0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x42, 0x04, 0x98, 0xc6, 0x2c, 0x01, 0x52, 0x0a, 0x72, 0x65,
	0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x40, 0x0a, 0x06, 0x72, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c,
	0x79, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x69,
	0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x52, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x52, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x12, 0x34, 0x0a, 0x16, 0x6f, 0x6e,
	0x6c, 0x79, 0x5f, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x73, 0x5f, 0x6d, 0x61, 0x74, 0x63,
	0x68, 0x69, 0x6e, 0x67, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x14, 0x6f, 0x6e, 0x6c, 0x79,
	0x42, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x73, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67,
	0x12, 0x17, 0x0a, 0x07, 0x73, 0x73, 0x68, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x73, 0x68, 0x4b, 0x65, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x6b, 0x6e, 0x6f,
	0x77, 0x6e, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x48, 0x6f, 0x73, 0x74, 0x73, 0x12, 0x2e, 0x0a, 0x13, 0x6b, 0x65,
	0x65, 0x70, 0x5f, 0x64, 0x69, 0x76, 0x65, 0x72, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x66,
	0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x6b, 0x65, 0x65, 0x70, 0x44, 0x69, 0x76,
	0x65, 0x72, 0x67, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x66, 0x73, 0x1a, 0x92, 0x01, 0x0a, 0x06, 0x52,
	0x65, 0x6d, 0x6f, 0x74, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x3a, 0x0a, 0x19, 0x68, 0x74, 0x74, 0x70, 0x5f,
	0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x68, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17, 0x68, 0x74, 0x74, 0x70,
	0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x10, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x64, 0x5f,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x72,
	0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x64, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4a, 0x04,
	0x08, 0x03, 0x10, 0x04, 0x52, 0x09, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x4a,
	0x04, 0x08, 0x02, 0x10, 0x03, 0x52, 0x08, 0x72, 0x65, 0x66, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x22,
	0x43, 0x0a, 0x1a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d,
	0x69, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a,
	0x0e, 0x64, 0x69, 0x76, 0x65, 0x72, 0x67, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x66, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0d, 0x64, 0x69, 0x76, 0x65, 0x72, 0x67, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x66, 0x73, 0x22, 0x5e, 0x0a, 0x1b, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x12, 0x27, 0x0a, 0x0c, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x04, 0x88, 0xc6, 0x2c, 0x01, 0x52, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x22, 0x36, 0x0a, 0x1c, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x65, 0x78, 0x69, 0x73, 0x74, 0x73, 0x22, 0xf9, 0x01, 0x0a,
	0x18, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x52,
	0x65, 0x66, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x38, 0x0a, 0x0a, 0x72, 0x65, 0x70,
	0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72,
	0x79, 0x42, 0x04, 0x98, 0xc6, 0x2c, 0x01, 0x52, 0x0a, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x6f, 0x72, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x5f, 0x75, 0x72,
	0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x55,
	0x72, 0x6c, 0x12, 0x3a, 0x0a, 0x19, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17, 0x68, 0x74, 0x74, 0x70, 0x41, 0x75, 0x74, 0x68, 0x6f,
	0x72, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x29,
	0x0a, 0x10, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x64, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x72, 0x65, 0x73, 0x6f, 0x6c, 0x76,
	0x65, 0x64, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4a, 0x04, 0x08, 0x02, 0x10, 0x03, 0x4a,
	0x04, 0x08, 0x05, 0x10, 0x06, 0x52, 0x06, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x09, 0x68,
	0x74, 0x74, 0x70, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x22, 0x2d, 0x0a, 0x19, 0x46, 0x69, 0x6e, 0x64,
	0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x66, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x72, 0x65, 0x66, 0x32, 0xc5, 0x02, 0x0a, 0x0d, 0x52, 0x65, 0x6d, 0x6f,
	0x74, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x65, 0x0a, 0x12, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x69, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x21, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x69, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x22, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x4d, 0x69, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x06, 0xfa, 0x97, 0x28, 0x02, 0x08, 0x02, 0x28, 0x01,
	0x12, 0x6b, 0x0a, 0x14, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x65,
	0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x23, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c,
	0x79, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6f,
	0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e,
	0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x74,
	0x65, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x08, 0xfa, 0x97, 0x28, 0x04, 0x08, 0x02, 0x10, 0x02, 0x12, 0x60, 0x0a,
	0x11, 0x46, 0x69, 0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x52,
	0x65, 0x66, 0x12, 0x20, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x46, 0x69, 0x6e, 0x64,
	0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x66, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2e, 0x46, 0x69,
	0x6e, 0x64, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x52, 0x65, 0x66, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x06, 0xfa, 0x97, 0x28, 0x02, 0x08, 0x02, 0x42,
	0x34, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x69,
	0x74, 0x6c, 0x61, 0x62, 0x2d, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2f,
	0x76, 0x31, 0x35, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x67, 0x69, 0x74,
	0x61, 0x6c, 0x79, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_remote_proto_rawDescOnce sync.Once
	file_remote_proto_rawDescData = file_remote_proto_rawDesc
)

func file_remote_proto_rawDescGZIP() []byte {
	file_remote_proto_rawDescOnce.Do(func() {
		file_remote_proto_rawDescData = protoimpl.X.CompressGZIP(file_remote_proto_rawDescData)
	})
	return file_remote_proto_rawDescData
}

var file_remote_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_remote_proto_goTypes = []interface{}{
	(*UpdateRemoteMirrorRequest)(nil),        // 0: gitaly.UpdateRemoteMirrorRequest
	(*UpdateRemoteMirrorResponse)(nil),       // 1: gitaly.UpdateRemoteMirrorResponse
	(*FindRemoteRepositoryRequest)(nil),      // 2: gitaly.FindRemoteRepositoryRequest
	(*FindRemoteRepositoryResponse)(nil),     // 3: gitaly.FindRemoteRepositoryResponse
	(*FindRemoteRootRefRequest)(nil),         // 4: gitaly.FindRemoteRootRefRequest
	(*FindRemoteRootRefResponse)(nil),        // 5: gitaly.FindRemoteRootRefResponse
	(*UpdateRemoteMirrorRequest_Remote)(nil), // 6: gitaly.UpdateRemoteMirrorRequest.Remote
	(*Repository)(nil),                       // 7: gitaly.Repository
}
var file_remote_proto_depIdxs = []int32{
	7, // 0: gitaly.UpdateRemoteMirrorRequest.repository:type_name -> gitaly.Repository
	6, // 1: gitaly.UpdateRemoteMirrorRequest.remote:type_name -> gitaly.UpdateRemoteMirrorRequest.Remote
	7, // 2: gitaly.FindRemoteRootRefRequest.repository:type_name -> gitaly.Repository
	0, // 3: gitaly.RemoteService.UpdateRemoteMirror:input_type -> gitaly.UpdateRemoteMirrorRequest
	2, // 4: gitaly.RemoteService.FindRemoteRepository:input_type -> gitaly.FindRemoteRepositoryRequest
	4, // 5: gitaly.RemoteService.FindRemoteRootRef:input_type -> gitaly.FindRemoteRootRefRequest
	1, // 6: gitaly.RemoteService.UpdateRemoteMirror:output_type -> gitaly.UpdateRemoteMirrorResponse
	3, // 7: gitaly.RemoteService.FindRemoteRepository:output_type -> gitaly.FindRemoteRepositoryResponse
	5, // 8: gitaly.RemoteService.FindRemoteRootRef:output_type -> gitaly.FindRemoteRootRefResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_remote_proto_init() }
func file_remote_proto_init() {
	if File_remote_proto != nil {
		return
	}
	file_lint_proto_init()
	file_shared_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_remote_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRemoteMirrorRequest); i {
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
		file_remote_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRemoteMirrorResponse); i {
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
		file_remote_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindRemoteRepositoryRequest); i {
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
		file_remote_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindRemoteRepositoryResponse); i {
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
		file_remote_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindRemoteRootRefRequest); i {
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
		file_remote_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FindRemoteRootRefResponse); i {
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
		file_remote_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRemoteMirrorRequest_Remote); i {
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
			RawDescriptor: file_remote_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_remote_proto_goTypes,
		DependencyIndexes: file_remote_proto_depIdxs,
		MessageInfos:      file_remote_proto_msgTypes,
	}.Build()
	File_remote_proto = out.File
	file_remote_proto_rawDesc = nil
	file_remote_proto_goTypes = nil
	file_remote_proto_depIdxs = nil
}
