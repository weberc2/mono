package mm

type Import struct {
	ID       ImportID     `json:"id"`
	InfoHash InfoHash     `json:"infoHash"`
	Film     *Film        `json:"film,omitempty"`
	Status   ImportStatus `json:"status"`
	Files    ImportFiles  `json:"files,omitempty"`
}

type ImportID string

type Film struct {
	Title            string         `json:"title"`
	Year             string         `json:"year"`
	PrimaryVideoFile string         `json:"primaryVideoFile"`
	PrimarySubtitles []SubtitleFile `json:"primarySubtitles,omitempty"`
}

type SubtitleFile struct {
	Path     string `json:"path"`
	Language string `json:"language"`
}

type ImportStatus string

const (
	ImportStatusPending  = "PENDING"
	ImportStatusComplete = "COMPLETE"
	ImportStatusError    = "ERROR"
)

type ImportFile struct {
	Path   string           `json:"path"`
	Status ImportFileStatus `json:"status"`
}

type ImportFileStatus string

const (
	ImportFileStatusPending  ImportFileStatus = "PENDING"
	ImportFileStatusComplete ImportFileStatus = "COMPLETE"
)

type ImportFiles []ImportFile

func (files ImportFiles) FromPath(path string) *ImportFile {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

func (files *ImportFiles) FromPathDefault(path string) (file *ImportFile) {
	if file = files.FromPath(path); file == nil {
		*files = append(
			*files,
			ImportFile{Path: path, Status: ImportFileStatusPending},
		)
		file = &(*files)[len(*files)-1]
	}
	return
}
