package handlers

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/pkg/archive"
	pkgxml "github.com/example/sample-app/pkg/xml"
	pkgyaml "github.com/example/sample-app/pkg/yaml"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) ImportConfig(c *gin.Context) {
	file, err := c.FormFile("config")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	config, err := pkgyaml.ParseConfigBounded(data, 64*1024)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid yaml"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (h *UploadHandler) ParseXML(c *gin.Context) {
	file, err := c.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	doc, err := pkgxml.ParseXMLNoEntities(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xml"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *UploadHandler) HandleGeneric(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var processor services.DataProcessor
	switch contentType {
	case "application/x-yaml", "text/yaml":
		processor = &services.YAMLProcessor{}
	case "application/json":
		processor = &services.JSONProcessor{}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported content type"})
		return
	}

	result, err := services.ProcessWithHandler(body, processor)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *UploadHandler) ExtractArchive(c *gin.Context) {
	file, err := c.FormFile("archive")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	tempPath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	destPath := c.PostForm("destination")
	if destPath == "" {
		destPath = "/var/uploads/extracted"
	}

	if err := archive.ExtractToPath(tempPath, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "extraction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "archive extracted", "path": destPath})
}

func (h *UploadHandler) ProcessMultipleFormats(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	data, _ := ioutil.ReadAll(f)

	yamlResult, yamlErr := pkgyaml.ParseYAMLAny(data)
	if yamlErr == nil {
		c.JSON(http.StatusOK, yamlResult)
		return
	}

	xmlResult, xmlErr := pkgxml.ParseXML(data)
	if xmlErr == nil {
		c.JSON(http.StatusOK, xmlResult)
		return
	}

	jsonResult := services.ParseJSONWithGjson(data, "@this")
	if jsonResult != "" {
		c.JSON(http.StatusOK, gin.H{"data": jsonResult})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
}

func (h *UploadHandler) ImportConfigV2(c *gin.Context) {
	file, err := c.FormFile("config")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	if file.Size > 1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	data, _ := ioutil.ReadAll(f)

	config, err := pkgyaml.ParseConfigBounded(data, 1024*1024)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid yaml"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (h *UploadHandler) ParseXMLV2(c *gin.Context) {
	file, err := c.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	if file.Size > 1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()

	data, _ := ioutil.ReadAll(f)

	decoder := xml.NewDecoder(ioutil.NopCloser(nil))
	decoder.Entity = make(map[string]string)

	doc, err := pkgxml.ParseXMLNoEntities(data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid xml"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *UploadHandler) ExtractArchiveV2(c *gin.Context) {
	file, err := c.FormFile("archive")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	tempPath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	destPath := "/var/uploads/extracted"

	if err := archive.ExtractZipBounded(tempPath, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "extraction failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "archive extracted"})
}

func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	filename := c.PostForm("filename")
	if filename == "" {
		filename = file.Filename
	}

	destPath := "/var/uploads/" + filename
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"path": destPath})
}

func (h *UploadHandler) UploadFileV2(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	allowedTypes := []string{".jpg", ".png", ".pdf", ".txt"}
	validExt := false
	for _, ext := range allowedTypes {
		if len(file.Filename) > len(ext) && file.Filename[len(file.Filename)-len(ext):] == ext {
			validExt = true
			break
		}
	}

	if !validExt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}

	if err := services.WriteFileWithBase("/var/uploads", file.Filename, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded"})
}
