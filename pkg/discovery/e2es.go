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

package discovery

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/golang/glog"
)

// kicker for e2es, look at discovery.
func rune2e(outpath string, dc *Config) error {
	var err error
	if dc.Runtests {
		var e2eout []byte

		// 1. Make the output directory.
		if err = os.MkdirAll(outpath, 0755); err != nil {
			return err
		}

		// 2. Setup the e2e test execution
		cmd := exec.Command("./battery.test", "--ginkgo.focus=\""+dc.TestFocusRegex+"\"", "--ginkgo.skip=\""+dc.TestSkipRegex+"\"", "--provider=\""+dc.Provider+"\"", "--report-dir="+outpath, "--ginkgo.noColor=true")
		cmd.Env = os.Environ()

		// TODO: OK this is a mess in the framework tooling.
		if len(dc.kubeconfig) > 0 {
			cmd.Env = append(cmd.Env, "KUBECONFIG="+dc.kubeconfig)
		}

		glog.Infof("Executing e2es: [%v %v]", cmd.Path, cmd.Args)

		// 3. blocking run
		e2eout, err = cmd.CombinedOutput()
		if e2eout != nil {
			if werr := ioutil.WriteFile(outpath+"/e2e.txt", e2eout, 0644); werr != nil {
				glog.Warningf("Failed to write e2e.txt file (%v)", werr)
			}
		}

	}
	return err
}
