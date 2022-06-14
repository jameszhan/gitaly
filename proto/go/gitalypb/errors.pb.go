// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.1
// source: errors.proto

package gitalypb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// HookType is the type of the hook that has been running. Please consult githooks(5) for more
// information about the specific types.
type CustomHookError_HookType int32

const (
	// HOOK_TYPE_UNSPECIFIED is the default hook type and should never be set.
	CustomHookError_HOOK_TYPE_UNSPECIFIED CustomHookError_HookType = 0
	// HOOK_TYPE_PRERECEIVE is executed after all changes have been written into a temporary staging
	// directory, but before any references in the repository have been updated. It is executed with
	// all references that are about to be updated at once. If this hook exits, then no references
	// will have been updated in the repository and staged objects will have been discarded.
	CustomHookError_HOOK_TYPE_PRERECEIVE CustomHookError_HookType = 1
	// HOOK_TYPE_UPDATE is executed after the pre-receive hook. It is executed per reference that is
	// about to be updated and can be used to reject only a subset of reference updates. If this
	// hook error is raised then a subset of references may have already been updated.
	CustomHookError_HOOK_TYPE_UPDATE CustomHookError_HookType = 2
	// HOOK_TYPE_POSTRECEIVE is executed after objects have been migrated into the repository and
	// after references have been updated. An error in this hook will not impact the changes
	// anymore as everything has already been persisted.
	CustomHookError_HOOK_TYPE_POSTRECEIVE CustomHookError_HookType = 3
)

// Enum value maps for CustomHookError_HookType.
var (
	CustomHookError_HookType_name = map[int32]string{
		0: "HOOK_TYPE_UNSPECIFIED",
		1: "HOOK_TYPE_PRERECEIVE",
		2: "HOOK_TYPE_UPDATE",
		3: "HOOK_TYPE_POSTRECEIVE",
	}
	CustomHookError_HookType_value = map[string]int32{
		"HOOK_TYPE_UNSPECIFIED": 0,
		"HOOK_TYPE_PRERECEIVE":  1,
		"HOOK_TYPE_UPDATE":      2,
		"HOOK_TYPE_POSTRECEIVE": 3,
	}
)

func (x CustomHookError_HookType) Enum() *CustomHookError_HookType {
	p := new(CustomHookError_HookType)
	*p = x
	return p
}

func (x CustomHookError_HookType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CustomHookError_HookType) Descriptor() protoreflect.EnumDescriptor {
	return file_errors_proto_enumTypes[0].Descriptor()
}

func (CustomHookError_HookType) Type() protoreflect.EnumType {
	return &file_errors_proto_enumTypes[0]
}

func (x CustomHookError_HookType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CustomHookError_HookType.Descriptor instead.
func (CustomHookError_HookType) EnumDescriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{9, 0}
}

// AccessCheckError is an error returned by GitLab's `/internal/allowed`
// endpoint.
type AccessCheckError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ErrorMessage is the error message as returned by the endpoint.
	ErrorMessage string `protobuf:"bytes,1,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	// Protocol is the protocol used.
	Protocol string `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`
	// UserId is the user ID as which changes had been pushed.
	UserId string `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	// Changes is the set of changes which have failed the access check.
	Changes []byte `protobuf:"bytes,4,opt,name=changes,proto3" json:"changes,omitempty"`
}

func (x *AccessCheckError) Reset() {
	*x = AccessCheckError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccessCheckError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccessCheckError) ProtoMessage() {}

func (x *AccessCheckError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccessCheckError.ProtoReflect.Descriptor instead.
func (*AccessCheckError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{0}
}

func (x *AccessCheckError) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *AccessCheckError) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *AccessCheckError) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *AccessCheckError) GetChanges() []byte {
	if x != nil {
		return x.Changes
	}
	return nil
}

// InvalidRefFormatError is an error returned when refs have an invalid format.
type InvalidRefFormatError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Refs are the offending refs with invalid formats.
	Refs [][]byte `protobuf:"bytes,2,rep,name=refs,proto3" json:"refs,omitempty"`
}

