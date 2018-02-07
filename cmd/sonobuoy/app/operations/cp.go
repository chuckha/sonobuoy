/*
Copyright 2018 Heptio Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package operations

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type FileSpec struct {
	PodNamespace string
	PodName      string
	File         string
}

func CopyResults(cmderr io.Writer, src, dest FileSpec) error {
	f := cmdutil.NewFactory(nil)
	config, err := f.ClientConfig()
	if err != nil {
		return fmt.Errorf("could not get client config: %v", err)
	}
	clientset, err := f.ClientSet()
	if err != nil {
		return fmt.Errorf("could not make a new clientset: %v", err)
	}

	reader, outStream := io.Pipe()
	options := &cmd.ExecOptions{
		StreamOptions: cmd.StreamOptions{
			In:  nil,
			Out: outStream,
			Err: cmderr,

			Namespace: src.PodNamespace,
			PodName:   src.PodName,
		},

		// TODO: Improve error messages by first testing if 'tar' is present in the container?
		Command:  []string{"tar", "cf", "-", src.File},
		Executor: &cmd.DefaultRemoteExecutor{},
	}

	go func() {
		options.Config = config
		options.PodClient = clientset.Core()

		defer outStream.Close()

		options.Validate()
		options.Run()
	}()
	prefix := strings.TrimLeft(src.File, "/")
	prefix = path.Clean(prefix)
	return untarAll(reader, dest.File, prefix)
}

func untarAll(reader io.Reader, destFile, prefix string) error {
	entrySeq := -1

	// TODO: use compression here?
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		entrySeq++
		mode := header.FileInfo().Mode()
		outFileName := path.Join(destFile, header.Name[len(prefix):])
		baseName := path.Dir(outFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(outFileName, 0755); err != nil {
				return err
			}
			continue
		}

		// handle coping remote file into local directory
		if entrySeq == 0 && !header.FileInfo().IsDir() {
			exists, err := dirExists(outFileName)
			if err != nil {
				return err
			}
			if exists {
				outFileName = filepath.Join(outFileName, path.Base(header.Name))
			}
		}

		if mode&os.ModeSymlink != 0 {
			err := os.Symlink(header.Linkname, outFileName)
			if err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(outFileName)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	if entrySeq == -1 {
		//if no file was copied
		errInfo := fmt.Sprintf("error: %s no such file or directory", prefix)
		return errors.New(errInfo)
	}
	return nil
}
