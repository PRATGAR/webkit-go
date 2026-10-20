package handlers

import (
	"io/ioutil"
	"net/http"
	"strings"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/internal/utils"
	"github.com/example/sample-app/pkg/archive"
)

type FileHandler struct {
	storage *services.StorageService
}

func NewFileHandler() *FileHandler {
	return &FileHandler{
		storage: services.NewStorageService(),
	}
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	filename := c.Query("filename")

	content, err := h.storage.ReadFile(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *FileHandler) ReadDocument(c *gin.Context) {
	directory := c.Query("dir")
	filename := c.Query("file")

	baseDir := "/var/data/documents"
	target := filepath.Clean(filepath.Join(baseDir, directory, filename))
	rel, err := filepath.Rel(baseDir, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access denied: invalid path"})
		return
	}

	content, err := ioutil.ReadFile(target)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Data(http.StatusOK, "text/plain", content)
}

func (h *FileHandler) ArchiveFiles(c *gin.Context) {
	var req models.ArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	extractor := archive.NewExtractor(req.DestPath)
	if err := extractor.ExtractZip(req.SourcePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "extraction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "archive extracted"})
}

func (h *FileHandler) GetDocument(c *gin.Context) {
	docID := c.Param("id")

	query := "SELECT id, owner_id, title, file_path FROM documents WHERE id = " + docID
	rows, err := services.QueryWithString(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	var doc models.Document
	rows.Scan(&doc.ID, &doc.OwnerID, &doc.Title, &doc.FilePath)

	content, err := ioutil.ReadFile(doc.FilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file read failed"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *FileHandler) DynamicRead(c *gin.Context) {
	basePath := c.Query("base")
	filePath := c.Query("path")

	var reader services.FileReader = &services.PathFileReader{}

	handler := &services.DynamicFileHandler{}
	handler.SetReader(reader)

	fullPath := filepath.Join(basePath, filePath)
	content, err := handler.HandleRead(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Data(http.StatusOK, "text/plain", content)
}

func (h *FileHandler) ResolvePath(c *gin.Context) {
	var req struct {
		Segments []string `json:"segments"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	builder := services.NewPathBuilder("/var/data")
	for _, segment := range req.Segments {
		builder.Add(segment)
	}

	content, err := services.ReadWithBuilder(builder)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Data(http.StatusOK, "text/plain", content)
}

func (h *FileHandler) DownloadFileV2(c *gin.Context) {
	filename := c.Query("filename")

	if !utils.ValidateFilename(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	content, err := services.ReadFileWithBase("/var/uploads", filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(filename))
	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *FileHandler) ReadDocumentV2(c *gin.Context) {
	filename := c.Query("file")

	if !utils.ValidateFilename(filename) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}

	content, err := services.ReadFileWithBase("/var/documents", filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Data(http.StatusOK, "text/plain", content)
}

func (h *FileHandler) ArchiveFilesV2(c *gin.Context) {
	var req models.ArchiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := archive.ExtractZipBounded(req.SourcePath, "/var/uploads/extracted"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "extraction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "archive extracted"})
}

func (h *FileHandler) GetDocumentV2(c *gin.Context) {
	docID := c.Param("id")
	currentUserID, _ := c.Get("user_id")

	if !utils.ValidateNumericID(docID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}

	rows, err := services.ExecuteNamedQuery(
		"SELECT id, owner_id, title, file_path FROM documents WHERE id = $1 AND (owner_id = $2 OR is_public = true)",
		docID, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}

	var doc models.Document
	rows.Scan(&doc.ID, &doc.OwnerID, &doc.Title, &doc.FilePath)

	content, err := services.ReadFileWithBase("/var/documents", filepath.Base(doc.FilePath))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file read failed"})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	directory := c.Query("dir")

	files, err := services.ListDirectory("/var/uploads/" + directory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) ListFilesV2(c *gin.Context) {
	files, err := services.ListDirectory("/var/uploads")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *FileHandler) AttachReport(c *gin.Context) {
	id := c.Param("id")
	downloadName := c.Query("download_name")

	storedPath := filepath.Join("/var/reports", filepath.Base(id)+".pdf")

	c.FileAttachment(storedPath, downloadName)
}

func (h *FileHandler) InlineReport(c *gin.Context) {
	id := c.Param("id")

	storedPath := filepath.Join("/var/reports", filepath.Base(id)+".pdf")

	c.File(storedPath)
}

func (h *FileHandler) DownloadSafeArchive(c *gin.Context) {
	fileName := c.Query("file")

	// Flawed sanitizer: single-pass replacement of "../"
	// Attacker input: "....//....//etc/passwd" becomes "../../etc/passwd"
	cleanName := strings.Replace(fileName, "../", "", -1)
	fullPath := filepath.Join("/var/data/archives", cleanName)

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+cleanName)
	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *FileHandler) ConfinedRead(c *gin.Context) {
	fileName := c.Query("path")
	baseDir := "/var/data/userfiles"

	target := filepath.Join(baseDir, fileName)
	cleanTarget := filepath.Clean(target)

	// Strict path confinement guard: target must remain strictly inside baseDir
	rel, err := filepath.Rel(baseDir, cleanTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		c.JSON(http.StatusBadRequest, gin.H{"error": "access denied: outside sandbox"})
		return
	}

	content, err := ioutil.ReadFile(cleanTarget)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Data(http.StatusOK, "text/plain", content)
}
