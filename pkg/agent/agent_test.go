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
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/heptio/sonobuoy/pkg/aggregator"
)

func TestRun(t *testing.T) {
	withFakeBinaries(t, []string{"ansible", "chroot"}, func() {
		hosts := []string{"node1", "node2", "node3", "node4", "node5"}
		withAggregator(t, hosts, []string{"ansible", "systemd_logs"}, func(aggr *aggregator.NodeAggregator) {
			for _, h := range hosts {
				cfg := &Config{
					Ansible:      true,
					SystemdLogs:  true,
					ChrootDir:    "/node",
					NodeName:     h,
					PhoneHomeURL: "http://:" + strconv.Itoa(aggregatorPort) + "/api/v1/results/by-node",
				}

				err := Run(cfg)
				if err != nil {
					t.Fatalf("Got error running agent: %v", err)
				}

				ansibleJSON := path.Join(aggr.OutputDir, h, "ansible.json")
				if _, err := os.Stat(ansibleJSON); err != nil && os.IsNotExist(err) {
					t.Errorf("Ansible agent ran, but couldn't find expected results json %v", ansibleJSON)
				}

				logsPath := path.Join(aggr.OutputDir, h, "systemd_logs.json")
				if _, err := os.Stat(logsPath); err != nil && os.IsNotExist(err) {
					t.Errorf("Systemd logs agent ran, but couldn't find expected results at %v", logsPath)
				}
			}
		})
	})
}

func TestRun_justansible(t *testing.T) {
	withFakeBinaries(t, []string{"ansible"}, func() {
		cfg := &Config{
			Ansible:      true,
			SystemdLogs:  false,
			ChrootDir:    "/node",
			NodeName:     "node1",
			PhoneHomeURL: "http://:" + strconv.Itoa(aggregatorPort) + "/api/v1/results/by-node",
		}

		withAggregator(t, []string{"node1"}, []string{"ansible"}, func(aggr *aggregator.NodeAggregator) {
			err := Run(cfg)
			if err != nil {
				t.Fatalf("Got error running agent: %v", err)
			}

			ansibleJSON := path.Join(aggr.OutputDir, "node1", "ansible.json")
			if _, err := os.Stat(ansibleJSON); err != nil && os.IsNotExist(err) {
				t.Errorf("Ansible agent ran, but couldn't find expected results json %v", ansibleJSON)
			}
		})
	})
}

func TestRun_justlogs(t *testing.T) {
	withFakeBinaries(t, []string{"chroot"}, func() {
		cfg := &Config{
			Ansible:      false,
			SystemdLogs:  true,
			ChrootDir:    "/node",
			NodeName:     "node1",
			PhoneHomeURL: "http://:" + strconv.Itoa(aggregatorPort) + "/api/v1/results/by-node",
		}

		withAggregator(t, []string{"node1"}, []string{"systemd_logs"}, func(aggr *aggregator.NodeAggregator) {
			err := Run(cfg)
			if err != nil {
				t.Fatalf("Got error running agent: %v", err)
			}

			logsPath := path.Join(aggr.OutputDir, "node1", "systemd_logs.json")
			if _, err := os.Stat(logsPath); err != nil && os.IsNotExist(err) {
				t.Errorf("Systemd logs agent ran, but couldn't find expected results at %v", logsPath)
			}
		})
	})
}

func withFakeBinaries(t *testing.T, binaries []string, callback func()) {
	// Write a fake ansible script out, which is just a shell script that exits 0
	tmpdir, err := ioutil.TempDir("", "sonobuoy_fake_binpath")
	//defer os.RemoveAll(tmpdir)
	if err != nil {
		t.Fatalf("Could not create temporary directory %v: %v", tmpdir, err)
	}

	// Put the fake binary in the PATH env var, making sure to put it back when
	// we're done
	oldpath := os.Getenv("PATH")
	os.Setenv("PATH", tmpdir+":"+os.Getenv("PATH"))
	defer os.Setenv("PATH", oldpath)

	for _, bin := range binaries {
		fakeBin := path.Join(tmpdir, bin)

		err = ioutil.WriteFile(fakeBin, []byte("#!/bin/sh\necho hello\nexit 0\n"), 0755)
		if err != nil {
			t.Fatalf("Could not create fake binary %v: %v", fakeBin, err)
		}
	}

	callback()
}

const aggregatorPort = 8090

func withAggregator(t *testing.T, nodenames []string, resultTypes []string, callback func(aggr *aggregator.NodeAggregator)) {
	// Create a temporary directory for results gathering
	tmpdir, err := ioutil.TempDir("", "sonobuoy_results_dir")
	//defer os.RemoveAll(tmpdir)
	if err != nil {
		t.Fatalf("Could not create temporary directory %v: %v", tmpdir, err)
	}

	// Launch an aggregator server for those results
	aggr := &aggregator.NodeAggregator{
		BindAddr:          ":" + strconv.Itoa(aggregatorPort),
		ExpectNodes:       nodenames,
		ExpectResultTypes: resultTypes,
		OutputDir:         tmpdir,
	}

	stop := make(chan bool)
	ready := make(chan bool, 1)
	complete := make(chan bool, 1)
	done := make(chan error)

	go func() {
		done <- aggr.GatherAndAwaitResults(stop, ready, complete)
	}()

	// Reset the default transport to clear any connection pooling
	http.DefaultTransport = &http.Transport{}

	select {
	case err := <-done:
		t.Errorf("Aggregation server closed prematurely with error: %v", err)
		break
	case <-ready:
		// Once the server is ready, call back to the test that needs
		// the aggregator, passing the aggr back.
		callback(aggr)
		stop <- true
		<-done
	case <-complete:
		stop <- true
		<-done
	}
}
