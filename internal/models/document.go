package models

import (
	"time"
)

type Document struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	FilePath    string    `json:"file_path"`
	MimeType    string    `json:"mime_type"`
	Size        int64     `json:"size"`
	IsPublic    bool      `json:"is_public"`
	AccessLevel string    `json:"access_level"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DocumentQuery struct {
	OwnerID     int64  `json:"owner_id"`
	SearchTerm  string `json:"search_term"`
	MimeType    string `json:"mime_type"`
	AccessLevel string `json:"access_level"`
}

type FileUpload struct {
	Filename    string `json:"filename"`
	Directory   string `json:"directory"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"-"`
}

type ArchiveRequest struct {
	SourcePath  string   `json:"source_path"`
	DestPath    string   `json:"dest_path"`
	FileList    []string `json:"file_list"`
	Format      string   `json:"format"`
	Compression string   `json:"compression"`
}

type ReportConfig struct {
	TemplateID   string                 `json:"template_id"`
	TemplateName string                 `json:"template_name"`
	Parameters   map[string]interface{} `json:"parameters"`
	OutputFormat string                 `json:"output_format"`
	Title        string                 `json:"title"`
	Subtitle     string                 `json:"subtitle"`
}
