package tarball_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/heptio/sonobuoy/pkg/tarball"
)

const (
	v8 = "0.8"
	v9 = "0.9"
)

func tarballData(t *testing.T, version string) *os.File {
	file, err := os.Open(fmt.Sprintf("test_data/%v/diagnostic.tar.gz", version))
	if err != nil {
		t.Fatalf("could not read test file: %v", err)
	}
	return file
}

func TestNodes(t *testing.T) {
	file := tarballData(t, v9)
	defer file.Close()
	client, err := tarball.New(file)
	if err != nil {
		t.Fatalf("could not make a new client: %v", err)
	}
	if len(client.Nodes()) != 3 {
		t.Fatalf("could not figure out Sonobuoy version")
	}
}

func TestConfigVersion(t *testing.T) {
	testcases := []struct {
		name      string
		sbVersion string
		expected  string
	}{
		{
			name:      "simple 0.8 version",
			sbVersion: v8,
			expected:  "v0.8.2",
		},
		{
			name:      "simple 0.9 version",
			sbVersion: v9,
			expected:  "v0.9.0",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			file := tarballData(t, tc.sbVersion)
			defer file.Close()
			client, err := tarball.New(file)
			if err != nil {
				t.Fatalf("could not make a new client: %v", err)
			}
			if client.Version() != tc.expected {
				t.Fatalf("could not figure out Sonobuoy version")
			}
		})
	}
}
