package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	pb "github.com/nxtvrtur/tages-file-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const chunkSize = 64 * 1024

func main() {
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	action := flag.String("action", "list", "action: upload | download | list")
	filePath := flag.String("file", "", "path to local file for upload/download destination")
	filename := flag.String("name", "", "remote filename")
	timeout := flag.Duration("timeout", 10*time.Second, "request timeout")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	switch *action {
	case "list":
		if err := listFiles(ctx, client); err != nil {
			log.Fatalf("list failed: %v", err)
		}
	case "upload":
		if *filePath == "" {
			log.Fatal("upload requires --file pointing to local path")
		}
		name := *filename
		if name == "" {
			name = filepath.Base(*filePath)
		}
		if err := uploadFile(ctx, client, *filePath, name); err != nil {
			log.Fatalf("upload failed: %v", err)
		}
	case "download":
		if *filename == "" {
			log.Fatal("download requires --name of remote file")
		}
		dest := *filePath
		if dest == "" {
			dest = *filename
		}
		if err := downloadFile(ctx, client, *filename, dest); err != nil {
			log.Fatalf("download failed: %v", err)
		}
	default:
		log.Fatalf("unknown action %q", *action)
	}
}

func uploadFile(ctx context.Context, client pb.FileServiceClient, localPath, remoteName string) error {
	stream, err := client.Upload(ctx)
	if err != nil {
		return err
	}

	if err := stream.Send(
		&pb.UploadRequest{
			Data: &pb.UploadRequest_Info{
				Info: &pb.FileInfo{Filename: remoteName},
			},
		},
	); err != nil {
		return err
	}

	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, chunkSize)
	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			if err := stream.Send(
				&pb.UploadRequest{
					Data: &pb.UploadRequest_Chunk{Chunk: buf[:n]},
				},
			); err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	fmt.Println(resp.GetMessage())
	return nil
}

func listFiles(ctx context.Context, client pb.FileServiceClient) error {
	resp, err := client.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		return err
	}
	if len(resp.Files) == 0 {
		fmt.Println("no files found")
		return nil
	}
	fmt.Println("Filename\tCreatedAt\tUpdatedAt")
	for _, f := range resp.Files {
		fmt.Printf(
			"%s\t%s\t%s\n", f.GetFilename(), f.GetCreatedAt().AsTime().Format(time.RFC3339),
			f.GetUpdatedAt().AsTime().Format(time.RFC3339),
		)
	}
	return nil
}

func downloadFile(ctx context.Context, client pb.FileServiceClient, remoteName, destPath string) error {
	stream, err := client.Download(ctx, &pb.DownloadRequest{Filename: remoteName})
	if err != nil {
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if _, err := out.Write(chunk.GetChunk()); err != nil {
			return err
		}
	}

	fmt.Printf("file %s downloaded to %s\n", remoteName, destPath)
	return nil
}
