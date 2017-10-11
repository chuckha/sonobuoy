package tarball

import (
	"bytes"
	"io"
)

type K8sVersion struct {
	gitVersion string `json:"gitVersion"`
}
func NewK8sVersion(serverVersionFile []byte) {}

type ServerVersion interface {
	Version() string
}




type Plugin interface {
	// Files can be used to see which files the Plugin exposes
	Name() string
}

type E2EPlugin struct {
	name string
}
func (e *E2EPlugin) Name() string {
	return "e2e"
}
func (e *E2EPlugin) Results() *TestResults {
	// unmarshal the results file into a TestResults object
}
func (e *E2EPlugin) Logs() *TestResultLog {
	// unmarshal the e2e.logs file
}

type SonobuoyData interface {
	// Get a list of plugins that this Sonobuoy run was configured with
	Plugins() []Plugin
	ServerVersion() ServerVersion
}

func New(reader io.Reader) SonobuoyData {
	// 1. read in the tarball and find the version
	// 2. read the tarball a file-at-a-time and populate various structs
	return &tarball{...}
}

func getTarballFromS3() io.Reader {
	bytes.NewReader([]byte("somebytedatagoeshere"))
}
