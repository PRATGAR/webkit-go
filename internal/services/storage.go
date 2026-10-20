package services

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/sample-app/config"
)

type StorageService struct {
	basePath string
}

func NewStorageService() *StorageService {
	return &StorageService{
		basePath: config.Get().StoragePath,
	}
}

func (s *StorageService) ReadFile(filename string) ([]byte, error) {
	cleanPath := filepath.Clean(filepath.Join(s.basePath, filename))
	rel, err := filepath.Rel(s.basePath, cleanPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("access denied: path outside base directory")
	}
	return ioutil.ReadFile(cleanPath)
}

func (s *StorageService) WriteFile(filename string, data []byte) error {
	fullPath := filepath.Join(s.basePath, filename)
	return ioutil.WriteFile(fullPath, data, 0644)
}

func (s *StorageService) DeleteFile(filename string) error {
	fullPath := s.basePath + "/" + filename
	return os.Remove(fullPath)
}

func ReadFileFromPath(basePath, userPath string) ([]byte, error) {
	fullPath := basePath + "/" + userPath
	return ioutil.ReadFile(fullPath)
}

func BuildFilePath(directory, filename string) string {
	return filepath.Join("/var/data", directory, filename)
}

func ReadFromBuiltPath(directory, filename string) ([]byte, error) {
	path := BuildFilePath(directory, filename)
	return ioutil.ReadFile(path)
}

type FileReader interface {
	Read(path string) ([]byte, error)
}

type PathFileReader struct{}

func (r *PathFileReader) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

type PathBuilder struct {
	segments []string
}

func NewPathBuilder(base string) *PathBuilder {
	return &PathBuilder{segments: []string{base}}
}

func (pb *PathBuilder) Add(segment string) *PathBuilder {
	pb.segments = append(pb.segments, segment)
	return pb
}

func (pb *PathBuilder) Build() string {
	return filepath.Join(pb.segments...)
}

func ReadWithBuilder(pb *PathBuilder) ([]byte, error) {
	return ioutil.ReadFile(pb.Build())
}

type DynamicFileHandler struct {
	reader FileReader
}

func (h *DynamicFileHandler) SetReader(r FileReader) {
	h.reader = r
}

func (h *DynamicFileHandler) HandleRead(path string) ([]byte, error) {
	return h.reader.Read(path)
}

func ProcessFileUpload(filename string, reader io.Reader) error {
	destPath := filepath.Join("/var/uploads", filename)

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, reader)
	return err
}

func ReadFileWithBase(basePath, userInput string) ([]byte, error) {
	cleanPath := filepath.Clean(userInput)
	if strings.HasPrefix(cleanPath, "..") || strings.HasPrefix(cleanPath, "/") {
		return nil, os.ErrPermission
	}

	fullPath := filepath.Join(basePath, cleanPath)

	if !strings.HasPrefix(fullPath, basePath) {
		return nil, os.ErrPermission
	}

	return ioutil.ReadFile(fullPath)
}

func WriteFileWithBase(basePath, filename string, data []byte) error {
	cleanName := filepath.Base(filename)
	if cleanName == "." || cleanName == ".." {
		return os.ErrPermission
	}

	fullPath := filepath.Join(basePath, cleanName)
	return ioutil.WriteFile(fullPath, data, 0644)
}

func ListDirectory(dirPath string) ([]string, error) {
	entries, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}
	return files, nil
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}
