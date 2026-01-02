package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"bharani/proto/frontend"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MaxBlockSize = 3 * 1024 * 1024

type FileStorage struct {
	client frontend.FrontendServiceClient
	ctx    context.Context
}

func NewFileStorage(addr string) (*FileStorage, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &FileStorage{
		client: frontend.NewFrontendServiceClient(conn),
		ctx:    context.Background(),
	}, nil
}

// PutFile uploads a file, splitting it into blocks if necessary
func (fs *FileStorage) PutFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fmt.Printf("Uploading file: %s (size: %d bytes)\n", filePath, fileInfo.Size())

	var blockHashes []string
	buffer := make([]byte, MaxBlockSize)
	blockNum := 0

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		if n == 0 {
			break
		}

		blockData := buffer[:n]
		blockNum++

		fmt.Printf("  Uploading block %d (%d bytes)...\n", blockNum, len(blockData))

		putResp, err := fs.client.Put(fs.ctx, &frontend.PutRequest{
			Data: blockData,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to put block %d: %w", blockNum, err)
		}

		if !putResp.Success {
			return nil, fmt.Errorf("put block %d failed: %s", blockNum, putResp.Error)
		}

		blockHashes = append(blockHashes, putResp.Hash)
		fmt.Printf("    Block %d stored with hash: %s\n", blockNum, putResp.Hash)
	}

	return blockHashes, nil
}

// GetFile downloads a file by its block hashes
func (fs *FileStorage) GetFile(blockHashes []string, outputPath string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	fmt.Printf("Downloading file to: %s (%d blocks)\n", outputPath, len(blockHashes))

	for i, hash := range blockHashes {
		fmt.Printf("  Downloading block %d/%d (hash: %s)...\n", i+1, len(blockHashes), hash)

		getResp, err := fs.client.Get(fs.ctx, &frontend.GetRequest{
			Hash: hash,
		})
		if err != nil {
			return fmt.Errorf("failed to get block %d: %w", i+1, err)
		}

		if !getResp.Success {
			return fmt.Errorf("get block %d failed: %s", i+1, getResp.Error)
		}

		_, err = outputFile.Write(getResp.Data)
		if err != nil {
			return fmt.Errorf("failed to write block %d: %w", i+1, err)
		}

		fmt.Printf("    Block %d downloaded (%d bytes)\n", i+1, len(getResp.Data))
	}

	return nil
}

// VerifyFile verifies that downloaded file matches original
func VerifyFile(originalPath, downloadedPath string) error {
	originalHash, err := computeFileHash(originalPath)
	if err != nil {
		return fmt.Errorf("failed to hash original: %w", err)
	}

	downloadedHash, err := computeFileHash(downloadedPath)
	if err != nil {
		return fmt.Errorf("failed to hash downloaded: %w", err)
	}

	if originalHash != downloadedHash {
		return fmt.Errorf("file verification failed: original=%s, downloaded=%s", originalHash, downloadedHash)
	}

	fmt.Printf("✓ File verification passed! Hash: %s\n", originalHash)
	return nil
}

func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func main() {
	action := flag.String("action", "put", "Action: put, get, test, or verify")
	filePath := flag.String("file", "", "File path to upload/download")
	outputPath := flag.String("output", "", "Output file path for download")
	hashesFile := flag.String("hashes", "", "File containing block hashes (one per line)")
	addr := flag.String("addr", "localhost:8080", "Frontend address")
	flag.Parse()

	fs, err := NewFileStorage(*addr)
	if err != nil {
		log.Fatalf("Failed to create file storage: %v", err)
	}

	switch *action {
	case "put":
		if *filePath == "" {
			log.Fatal("--file is required for put action")
		}

		blockHashes, err := fs.PutFile(*filePath)
		if err != nil {
			log.Fatalf("Put failed: %v", err)
		}

		fmt.Printf("\n✓ Upload successful!\n")
		fmt.Printf("Block hashes (%d blocks):\n", len(blockHashes))
		for i, hash := range blockHashes {
			fmt.Printf("  Block %d: %s\n", i+1, hash)
		}

		hashesOutput := *filePath + ".hashes"
		hashesFile, err := os.Create(hashesOutput)
		if err == nil {
			for _, hash := range blockHashes {
				fmt.Fprintln(hashesFile, hash)
			}
			hashesFile.Close()
			fmt.Printf("\nHashes saved to: %s\n", hashesOutput)
		}

	case "get":
		if *hashesFile == "" {
			log.Fatal("--hashes is required for get action")
		}
		if *outputPath == "" {
			log.Fatal("--output is required for get action")
		}

		hashesData, err := os.ReadFile(*hashesFile)
		if err != nil {
			log.Fatalf("Failed to read hashes file: %v", err)
		}

		hashes := strings.Split(strings.TrimSpace(string(hashesData)), "\n")
		var blockHashes []string
		for _, h := range hashes {
			h = strings.TrimSpace(h)
			if h != "" {
				blockHashes = append(blockHashes, h)
			}
		}

		err = fs.GetFile(blockHashes, *outputPath)
		if err != nil {
			log.Fatalf("Get failed: %v", err)
		}

		fmt.Printf("\n✓ Download successful!\n")

	case "test":
		if *filePath == "" {
			log.Fatal("--file is required for test action")
		}

		fmt.Println("=== UPLOAD TEST ===")
		blockHashes, err := fs.PutFile(*filePath)
		if err != nil {
			log.Fatalf("Upload failed: %v", err)
		}

		hashesOutput := *filePath + ".hashes"
		hashesFile, err := os.Create(hashesOutput)
		if err == nil {
			for _, hash := range blockHashes {
				fmt.Fprintln(hashesFile, hash)
			}
			hashesFile.Close()
		}

		fmt.Println("\n=== DOWNLOAD TEST ===")
		outputPath := *filePath + ".downloaded"
		err = fs.GetFile(blockHashes, outputPath)
		if err != nil {
			log.Fatalf("Download failed: %v", err)
		}

		fmt.Println("\n=== VERIFICATION TEST ===")
		err = VerifyFile(*filePath, outputPath)
		if err != nil {
			log.Fatalf("Verification failed: %v", err)
		}

		fmt.Printf("\n✓ All tests passed!\n")
		fmt.Printf("Original: %s\n", *filePath)
		fmt.Printf("Downloaded: %s\n", outputPath)
		fmt.Printf("Hashes: %s\n", hashesOutput)

	case "verify":
		if *filePath == "" || *outputPath == "" {
			log.Fatal("--file and --output are required for verify action")
		}

		err := VerifyFile(*filePath, *outputPath)
		if err != nil {
			log.Fatalf("Verification failed: %v", err)
		}

	default:
		log.Fatalf("Unknown action: %s (use: put, get, test, or verify)", *action)
	}
}