func (x *InvalidRefFormatError) Reset() {
	*x = InvalidRefFormatError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InvalidRefFormatError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InvalidRefFormatError) ProtoMessage() {}

func (x *InvalidRefFormatError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InvalidRefFormatError.ProtoReflect.Descriptor instead.
func (*InvalidRefFormatError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{1}
}

func (x *InvalidRefFormatError) GetRefs() [][]byte {
	if x != nil {
		return x.Refs
	}
	return nil
}

// NotAncestorError is an error returned when parent_revision is not an ancestor
// of the child_revision.
type NotAncestorError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ParentRevision is the revision checked against ChildRevision for whether it
	// is an ancestor of ChildRevision
	ParentRevision []byte `protobuf:"bytes,1,opt,name=parent_revision,json=parentRevision,proto3" json:"parent_revision,omitempty"`
	// ChildRevision is the revision checked against ParentRevision for whether
	// it is a descendent of ChildRevision.
	ChildRevision []byte `protobuf:"bytes,2,opt,name=child_revision,json=childRevision,proto3" json:"child_revision,omitempty"`
}

func (x *NotAncestorError) Reset() {
	*x = NotAncestorError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotAncestorError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotAncestorError) ProtoMessage() {}

func (x *NotAncestorError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotAncestorError.ProtoReflect.Descriptor instead.
func (*NotAncestorError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{2}
}

func (x *NotAncestorError) GetParentRevision() []byte {
	if x != nil {
		return x.ParentRevision
	}
	return nil
}

func (x *NotAncestorError) GetChildRevision() []byte {
	if x != nil {
		return x.ChildRevision
	}
	return nil
}

// ChangesAlreadyAppliedError is an error returned when the operation would
// have resulted in no changes because these changes have already been applied.
type ChangesAlreadyAppliedError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ChangesAlreadyAppliedError) Reset() {
	*x = ChangesAlreadyAppliedError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangesAlreadyAppliedError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangesAlreadyAppliedError) ProtoMessage() {}

func (x *ChangesAlreadyAppliedError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangesAlreadyAppliedError.ProtoReflect.Descriptor instead.
func (*ChangesAlreadyAppliedError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{3}
}

// MergeConflictError is an error returned in the case when merging two commits
// fails due to a merge conflict.
type MergeConflictError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ConflictingFiles is the set of files which have been conflicting. If this
	// field is empty, then there has still been a merge conflict, but it wasn't
	// able to determine which files have been conflicting.
	ConflictingFiles [][]byte `protobuf:"bytes,1,rep,name=conflicting_files,json=conflictingFiles,proto3" json:"conflicting_files,omitempty"`
	// ConflictingCommitIds is the set of commit IDs that caused the conflict. In the general case,
	// this should be set to two commit IDs.
	ConflictingCommitIds []string `protobuf:"bytes,2,rep,name=conflicting_commit_ids,json=conflictingCommitIds,proto3" json:"conflicting_commit_ids,omitempty"`
}

func (x *MergeConflictError) Reset() {
	*x = MergeConflictError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MergeConflictError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MergeConflictError) ProtoMessage() {}

func (x *MergeConflictError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MergeConflictError.ProtoReflect.Descriptor instead.
func (*MergeConflictError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{4}
}

func (x *MergeConflictError) GetConflictingFiles() [][]byte {
	if x != nil {
		return x.ConflictingFiles
	}
	return nil
}

func (x *MergeConflictError) GetConflictingCommitIds() []string {
	if x != nil {
		return x.ConflictingCommitIds
	}
	return nil
}

// ReferencesLockedError is an error returned when an ref update fails because
// the references have already been locked by another process.
type ReferencesLockedError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ReferencesLockedError) Reset() {
	*x = ReferencesLockedError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReferencesLockedError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReferencesLockedError) ProtoMessage() {}

