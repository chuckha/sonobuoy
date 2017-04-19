/*
Copyright 2017 Heptio Inc.

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

// Package ansible is responsible for wrapping calls to the ansible binary
// (calling the "setup" module), and returning the resulting JSON as a byte
// array.
package ansible

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

const (
	// ConfigLocation is the directory under the configured ansible output path under which the configuration is stored
	ConfigLocation = "config"
	// ResultsLocation is the directory under the configured ansible output path under which the resulting host data
	ResultsLocation = "results"
)

// Config represents the configuration of ansible for reaching out to physical
// hosts in the cluster.
type Config struct {
	// Chroot is the directory contianing the host's filesystem
	Chroot string
}

// Run ansible locally in the given chroot
func Run(chroot string) ([]byte, error) {
	// Find the ansible command
	ansiblePath, err := exec.LookPath("ansible")
	if err != nil {
		return nil, fmt.Errorf("could not find ansible binary in $PATH: %v", err)
	}

	var out bytes.Buffer
	out.Grow(16384) // Reasonable guess for output length
	cmd := exec.Command(
		ansiblePath,
		"all",
		// The comma is intentional, adding a trailing comma after is what convinces ansible to do the right thing.
		"--inventory-file="+chroot+",",
		"--connection=chroot",
		"--module-name=setup",
		"--one-line",
	)
	cmd.Stdout = &out
	cmd.Stderr = os.Stdout

	// Start the command in the background
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		glog.Errorf("Error running ansible: %v\n", out.String())
		return nil, fmt.Errorf("ansible returned an error: %v", err)
	}

	// This is kind of hackish, ansible returns output that looks like:
	//
	// /node | SUCCESS => {...}
	//
	// And we just want the json inside the {...}. So skip the first bit of the
	// string before the first `{`.
	outbytes := out.Bytes()
	beginloc := 0
	for beginloc < len(outbytes) {
		if outbytes[beginloc] == '{' {
			break
		}
		beginloc++
	}

	return outbytes[beginloc:len(outbytes)], err
}
