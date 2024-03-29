// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CapReadDirectoryClient is the client API for CapReadDirectory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CapReadDirectoryClient interface {
	// Get a list of TD documents the user is authorized for
	ListTD(ctx context.Context, in *ListTD_Args, opts ...grpc.CallOption) (*ThingValueList, error)
}

type capReadDirectoryClient struct {
	cc grpc.ClientConnInterface
}

func NewCapReadDirectoryClient(cc grpc.ClientConnInterface) CapReadDirectoryClient {
	return &capReadDirectoryClient{cc}
}

func (c *capReadDirectoryClient) ListTD(ctx context.Context, in *ListTD_Args, opts ...grpc.CallOption) (*ThingValueList, error) {
	out := new(ThingValueList)
	err := c.cc.Invoke(ctx, "/hiveot.grpc.CapReadDirectory/ListTD", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CapReadDirectoryServer is the server API for CapReadDirectory service.
// All implementations must embed UnimplementedCapReadDirectoryServer
// for forward compatibility
type CapReadDirectoryServer interface {
	// Get a list of TD documents the user is authorized for
	ListTD(context.Context, *ListTD_Args) (*ThingValueList, error)
	mustEmbedUnimplementedCapReadDirectoryServer()
}

// UnimplementedCapReadDirectoryServer must be embedded to have forward compatible implementations.
type UnimplementedCapReadDirectoryServer struct {
}

func (UnimplementedCapReadDirectoryServer) ListTD(context.Context, *ListTD_Args) (*ThingValueList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTD not implemented")
}
func (UnimplementedCapReadDirectoryServer) mustEmbedUnimplementedCapReadDirectoryServer() {}

// UnsafeCapReadDirectoryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CapReadDirectoryServer will
// result in compilation errors.
type UnsafeCapReadDirectoryServer interface {
	mustEmbedUnimplementedCapReadDirectoryServer()
}

func RegisterCapReadDirectoryServer(s grpc.ServiceRegistrar, srv CapReadDirectoryServer) {
	s.RegisterService(&CapReadDirectory_ServiceDesc, srv)
}

func _CapReadDirectory_ListTD_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListTD_Args)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CapReadDirectoryServer).ListTD(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/hiveot.grpc.CapReadDirectory/ListTD",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CapReadDirectoryServer).ListTD(ctx, req.(*ListTD_Args))
	}
	return interceptor(ctx, in, info, handler)
}

// CapReadDirectory_ServiceDesc is the grpc.ServiceDesc for CapReadDirectory service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CapReadDirectory_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "hiveot.grpc.CapReadDirectory",
	HandlerType: (*CapReadDirectoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListTD",
			Handler:    _CapReadDirectory_ListTD_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "directory.proto",
}