func (x *ReferencesLockedError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReferencesLockedError.ProtoReflect.Descriptor instead.
func (*ReferencesLockedError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{5}
}

// ReferenceUpdateError is an error returned when updating a reference has
// failed.
type ReferenceUpdateError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ReferenceName is the name of the reference that failed to be updated.
	ReferenceName []byte `protobuf:"bytes,1,opt,name=reference_name,json=referenceName,proto3" json:"reference_name,omitempty"`
	// OldOid is the object ID the reference should have pointed to before the update.
	OldOid string `protobuf:"bytes,2,opt,name=old_oid,json=oldOid,proto3" json:"old_oid,omitempty"`
	// NewOid is the object ID the reference should have pointed to after the update.
	NewOid string `protobuf:"bytes,3,opt,name=new_oid,json=newOid,proto3" json:"new_oid,omitempty"`
}

func (x *ReferenceUpdateError) Reset() {
	*x = ReferenceUpdateError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReferenceUpdateError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReferenceUpdateError) ProtoMessage() {}

func (x *ReferenceUpdateError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReferenceUpdateError.ProtoReflect.Descriptor instead.
func (*ReferenceUpdateError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{6}
}

func (x *ReferenceUpdateError) GetReferenceName() []byte {
	if x != nil {
		return x.ReferenceName
	}
	return nil
}

func (x *ReferenceUpdateError) GetOldOid() string {
	if x != nil {
		return x.OldOid
	}
	return ""
}

func (x *ReferenceUpdateError) GetNewOid() string {
	if x != nil {
		return x.NewOid
	}
	return ""
}

// ResolveRevisionError is an error returned when resolving a specific revision
// has failed.
type ResolveRevisionError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Revision is the name of the revision that was tried to be resolved.
	Revision []byte `protobuf:"bytes,1,opt,name=revision,proto3" json:"revision,omitempty"`
}

func (x *ResolveRevisionError) Reset() {
	*x = ResolveRevisionError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResolveRevisionError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResolveRevisionError) ProtoMessage() {}

