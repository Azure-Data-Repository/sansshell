// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package localfile

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// LocalFileClient is the client API for LocalFile service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LocalFileClient interface {
	// Read reads a file from the disk and returns it contents.
	Read(ctx context.Context, in *ReadActionRequest, opts ...grpc.CallOption) (LocalFile_ReadClient, error)
	// Stat returns metadata about a filesytem path.
	Stat(ctx context.Context, opts ...grpc.CallOption) (LocalFile_StatClient, error)
	// Sum calculates a sum over the data in a single file.
	Sum(ctx context.Context, opts ...grpc.CallOption) (LocalFile_SumClient, error)
	// Write writes a file from the incoming RPC to a local file.
	Write(ctx context.Context, opts ...grpc.CallOption) (LocalFile_WriteClient, error)
	// Copy retrieves a file from the given blob URL and writes it to a local
	// file.
	Copy(ctx context.Context, in *CopyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// List returns StatReply entries for the entities contained at a given path.
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (LocalFile_ListClient, error)
	// SetFileAttributes takes a given filename and sets the given attributes.
	SetFileAttributes(ctx context.Context, in *SetFileAttributesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Rm removes the given file.
	Rm(ctx context.Context, in *RmRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Rmdir removes the given directory (must be empty).
	Rmdir(ctx context.Context, in *RmdirRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// Rename renames (moves) the given file to a new name. If the destination
	// is not a directory it will be replaced if it already exists.
	// OS restrictions may apply if the old and new names are outside of the same
	// directory.
	Rename(ctx context.Context, in *RenameRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type localFileClient struct {
	cc grpc.ClientConnInterface
}

func NewLocalFileClient(cc grpc.ClientConnInterface) LocalFileClient {
	return &localFileClient{cc}
}

func (c *localFileClient) Read(ctx context.Context, in *ReadActionRequest, opts ...grpc.CallOption) (LocalFile_ReadClient, error) {
	stream, err := c.cc.NewStream(ctx, &LocalFile_ServiceDesc.Streams[0], "/LocalFile.LocalFile/Read", opts...)
	if err != nil {
		return nil, err
	}
	x := &localFileReadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type LocalFile_ReadClient interface {
	Recv() (*ReadReply, error)
	grpc.ClientStream
}

type localFileReadClient struct {
	grpc.ClientStream
}

func (x *localFileReadClient) Recv() (*ReadReply, error) {
	m := new(ReadReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *localFileClient) Stat(ctx context.Context, opts ...grpc.CallOption) (LocalFile_StatClient, error) {
	stream, err := c.cc.NewStream(ctx, &LocalFile_ServiceDesc.Streams[1], "/LocalFile.LocalFile/Stat", opts...)
	if err != nil {
		return nil, err
	}
	x := &localFileStatClient{stream}
	return x, nil
}

type LocalFile_StatClient interface {
	Send(*StatRequest) error
	Recv() (*StatReply, error)
	grpc.ClientStream
}

type localFileStatClient struct {
	grpc.ClientStream
}

func (x *localFileStatClient) Send(m *StatRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *localFileStatClient) Recv() (*StatReply, error) {
	m := new(StatReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *localFileClient) Sum(ctx context.Context, opts ...grpc.CallOption) (LocalFile_SumClient, error) {
	stream, err := c.cc.NewStream(ctx, &LocalFile_ServiceDesc.Streams[2], "/LocalFile.LocalFile/Sum", opts...)
	if err != nil {
		return nil, err
	}
	x := &localFileSumClient{stream}
	return x, nil
}

type LocalFile_SumClient interface {
	Send(*SumRequest) error
	Recv() (*SumReply, error)
	grpc.ClientStream
}

type localFileSumClient struct {
	grpc.ClientStream
}

func (x *localFileSumClient) Send(m *SumRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *localFileSumClient) Recv() (*SumReply, error) {
	m := new(SumReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *localFileClient) Write(ctx context.Context, opts ...grpc.CallOption) (LocalFile_WriteClient, error) {
	stream, err := c.cc.NewStream(ctx, &LocalFile_ServiceDesc.Streams[3], "/LocalFile.LocalFile/Write", opts...)
	if err != nil {
		return nil, err
	}
	x := &localFileWriteClient{stream}
	return x, nil
}

type LocalFile_WriteClient interface {
	Send(*WriteRequest) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type localFileWriteClient struct {
	grpc.ClientStream
}

func (x *localFileWriteClient) Send(m *WriteRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *localFileWriteClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *localFileClient) Copy(ctx context.Context, in *CopyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/LocalFile.LocalFile/Copy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localFileClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (LocalFile_ListClient, error) {
	stream, err := c.cc.NewStream(ctx, &LocalFile_ServiceDesc.Streams[4], "/LocalFile.LocalFile/List", opts...)
	if err != nil {
		return nil, err
	}
	x := &localFileListClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type LocalFile_ListClient interface {
	Recv() (*ListReply, error)
	grpc.ClientStream
}

type localFileListClient struct {
	grpc.ClientStream
}

func (x *localFileListClient) Recv() (*ListReply, error) {
	m := new(ListReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *localFileClient) SetFileAttributes(ctx context.Context, in *SetFileAttributesRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/LocalFile.LocalFile/SetFileAttributes", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localFileClient) Rm(ctx context.Context, in *RmRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/LocalFile.LocalFile/Rm", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localFileClient) Rmdir(ctx context.Context, in *RmdirRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/LocalFile.LocalFile/Rmdir", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *localFileClient) Rename(ctx context.Context, in *RenameRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/LocalFile.LocalFile/Rename", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LocalFileServer is the server API for LocalFile service.
// All implementations should embed UnimplementedLocalFileServer
// for forward compatibility
type LocalFileServer interface {
	// Read reads a file from the disk and returns it contents.
	Read(*ReadActionRequest, LocalFile_ReadServer) error
	// Stat returns metadata about a filesytem path.
	Stat(LocalFile_StatServer) error
	// Sum calculates a sum over the data in a single file.
	Sum(LocalFile_SumServer) error
	// Write writes a file from the incoming RPC to a local file.
	Write(LocalFile_WriteServer) error
	// Copy retrieves a file from the given blob URL and writes it to a local
	// file.
	Copy(context.Context, *CopyRequest) (*emptypb.Empty, error)
	// List returns StatReply entries for the entities contained at a given path.
	List(*ListRequest, LocalFile_ListServer) error
	// SetFileAttributes takes a given filename and sets the given attributes.
	SetFileAttributes(context.Context, *SetFileAttributesRequest) (*emptypb.Empty, error)
	// Rm removes the given file.
	Rm(context.Context, *RmRequest) (*emptypb.Empty, error)
	// Rmdir removes the given directory (must be empty).
	Rmdir(context.Context, *RmdirRequest) (*emptypb.Empty, error)
	// Rename renames (moves) the given file to a new name. If the destination
	// is not a directory it will be replaced if it already exists.
	// OS restrictions may apply if the old and new names are outside of the same
	// directory.
	Rename(context.Context, *RenameRequest) (*emptypb.Empty, error)
}

// UnimplementedLocalFileServer should be embedded to have forward compatible implementations.
type UnimplementedLocalFileServer struct {
}

func (UnimplementedLocalFileServer) Read(*ReadActionRequest, LocalFile_ReadServer) error {
	return status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedLocalFileServer) Stat(LocalFile_StatServer) error {
	return status.Errorf(codes.Unimplemented, "method Stat not implemented")
}
func (UnimplementedLocalFileServer) Sum(LocalFile_SumServer) error {
	return status.Errorf(codes.Unimplemented, "method Sum not implemented")
}
func (UnimplementedLocalFileServer) Write(LocalFile_WriteServer) error {
	return status.Errorf(codes.Unimplemented, "method Write not implemented")
}
func (UnimplementedLocalFileServer) Copy(context.Context, *CopyRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Copy not implemented")
}
func (UnimplementedLocalFileServer) List(*ListRequest, LocalFile_ListServer) error {
	return status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedLocalFileServer) SetFileAttributes(context.Context, *SetFileAttributesRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetFileAttributes not implemented")
}
func (UnimplementedLocalFileServer) Rm(context.Context, *RmRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rm not implemented")
}
func (UnimplementedLocalFileServer) Rmdir(context.Context, *RmdirRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rmdir not implemented")
}
func (UnimplementedLocalFileServer) Rename(context.Context, *RenameRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rename not implemented")
}

// UnsafeLocalFileServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LocalFileServer will
// result in compilation errors.
type UnsafeLocalFileServer interface {
	mustEmbedUnimplementedLocalFileServer()
}

func RegisterLocalFileServer(s grpc.ServiceRegistrar, srv LocalFileServer) {
	s.RegisterService(&LocalFile_ServiceDesc, srv)
}

func _LocalFile_Read_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ReadActionRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LocalFileServer).Read(m, &localFileReadServer{stream})
}

type LocalFile_ReadServer interface {
	Send(*ReadReply) error
	grpc.ServerStream
}

type localFileReadServer struct {
	grpc.ServerStream
}

func (x *localFileReadServer) Send(m *ReadReply) error {
	return x.ServerStream.SendMsg(m)
}

func _LocalFile_Stat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LocalFileServer).Stat(&localFileStatServer{stream})
}

type LocalFile_StatServer interface {
	Send(*StatReply) error
	Recv() (*StatRequest, error)
	grpc.ServerStream
}

type localFileStatServer struct {
	grpc.ServerStream
}

func (x *localFileStatServer) Send(m *StatReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *localFileStatServer) Recv() (*StatRequest, error) {
	m := new(StatRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _LocalFile_Sum_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LocalFileServer).Sum(&localFileSumServer{stream})
}

type LocalFile_SumServer interface {
	Send(*SumReply) error
	Recv() (*SumRequest, error)
	grpc.ServerStream
}

type localFileSumServer struct {
	grpc.ServerStream
}

func (x *localFileSumServer) Send(m *SumReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *localFileSumServer) Recv() (*SumRequest, error) {
	m := new(SumRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _LocalFile_Write_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LocalFileServer).Write(&localFileWriteServer{stream})
}

type LocalFile_WriteServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*WriteRequest, error)
	grpc.ServerStream
}

type localFileWriteServer struct {
	grpc.ServerStream
}

func (x *localFileWriteServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *localFileWriteServer) Recv() (*WriteRequest, error) {
	m := new(WriteRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _LocalFile_Copy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CopyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalFileServer).Copy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/LocalFile.LocalFile/Copy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalFileServer).Copy(ctx, req.(*CopyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalFile_List_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LocalFileServer).List(m, &localFileListServer{stream})
}

type LocalFile_ListServer interface {
	Send(*ListReply) error
	grpc.ServerStream
}

type localFileListServer struct {
	grpc.ServerStream
}

func (x *localFileListServer) Send(m *ListReply) error {
	return x.ServerStream.SendMsg(m)
}

func _LocalFile_SetFileAttributes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetFileAttributesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalFileServer).SetFileAttributes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/LocalFile.LocalFile/SetFileAttributes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalFileServer).SetFileAttributes(ctx, req.(*SetFileAttributesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalFile_Rm_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RmRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalFileServer).Rm(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/LocalFile.LocalFile/Rm",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalFileServer).Rm(ctx, req.(*RmRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalFile_Rmdir_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RmdirRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalFileServer).Rmdir(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/LocalFile.LocalFile/Rmdir",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalFileServer).Rmdir(ctx, req.(*RmdirRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocalFile_Rename_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocalFileServer).Rename(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/LocalFile.LocalFile/Rename",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocalFileServer).Rename(ctx, req.(*RenameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LocalFile_ServiceDesc is the grpc.ServiceDesc for LocalFile service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LocalFile_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "LocalFile.LocalFile",
	HandlerType: (*LocalFileServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Copy",
			Handler:    _LocalFile_Copy_Handler,
		},
		{
			MethodName: "SetFileAttributes",
			Handler:    _LocalFile_SetFileAttributes_Handler,
		},
		{
			MethodName: "Rm",
			Handler:    _LocalFile_Rm_Handler,
		},
		{
			MethodName: "Rmdir",
			Handler:    _LocalFile_Rmdir_Handler,
		},
		{
			MethodName: "Rename",
			Handler:    _LocalFile_Rename_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Read",
			Handler:       _LocalFile_Read_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Stat",
			Handler:       _LocalFile_Stat_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Sum",
			Handler:       _LocalFile_Sum_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Write",
			Handler:       _LocalFile_Write_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "List",
			Handler:       _LocalFile_List_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "localfile.proto",
}
