package tarball

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

type SonobuoyData interface {
	Config
}

type Config interface {
	Version() string
}

// config is the config.json object (the sonobuoy configuration used for this particular run)
type config struct {
	SonobuoyVersion string `json:"Version"`
}

func (c *config) Version() string {
	return c.SonobuoyVersion
}

type client struct {
	config
}

func (c *client) WithConfig(data []byte) error {
	return json.Unmarshal(data, &c.config)
}

// client's implementation of SonobuoyData
// client.config implements Config

func New(reader io.Reader) (SonobuoyData, error) {
	client := &client{}
	// 0. Setup
	gz, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(gz)
	// 1. Read through it and grab the files we care about
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if isConfigFile(header.Name) {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, tarReader)
			if err != nil {
				return nil, err
			}
			err = client.WithConfig(buf.Bytes())
			if err != nil {
				return nil, err
			}
		}
	}

	// 2. read the tarball a file-at-a-time and populate various structs
	return client, nil
}

func isConfigFile(filename string) bool {
	return filename == "config.json" || filename == "meta/config.json"
}
