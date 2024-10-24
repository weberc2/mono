package mm

import "github.com/danielgtaylor/huma/v2"

type Download struct {
	ID       InfoHash       `json:"id"`
	Status   DownloadStatus `json:"status"`
	Size     uint64         `json:"size,omitempty"`
	Progress uint64         `json:"progress,omitempty"`
	Files    DownloadFiles  `json:"files,omitempty"`
}

type DownloadStatus string

const (
	DownloadStatusPending  DownloadStatus = "PENDING"
	DownloadStatusMetadata DownloadStatus = "METADATA"
	DownloadStatusProgress DownloadStatus = "PROGRESS"
	DownloadStatusComplete DownloadStatus = "COMPLETE"
)

type DownloadFiles []DownloadFile

func (files *DownloadFiles) MarshalJSON() ([]byte, error) {
	return (*Slice[DownloadFile])(files).MarshalJSON()
}

func (files *DownloadFiles) Schema(r huma.Registry) *huma.Schema {
	return (*Slice[DownloadFile])(files).Schema(r)
}

func (files DownloadFiles) ToMap() map[string]*DownloadFile {
	out := map[string]*DownloadFile{}
	for i := range files {
		out[files[i].Path] = &files[i]
	}
	return out
}

func (files DownloadFiles) FromPath(path string) *DownloadFile {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

type DownloadFile struct {
	Path     string `json:"path"`
	Size     uint64 `json:"size"`
	Progress uint64 `json:"progress"`
}

func (file *DownloadFile) Complete() bool {
	return file.Progress >= file.Size
}
