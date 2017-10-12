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
	Nodes
}

type Nodes interface {
	Nodes() []Node
}

type Node interface {
	Name() string
}

type Config interface {
	Version() string
}

type node struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
}

func (n *node) Name() string {
	return n.Metadata.Name
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
	nodes []*node
}

func newClient() *client {
	return &client{
		nodes: make([]*node, 0),
	}
}

func (c *client) Nodes() []Node {
	n := make([]Node, len(c.nodes))
	for i, node := range c.nodes {
		n[i] = node
	}
	return n
}

func (c *client) WithConfig(data []byte) error {
	return json.Unmarshal(data, &c.config)
}
func (c *client) WithNodes(data []byte) error {
	return json.Unmarshal(data, &c.nodes)
}

// client's implementation of SonobuoyData
// client.config implements Config

func New(reader io.Reader) (SonobuoyData, error) {
	client := newClient()
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

		if isNodesFile(header.Name) {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, tarReader)
			if err != nil {
				return nil, err
			}
			err = client.WithNodes(buf.Bytes())
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
func isNodesFile(filename string) bool {
	return filename == "resources/non-ns/Nodes.json" || filename == "resources/cluster/Nodes.json"
}
