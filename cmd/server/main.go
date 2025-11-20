package main

import (
	"context"
	"github.com/nxtvrtur/tages-file-service/internal/server"
	pb "github.com/nxtvrtur/tages-file-service/proto"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(loggingUnaryInterceptor),
		grpc.StreamInterceptor(loggingStreamInterceptor),
	)

	pb.RegisterFileServiceServer(s, server.New())
	reflection.Register(s)

	log.Println("gRPC server is running on :50051")
	log.Println("To test: grpcurl -plaintext localhost:50051 list")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func loggingUnaryInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("Unary call: %s", info.FullMethod)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	return resp, err
}

func loggingStreamInterceptor(
	srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	log.Printf("Stream call: %s", info.FullMethod)
	err := handler(srv, ss)
	if err != nil {
		log.Printf("Stream error: %v", err)
	}
	return err
}
