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

package hostdata

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/golang/glog"
)

// RunSystemdLogs gets logs from the current host using journalctl, writes them out
// to a file, and returns the path to that file.
func RunSystemdLogs(duration time.Duration, chroot string) (string, error) {
	t, err := ioutil.TempFile("", "sonobuoy_logs")
	if err != nil {
		return "", fmt.Errorf("could not create temporary file for log gathering: %v", err)
	}
	t.Close() // let the journalctl command write to the file
	logfile := t.Name()

	startDate := time.Now().UTC().Add(0 - duration)
	startDateStr := startDate.Format("2006-01-02 15:04:05 UTC")

	// We just pass the whole command as a string to `/bin/sh -c` and let the
	// shell handle the file redirection. With the way this is structured,
	// `journalctl` is run in a chroot, but the out file is stored in the local
	// fs.
	cmdStr := fmt.Sprintf("chroot '%s' /bin/journalctl -o json -a --no-pager --since '%s' >'%s'", chroot, startDateStr, logfile)

	var out bytes.Buffer
	out.Grow(4096) // Allocate some space for any error messages from journalctl
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Start the command in the background
	err = cmd.Start()
	if err != nil {
		return logfile, err
	}

	err = cmd.Wait()
	if err != nil {
		glog.Errorf("Error running journalctl: %v\n", out.String())
		return logfile, fmt.Errorf("journalctl returned an error: %v", err)
	}

	return logfile, err
}
