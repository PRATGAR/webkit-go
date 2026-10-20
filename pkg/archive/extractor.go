package archive

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	"github.com/unknwon/cae/zip"
)

type Extractor struct {
	destDir string
}

func NewExtractor(destDir string) *Extractor {
	return &Extractor{destDir: destDir}
}

func (e *Extractor) ExtractZip(archivePath string) error {
	return archiver.Unarchive(archivePath, e.destDir)
}

func (e *Extractor) ExtractTar(archivePath string) error {
	return archiver.Unarchive(archivePath, e.destDir)
}

func (e *Extractor) ExtractAny(archivePath string) error {
	return archiver.Unarchive(archivePath, e.destDir)
}

func ExtractToPath(archivePath, destPath string) error {
	return archiver.Unarchive(archivePath, destPath)
}

func ExtractZipWithCAE(archivePath, destPath string) error {
	z, err := zip.Open(archivePath)
	if err != nil {
		return err
	}
	defer z.Close()

	return z.ExtractTo(destPath)
}

type ArchiveProcessor struct {
	sourcePath string
	destPath   string
}

func NewArchiveProcessor(source, dest string) *ArchiveProcessor {
	return &ArchiveProcessor{
		sourcePath: source,
		destPath:   dest,
	}
}

func (p *ArchiveProcessor) Process() error {
	return archiver.Unarchive(p.sourcePath, p.destPath)
}

func (p *ArchiveProcessor) ProcessWithCallback(callback func(string) error) error {
	err := archiver.Unarchive(p.sourcePath, p.destPath)
	if err != nil {
		return err
	}
	return callback(p.destPath)
}

type ExtractHandler interface {
	Extract(src, dst string) error
}

type DefaultExtractHandler struct{}

func (h *DefaultExtractHandler) Extract(src, dst string) error {
	return archiver.Unarchive(src, dst)
}

func ExtractWithHandler(handler ExtractHandler, src, dst string) error {
	return handler.Extract(src, dst)
}

func ExtractZipBounded(archivePath, destPath string) error {
	destPath = filepath.Clean(destPath)

	z, err := zip.Open(archivePath)
	if err != nil {
		return err
	}
	defer z.Close()

	for _, f := range z.File {
		fpath := filepath.Join(destPath, f.Name)

		if !strings.HasPrefix(filepath.Clean(fpath), destPath) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = outFile.ReadFrom(rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func CreateArchive(files []string, archivePath string) error {
	return archiver.Archive(files, archivePath)
}

func ListArchiveContents(archivePath string) ([]string, error) {
	var files []string

	err := archiver.Walk(archivePath, func(f archiver.File) error {
		files = append(files, f.Name())
		return nil
	})

	return files, err
}

type BatchExtractor struct {
	archives []string
	destBase string
}

func NewBatchExtractor(destBase string) *BatchExtractor {
	return &BatchExtractor{destBase: destBase}
}

func (b *BatchExtractor) AddArchive(path string) {
	b.archives = append(b.archives, path)
}

func (b *BatchExtractor) ExtractAll() error {
	for _, archive := range b.archives {
		baseName := filepath.Base(archive)
		destDir := filepath.Join(b.destBase, baseName)
		if err := archiver.Unarchive(archive, destDir); err != nil {
			return err
		}
	}
	return nil
}