func (x *ResolveRevisionError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResolveRevisionError.ProtoReflect.Descriptor instead.
func (*ResolveRevisionError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{7}
}

func (x *ResolveRevisionError) GetRevision() []byte {
	if x != nil {
		return x.Revision
	}
	return nil
}

// LimitError is an error returned when Gitaly enforces request limits.
type LimitError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ErrorMessage provides context into why a limit was enforced.
	ErrorMessage string `protobuf:"bytes,1,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	// RetryAfter provides the duration after which a retry is safe.
	// 0 indicates non-retryable.
	RetryAfter *durationpb.Duration `protobuf:"bytes,2,opt,name=retry_after,json=retryAfter,proto3" json:"retry_after,omitempty"`
}

func (x *LimitError) Reset() {
	*x = LimitError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LimitError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LimitError) ProtoMessage() {}

func (x *LimitError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LimitError.ProtoReflect.Descriptor instead.
func (*LimitError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{8}
}

func (x *LimitError) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

func (x *LimitError) GetRetryAfter() *durationpb.Duration {
	if x != nil {
		return x.RetryAfter
	}
	return nil
}

// CustomHookError is an error returned when Gitaly executes a custom hook and the hook returns
// a non-zero return code.
type CustomHookError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Stdout is the standard output of the hook that has failed, if any. Data may be truncated.
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// Stderr is the standard error of the hook that has failed, if any. Data may be truncated.
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// HookType is the type of the hook.
	HookType CustomHookError_HookType `protobuf:"varint,3,opt,name=hook_type,json=hookType,proto3,enum=gitaly.CustomHookError_HookType" json:"hook_type,omitempty"`
}

func (x *CustomHookError) Reset() {
	*x = CustomHookError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_errors_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CustomHookError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CustomHookError) ProtoMessage() {}

func (x *CustomHookError) ProtoReflect() protoreflect.Message {
	mi := &file_errors_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CustomHookError.ProtoReflect.Descriptor instead.
func (*CustomHookError) Descriptor() ([]byte, []int) {
	return file_errors_proto_rawDescGZIP(), []int{9}
}

func (x *CustomHookError) GetStdout() []byte {
	if x != nil {
		return x.Stdout
	}
	return nil
}

func (x *CustomHookError) GetStderr() []byte {
	if x != nil {
		return x.Stderr
	}
	return nil
}

func (x *CustomHookError) GetHookType() CustomHookError_HookType {
	if x != nil {
		return x.HookType
	}
	return CustomHookError_HOOK_TYPE_UNSPECIFIED
}

var File_errors_proto protoreflect.FileDescriptor

var file_errors_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x86, 0x01, 0x0a, 0x10, 0x41, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x17, 0x0a, 0x07,
	0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75,
	0x73, 0x65, 0x72, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x22,
	0x2b, 0x0a, 0x15, 0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x52, 0x65, 0x66, 0x46, 0x6f, 0x72,
	0x6d, 0x61, 0x74, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x65, 0x66, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x72, 0x65, 0x66, 0x73, 0x22, 0x62, 0x0a, 0x10,
	0x4e, 0x6f, 0x74, 0x41, 0x6e, 0x63, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x12, 0x27, 0x0a, 0x0f, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x72, 0x65, 0x76, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x70, 0x61, 0x72, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x63, 0x68, 0x69,
	0x6c, 0x64, 0x5f, 0x72, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0d, 0x63, 0x68, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x22, 0x1c, 0x0a, 0x1a, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x41, 0x6c, 0x72, 0x65, 0x61,
	0x64, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x65, 0x64, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x77,
	0x0a, 0x12, 0x4d, 0x65, 0x72, 0x67, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x6c, 0x69, 0x63, 0x74, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x12, 0x2b, 0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x66, 0x6c, 0x69, 0x63, 0x74,
	0x69, 0x6e, 0x67, 0x5f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52,
	0x10, 0x63, 0x6f, 0x6e, 0x66, 0x6c, 0x69, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x46, 0x69, 0x6c, 0x65,
	0x73, 0x12, 0x34, 0x0a, 0x16, 0x63, 0x6f, 0x6e, 0x66, 0x6c, 0x69, 0x63, 0x74, 0x69, 0x6e, 0x67,
	0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x14, 0x63, 0x6f, 0x6e, 0x66, 0x6c, 0x69, 0x63, 0x74, 0x69, 0x6e, 0x67, 0x43, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x49, 0x64, 0x73, 0x22, 0x17, 0x0a, 0x15, 0x52, 0x65, 0x66, 0x65, 0x72,
	0x65, 0x6e, 0x63, 0x65, 0x73, 0x4c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x22, 0x6f, 0x0a, 0x14, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x66, 0x65,
	0x72, 0x65, 0x6e, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0d, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x17, 0x0a, 0x07, 0x6f, 0x6c, 0x64, 0x5f, 0x6f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6f, 0x6c, 0x64, 0x4f, 0x69, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x6e, 0x65, 0x77, 0x5f,
	0x6f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x65, 0x77, 0x4f, 0x69,
	0x64, 0x22, 0x32, 0x0a, 0x14, 0x52, 0x65, 0x73, 0x6f, 0x6c, 0x76, 0x65, 0x52, 0x65, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x72, 0x65, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x6d, 0x0a, 0x0a, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x3a, 0x0a, 0x0b, 0x72, 0x65, 0x74, 0x72,
	0x79, 0x5f, 0x61, 0x66, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x72, 0x65, 0x74, 0x72, 0x79, 0x41,
	0x66, 0x74, 0x65, 0x72, 0x22, 0xf2, 0x01, 0x0a, 0x0f, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x48,
	0x6f, 0x6f, 0x6b, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x64, 0x6f,
	0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x73, 0x74, 0x64, 0x65, 0x72, 0x72, 0x12, 0x3d, 0x0a, 0x09, 0x68, 0x6f, 0x6f, 0x6b,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x67, 0x69,
	0x74, 0x61, 0x6c, 0x79, 0x2e, 0x43, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x48, 0x6f, 0x6f, 0x6b, 0x45,
	0x72, 0x72, 0x6f, 0x72, 0x2e, 0x48, 0x6f, 0x6f, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x52, 0x08, 0x68,
	0x6f, 0x6f, 0x6b, 0x54, 0x79, 0x70, 0x65, 0x22, 0x70, 0x0a, 0x08, 0x48, 0x6f, 0x6f, 0x6b, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x19, 0x0a, 0x15, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x54, 0x59, 0x50, 0x45,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x18,
	0x0a, 0x14, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x52, 0x45, 0x52,
	0x45, 0x43, 0x45, 0x49, 0x56, 0x45, 0x10, 0x01, 0x12, 0x14, 0x0a, 0x10, 0x48, 0x4f, 0x4f, 0x4b,
	0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x50, 0x44, 0x41, 0x54, 0x45, 0x10, 0x02, 0x12, 0x19,
	0x0a, 0x15, 0x48, 0x4f, 0x4f, 0x4b, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x4f, 0x53, 0x54,
	0x52, 0x45, 0x43, 0x45, 0x49, 0x56, 0x45, 0x10, 0x03, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69, 0x74,
	0x6c, 0x61, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x69, 0x74, 0x6c, 0x61, 0x62, 0x2d, 0x6f,
	0x72, 0x67, 0x2f, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x2f, 0x76, 0x31, 0x35, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x67, 0x69, 0x74, 0x61, 0x6c, 0x79, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_errors_proto_rawDescOnce sync.Once
	file_errors_proto_rawDescData = file_errors_proto_rawDesc
)

func file_errors_proto_rawDescGZIP() []byte {
	file_errors_proto_rawDescOnce.Do(func() {
		file_errors_proto_rawDescData = protoimpl.X.CompressGZIP(file_errors_proto_rawDescData)
	})
	return file_errors_proto_rawDescData
}

var file_errors_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_errors_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_errors_proto_goTypes = []interface{}{
	(CustomHookError_HookType)(0),      // 0: gitaly.CustomHookError.HookType
	(*AccessCheckError)(nil),           // 1: gitaly.AccessCheckError
	(*InvalidRefFormatError)(nil),      // 2: gitaly.InvalidRefFormatError
	(*NotAncestorError)(nil),           // 3: gitaly.NotAncestorError
	(*ChangesAlreadyAppliedError)(nil), // 4: gitaly.ChangesAlreadyAppliedError
	(*MergeConflictError)(nil),         // 5: gitaly.MergeConflictError
	(*ReferencesLockedError)(nil),      // 6: gitaly.ReferencesLockedError
	(*ReferenceUpdateError)(nil),       // 7: gitaly.ReferenceUpdateError
	(*ResolveRevisionError)(nil),       // 8: gitaly.ResolveRevisionError
	(*LimitError)(nil),                 // 9: gitaly.LimitError
	(*CustomHookError)(nil),            // 10: gitaly.CustomHookError
	(*durationpb.Duration)(nil),        // 11: google.protobuf.Duration
}
var file_errors_proto_depIdxs = []int32{
	11, // 0: gitaly.LimitError.retry_after:type_name -> google.protobuf.Duration
	0,  // 1: gitaly.CustomHookError.hook_type:type_name -> gitaly.CustomHookError.HookType
	2,  // [2:2] is the sub-list for method output_type
	2,  // [2:2] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_errors_proto_init() }
func file_errors_proto_init() {
	if File_errors_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_errors_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccessCheckError); i {
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
		file_errors_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InvalidRefFormatError); i {
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
		file_errors_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotAncestorError); i {
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
		file_errors_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangesAlreadyAppliedError); i {
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
		file_errors_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MergeConflictError); i {
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
		file_errors_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReferencesLockedError); i {
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
		file_errors_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReferenceUpdateError); i {
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
		file_errors_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResolveRevisionError); i {
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
		file_errors_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LimitError); i {
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
		file_errors_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CustomHookError); i {
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
			RawDescriptor: file_errors_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_errors_proto_goTypes,
		DependencyIndexes: file_errors_proto_depIdxs,
		EnumInfos:         file_errors_proto_enumTypes,
		MessageInfos:      file_errors_proto_msgTypes,
	}.Build()
	File_errors_proto = out.File
	file_errors_proto_rawDesc = nil
	file_errors_proto_goTypes = nil
	file_errors_proto_depIdxs = nil
}
