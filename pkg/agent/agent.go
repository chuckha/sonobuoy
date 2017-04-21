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

// Package agent is responsible for gathering information about a local
// node and submitting the results to a master instance, via the phone-home
// URL.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/hostdata"
)

// Run the sonobuoy agent
func Run(cfg *Config) error {
	if cfg.PhoneHomeURL == "" {
		return fmt.Errorf("no phone home URL set, cannot continue")
	}
	urlBase := cfg.PhoneHomeURL + "/" + cfg.NodeName

	// 1. Run ansible
	if cfg.Ansible {
		doRequest(urlBase+"/ansible", func() (io.Reader, error) {
			output, err := hostdata.RunAnsible(cfg.ChrootDir)
			glog.V(5).Infof("Got ansible results: %v", string(output))
			return bytes.NewReader(output), err
		})
	}

	// 2. Gather systemd logs with journalctl
	if cfg.SystemdLogs {
		// up-scope the file we write out, so we can delete it after it's uploaded.
		// We don't want to do this in the callback function, or else it'll get
		// deleted before we can upload it.
		var outfile *os.File
		defer func() {
			if outfile != nil {
				outfile.Close()
				os.Remove(outfile.Name())
			}
		}()

		doRequest(urlBase+"/systemd_logs", func() (io.Reader, error) {
			f, err := hostdata.RunSystemdLogs(time.Duration(cfg.SystemdLogMinutes)*time.Minute, cfg.ChrootDir)
			if err != nil {
				return nil, err
			}
			glog.V(5).Infof("Got systemd logs: %v", f)
			outfile, err = os.Open(f)
			if err != nil {
				return nil, err
			}
			return outfile, err
		})
	}

	return nil
}

// doRequest calls the given callback which returns an io.Reader, and submits
// the results, with error handling, and falls back on uploading JSON with the
// error message if the callback fails. (This way, problems gathering data
// don't result in the server waiting forever for results that will never
// come.)
func doRequest(url string, callback func() (io.Reader, error)) error {
	client := &http.Client{}
	input, err := callback()
	if err != nil {
		glog.Errorf("Error gathering host data: %v", err)

		// If the callback couldn't get the data, we should send the reason why to
		// the server.
		errobj := map[string]string{
			"error": err.Error(),
		}
		errbody, err := json.Marshal(errobj)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(errbody))
		if err != nil {
			return err
		}

		// And if we can't even do that, log it.
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			glog.Errorf("Could not send error message to phone-home URL (%v): %v", url, err)
		}

		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, input)
	if err != nil {
		glog.Errorf("Error constructing phone-home request to %v: %v", url, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error phoning home to %v: %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		// TODO: retry logic for something like a 429 or otherwise
		return fmt.Errorf("Got a %v response when phoning home to %v", resp.StatusCode, url)
	}

	return nil
}
