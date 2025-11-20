package server

import (
	"context"
	"github.com/nxtvrtur/tages-file-service/internal/limiter"
	"github.com/nxtvrtur/tages-file-service/internal/storage"
	pb "github.com/nxtvrtur/tages-file-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"path/filepath"
)

type Server struct {
	pb.UnimplementedFileServiceServer

	transferLimiter *limiter.Limiter // Upload & Download — max 10
	listLimiter     *limiter.Limiter // ListFiles — max 100
}

func New() *Server {
	return &Server{
		transferLimiter: limiter.New(10),
		listLimiter:     limiter.New(100),
	}
}

func (s *Server) Upload(stream pb.FileService_UploadServer) (uploadErr error) {
	if err := s.transferLimiter.Acquire(stream.Context()); err != nil {
		return status.Error(codes.ResourceExhausted, "too many concurrent uploads")
	}
	defer s.transferLimiter.Release()

	var filename string
	var gotInfo bool
	var hasData bool
	var file *os.File
	var filePath string

	defer func() {
		if file != nil {
			_ = file.Close()
			if uploadErr != nil && filePath != "" {
				_ = os.Remove(filePath)
			}
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		if info := req.GetInfo(); info != nil {
			if gotInfo {
				return status.Error(codes.InvalidArgument, "file info already received")
			}
			cleanName := filepath.Base(info.Filename)
			if cleanName == "" {
				return status.Error(codes.InvalidArgument, "filename is required")
			}
			if cleanName != info.Filename {
				return status.Error(codes.InvalidArgument, "filename must not contain path")
			}
			filename = cleanName
			filePath = storage.FilePath(filename)
			file, err = os.Create(filePath)
			if err != nil {
				return status.Errorf(codes.Internal, "failed to create file: %v", err)
			}
			gotInfo = true
			continue
		}

		if chunk := req.GetChunk(); chunk != nil {
			if !gotInfo {
				return status.Error(codes.InvalidArgument, "file info must be sent first")
			}
			if len(chunk) == 0 {
				continue
			}
			if _, err := file.Write(chunk); err != nil {
				return status.Errorf(codes.Internal, "failed to write chunk: %v", err)
			}
			hasData = true
		}
	}

	if !gotInfo {
		return status.Error(codes.InvalidArgument, "file info not received")
	}

	if !hasData {
		return status.Error(codes.InvalidArgument, "uploaded file is empty")
	}

	return stream.SendAndClose(
		&pb.UploadResponse{
			Message: "file uploaded successfully: " + filename,
		},
	)
}

func (s *Server) ListFiles(ctx context.Context, _ *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	if err := s.listLimiter.Acquire(ctx); err != nil {
		return nil, status.Error(codes.ResourceExhausted, "too many concurrent list requests")
	}
	defer s.listLimiter.Release()

	files, err := storage.ListFiles()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read files: %v", err)
	}

	if files == nil {
		files = make([]*pb.FileMetadata, 0)
	}

	return &pb.ListFilesResponse{Files: files}, nil
}

func (s *Server) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	if err := s.transferLimiter.Acquire(stream.Context()); err != nil {
		return status.Error(codes.ResourceExhausted, "too many concurrent downloads requests")
	}
	defer s.transferLimiter.Release()

	path := storage.FilePath(req.Filename)
	file, err := os.Open(path)
	if err != nil {
		return status.Errorf(codes.NotFound, "file %s not found", req.Filename)
	}
	defer file.Close()

	buf := make([]byte, 64*1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to read file: %v", err)
		}
		if err := stream.Send(&pb.DownloadResponse{Chunk: buf[:n]}); err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
		}
	}
	return nil
}
