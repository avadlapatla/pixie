package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PhotoPlugin is the implementation of the PhotoPlugin service
type PhotoPlugin struct {
	// Implement the UnimplementedPhotoPluginServer to satisfy the interface
	UnimplementedPhotoPluginServer
}

// Photo represents a photo in the system
type Photo struct {
	ID     string
	S3Key  string
	Mime   string
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query string
}

// SearchResult represents a search result
type SearchResult struct {
	IDs []string
}

// UnimplementedPhotoPluginServer must be embedded to have forward compatible implementations
type UnimplementedPhotoPluginServer struct{}

func (*UnimplementedPhotoPluginServer) ProcessPhoto(context.Context, *Photo) (*emptypb.Empty, error) {
	return nil, nil
}

func (*UnimplementedPhotoPluginServer) Search(context.Context, *SearchRequest) (*SearchResult, error) {
	return nil, nil
}

// ProcessPhoto processes a photo (noop implementation)
func (p *PhotoPlugin) ProcessPhoto(ctx context.Context, photo *Photo) (*emptypb.Empty, error) {
	log.Printf("Received ProcessPhoto request for photo ID: %s", photo.ID)
	return &emptypb.Empty{}, nil
}

// Search searches for photos (noop implementation)
func (p *PhotoPlugin) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	log.Printf("Received Search request with query: %s", req.Query)
	return &SearchResult{IDs: []string{}}, nil
}

func main() {
	// Parse command line flags
	port := flag.Int("port", 0, "The server port")
	flag.Parse()

	// If port is 0, let the OS choose a port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Print the port for the plugin loader to read
	fmt.Printf("PORT=%d\n", lis.Addr().(*net.TCPAddr).Port)

	// Create a gRPC server with middleware
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	// Register the health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register the plugin service
	RegisterPhotoPluginServer(s, &PhotoPlugin{})

	// Handle shutdown gracefully
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down gRPC server...")
		s.GracefulStop()
	}()

	// Start the server
	log.Printf("Starting noop plugin server on port %d", lis.Addr().(*net.TCPAddr).Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// RegisterPhotoPluginServer registers the PhotoPluginServer to the gRPC server
func RegisterPhotoPluginServer(s *grpc.Server, srv *PhotoPlugin) {
	s.RegisterService(&_PhotoPlugin_serviceDesc, srv)
}

// Service descriptor for the PhotoPlugin service
var _PhotoPlugin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "plugin.v1.PhotoPlugin",
	HandlerType: (*interface{})(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessPhoto",
			Handler:    _PhotoPlugin_ProcessPhoto_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _PhotoPlugin_Search_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "plugin/v1/plugin.proto",
}

// Handler for ProcessPhoto
func _PhotoPlugin_ProcessPhoto_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Photo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*PhotoPlugin).ProcessPhoto(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.PhotoPlugin/ProcessPhoto",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*PhotoPlugin).ProcessPhoto(ctx, req.(*Photo))
	}
	return interceptor(ctx, in, info, handler)
}

// Handler for Search
func _PhotoPlugin_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(*PhotoPlugin).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.v1.PhotoPlugin/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(*PhotoPlugin).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}
