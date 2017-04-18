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
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/heptio/sonobuoy/pkg/aggregator"
)

func TestRun_ansible(t *testing.T) {
	withFakeAnsible(t, func() {
		cfg := &Config{
			Ansible:      true,
			ChrootDir:    "/node",
			NodeName:     "node1",
			PhoneHomeURL: "http://:" + strconv.Itoa(aggregatorPort) + "/api/v1/results/by-node",
		}

		withAggregator(t, "node1", func(aggr *aggregator.NodeAggregator) {
			err := Run(cfg)
			if err != nil {
				t.Errorf("Got error running agent: %v", err)
			}

			ansibleJSON := path.Join(aggr.OutputDir, "node1", "ansible.json")
			if _, err := os.Stat(ansibleJSON); err != nil && os.IsNotExist(err) {
				t.Errorf("Ansible agent ran, but couldn't find expected reults json %v", ansibleJSON)
			}
		})
	})
}

func withFakeAnsible(t *testing.T, callback func()) {
	// Write a fake ansible script out, which is just a shell script that exits 0
	tmpdir, err := ioutil.TempDir("", "fake_ansible")
	defer os.RemoveAll(tmpdir)
	if err != nil {
		t.Fatalf("Could not create temporary directory %v: %v", tmpdir, err)
	}

	fakeAnsible := path.Join(tmpdir, "ansible")

	err = ioutil.WriteFile(fakeAnsible, []byte("#!/bin/sh\nexit 0\n"), 0755)
	if err != nil {
		t.Fatalf("Could not create fake ansible binary %v: %v", fakeAnsible, err)
	}

	// Put the fake ansible in the PATH env var, making sure to put it back when
	// we're done
	oldpath := os.Getenv("PATH")
	os.Setenv("PATH", tmpdir+":"+os.Getenv("PATH"))
	defer os.Setenv("PATH", oldpath)

	callback()
}

const aggregatorPort = 8090

func withAggregator(t *testing.T, nodename string, callback func(aggr *aggregator.NodeAggregator)) {
	// Create a temporary directory for results gathering
	tmpdir, err := ioutil.TempDir("", "results_dir")
	defer os.RemoveAll(tmpdir)
	if err != nil {
		t.Fatalf("Could not create temporary directory %v: %v", tmpdir, err)
	}

	// Launch an aggregator server for those results
	aggr := &aggregator.NodeAggregator{
		BindAddr:    ":" + strconv.Itoa(aggregatorPort),
		ExpectNodes: []string{nodename},
		OutputDir:   tmpdir,
	}

	stop := make(chan bool)
	ready := make(chan bool, 1)
	complete := make(chan bool, 1)
	done := make(chan error, 1)

	defer func() {
		close(stop)
		close(ready)
		close(complete)
		close(done)
	}()

	go func() {
		done <- aggr.GatherAndAwaitResults(stop, ready, complete)
	}()

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
