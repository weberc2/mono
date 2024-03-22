package main

// Image describes the image to be built.
type Image struct {
	// Source contains the source details for the image.
	Source ImageSource `json:"source"`

	// Tags is the fully-qualified image tag (i.e., including the regsitry URI,
	// the image name, and the image tag portion). All tags will be pushed.
	Tags []string `json:"tags"`
}

// ImageRegistry is an image registry.
type ImageRegistry struct {
	// URI is the container image registry URI.
	URI string

	// AuthType is the type of authentication to use for pushing the image to
	// the target registry.
	AuthType PushAuthType

	// UsernameSecret is the address of the secret containing the username. It
	// is ignored if the `AuthType` field is not `PushAuthTypeUsernamePassword`.
	UsernameSecret SecretRef `json:"usernameSecret"`

	// PasswordSecret is the address of the secret containing the password. It
	// is ignored if the `AuthType` field is not `PushAuthTypeUsernamePassword`.
	PasswordSecret SecretRef `json:"passwordSecret"`
}

type PushAuthType string

const (
	PushAuthTypeUsernamePassword PushAuthType = "USERNAMEPASSWORD"
)

// ImageSource is the address information for the
type ImageSource struct {
	// Type is the type of the image source.
	Type ImageSourceType `json:"type"`

	// Git contains the image source details for a git image source. This is
	// ignored unless the `Type` field is `ImageSourceTypeGit`.
	Git ImageSourceGit `json:"git"`
}

// ImageSourceType identifies the type of an image source.
type ImageSourceType string

const (
	ImageSourceTypeGit ImageSourceType = "GIT"
)

// ImageSourceGit is the image source for a git repository.
type ImageSourceGit struct {
	// Repository is the URI for the git repo. E.g.,
	// `git@github.com:weberc2/mono.git`.
	Repository string `json:"repository"`

	// Branch is the branch within the git repo to build. If omitted, the
	// default branch for the repository will be used.
	Branch string `json:"branch,omitempty"`

	// Directory is the directory within the git repo to build. It must contain
	// all of the artifacts required to build the image. If omitted, the root of
	// the repo will be used.
	Directory string `json:"directory,omitempty"`

	// Dockerfile is a relative path within the build directory to the
	// dockerfile. If omitted, it's assumed there is a `Dockerfile` in the root
	// of the build directory.
	Dockerfile string `json:"dockerfile,omitempty"`

	// PrivateKeySecret is a reference to the Kubernetes secret containing the
	// PEM-encoded git private key. This will be mounted into the environment
	// of the image build job, and as such the secret must be in the same
	// namespace that the build jobs run in.
	PrivateKeySecret SecretRef `json:"privateKeySecret"`
}

// SecretRef is the address of a Kubernetes secret.
type SecretRef struct {
	// Name is the name of a Kubernetes secret.
	Name string

	// Field is the optional field within the Kubernetes secret. If no field is
	// specified, then it references the entire secret contents.
	Field string
}
