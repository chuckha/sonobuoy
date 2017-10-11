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
	file, err := os.Open(fmt.Sprintf("test_dta/%v/diagnostic.tar.gz", version))
	if err != nil {
		t.Fatalf("could not read test file: %v", err)
	}
	return file
}

func TestZeroEightClient(t *testing.T) {
	file := tarballData(t, v8)
	defer file.Close()
	client := tarball.New(file)
	sv := client.ServerVersion()
	if sv.Version() != "v1.7.2" {
		t.Fatalf("expected v1.7.2 as the server version, got %v", sv.Versions())
	}
}
