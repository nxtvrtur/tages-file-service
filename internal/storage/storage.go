package storage

import (
	pb "github.com/nxtvrtur/tages-file-service/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"path/filepath"
)

const uploadDir = "uploads"

func init() {
	_ = os.MkdirAll(uploadDir, 0o755)
}

func SaveFile(filename string, data []byte) error {
	path := filepath.Join(uploadDir, filename)
	return os.WriteFile(path, data, 0644)
}

func FilePath(filename string) string {
	return filepath.Join(uploadDir, filename)
}

func ListFiles() ([]*pb.FileMetadata, error) {
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return nil, err
	}
	var result []*pb.FileMetadata
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(
			result, &pb.FileMetadata{
				Filename:  entry.Name(),
				CreatedAt: timestamppb.New(info.ModTime()),
				UpdatedAt: timestamppb.New(info.ModTime()),
			},
		)
	}
	return result, nil
}
