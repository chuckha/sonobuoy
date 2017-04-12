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

package agent

import (
	"fmt"
	"net/http"

	"bytes"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/ansible"
)

// Run the sonobuoy agent
func Run(cfg *Config) error {
	if cfg.PhoneHomeURL == "" {
		return fmt.Errorf("no phone home URL set, cannot continue")
	}

	output, err := ansible.Run(cfg.ChrootDir)
	if err != nil {
		return err
	}
	glog.Infof("Got ansible results: %v", string(output))
	err = submitResults(output, cfg.PhoneHomeURL+"/"+cfg.NodeName+"/ansible")
	return err
}

func submitResults(json []byte, url string) error {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(json))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error phoning home to %v: %v", url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got a %v response when phoning home to %v", resp.StatusCode, url)
	}
	return nil
}
