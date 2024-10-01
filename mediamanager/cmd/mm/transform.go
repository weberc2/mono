package main

// Transform is a file transformation.
type Transform struct {
	// ID identifies the transform.
	ID TransformID `json:"id"`

	// Spec contains the transorm specification.
	Spec TransformSpec `json:"spec"`

	// Files contains the status information for each file in the transform.
	Files []TransformFile `json:"files,omitempty"`
}

// TransformID identifies transformations.
type TransformID string

// TransformSpec specifies the transformation.
type TransformSpec struct {
	// InfoHash identifies the download which is the source of the transform.
	InfoHash InfoHash `json:"infoHash"`

	// Type indicates which transform type to use.
	Type TransformType `json:"type"`

	// Film contains the configuration for `FILM` transform.
	Film *FilmTransform `json:"film,omitempty"`
}

// TransformType enumerates the types of transforms.
type TransformType string

const (
	TransformTypeFilm TransformType = "FILM"
)

// TransformFile contains the status information for a file in a transform.
type TransformFile struct {
	// Path is the path to the file in the download.
	Path string `json:"path"`

	// Status is the status of the transformation of the file.
	Status TransformFileStatus `json:"status"`
}

// TransformFileStatus enumerates the transform file statuses.
type TransformFileStatus string

const (
	TransformFileStatusPending  TransformFileStatus = "PENDING"
	TransformFileStatusProgress TransformFileStatus = "PROGRESS"
	TransformFileStatusSuccess  TransformFileStatus = "SUCCESS"
)
