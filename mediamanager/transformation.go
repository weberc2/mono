package main

type Transformation struct {
	ID          TransformationID     `json:"id"`
	Labels      []string             `json:"labels"`
	Annotations map[string]string    `json:"annotations"`
	Spec        TransformationSpec   `json:"spec"`
	Status      TransformationStatus `json:"status"`
	Error       string               `json:"error,omitempty"`
}

type TransformationID string

type TransformationStatus string

const (
	TransformationStatusPending TransformationStatus = "PENDING"
	TransformationStatusSuccess TransformationStatus = "SUCCESS"
	TransformationStatusError   TransformationStatus = "ERROR"

	// TransformationStatusFailure indicates an unrecoverable transformation
	// failure.
	TransformationStatusFailure TransformationStatus = "FAILURE"
)

type TransformationSpec struct {
	Type       TransformationType `json:"type"`
	SourcePath string             `json:"sourcePath"`
	TargetPath string             `json:"targetPath"`
}

type TransformationType string

const (
	TransformationTypeLink TransformationType = "LINK"
)
