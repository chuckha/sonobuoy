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

package aggregator

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestGatherAndAwaitResults(t *testing.T) {
	// Happy path
	withNewAggregator(t, []string{"node1"}, func(agg *NodeAggregator) {
		stop := make(chan bool)
		done := make(chan error)
		ready := make(chan bool, 1)
		complete := make(chan bool, 1)

		go func() {
			done <- agg.GatherAndAwaitResults(stop, ready, complete)
		}()
		select {
		case <-ready:
			break
		case err := <-done:
			t.Fatalf("Server stopped prematurely with error: %v", err)
		}

		resp := doRequest(t, "PUT", "/api/v1/results/by-node/node1/ansible", "foo")
		if resp.StatusCode != 200 {
			t.Errorf("Got non-200 response from server: %v", resp.StatusCode)
		}

		stop <- true
		err := <-done
		if err != nil {
			t.Errorf("GatherAndAwaitResults returned error %v", err)
		}

		if result, ok := agg.results["node1/ansible"]; ok {
			bytes, err := ioutil.ReadFile(result.Path)
			if string(bytes) != "foo" {
				t.Errorf("Results for node1 incorrect (got %v): %v", string(bytes), err)
			}
		} else {
			t.Errorf("GatherAndAwaitResults didn't record a result for node1")
		}

	})
}

func TestGatherAndAwaitResults_wrongnodes(t *testing.T) {
	withNewAggregator(t, []string{"node1"}, func(agg *NodeAggregator) {
		stop := make(chan bool)
		done := make(chan error)
		ready := make(chan bool, 1)
		complete := make(chan bool, 1)

		// Check in an unexpected node, should get forbidden
		go func() { done <- agg.GatherAndAwaitResults(stop, ready, complete) }()
		select {
		case <-ready:
			break
		case err := <-done:
			t.Fatalf("Server stopped prematurely with error: %v", err)
		}

		resp := doRequest(t, "PUT", "/api/v1/results/by-node/node10/ansible", "foo")
		if resp.StatusCode != 403 {
			t.Errorf("Expected a 403 forbidden for checking in an unexpected node, got %v", resp.StatusCode)
		}

		stop <- true
		if err := <-done; err != nil {
			t.Errorf("GatherAndAwaitResults returned error %v", err)
		}

		if _, ok := agg.results["node10"]; ok {
			t.Fatal("NodeAggregator accepted a result from an unexpected host")
			t.Fail()
		}
	})
}

func TestGatherAndAwaitResults_duplicates(t *testing.T) {
	// Expect 2 nodes so the server doesn't automatically quit
	expectNodes := []string{"node1", "node2"}
	withNewAggregator(t, expectNodes, func(agg *NodeAggregator) {
		stop := make(chan bool)
		done := make(chan error)
		ready := make(chan bool, 1)
		complete := make(chan bool, 1)

		go func() { done <- agg.GatherAndAwaitResults(stop, ready, complete) }()
		select {
		case <-ready:
			break
		case err := <-done:
			t.Fatalf("Server stopped prematurely with error: %v", err)
		}

		// Check in a node
		resp := doRequest(t, "PUT", "/api/v1/results/by-node/node1/ansible", "foo")
		if resp.StatusCode != 200 {
			t.Errorf("Got non-200 response from server: %v", resp.StatusCode)
		}

		// Check in the same node again, should conflict
		resp = doRequest(t, "PUT", "/api/v1/results/by-node/node1/ansible", "foo")
		if resp.StatusCode != 409 {
			t.Errorf("Expected a 409 conflict for checking in a duplicate node, got %v", resp.StatusCode)
		}

		stop <- true
		if err := <-done; err != nil {
			t.Errorf("GatherAndAwaitResults returned error %v", err)
		}

		if _, ok := agg.results["node10"]; ok {
			t.Fatal("NodeAggregator accepted a result from an unexpected host")
			t.Fail()
		}
	})
}

func withNewAggregator(t *testing.T, expectedNodes []string, callback func(*NodeAggregator)) {
	dir, err := ioutil.TempDir("", "sonobuoy_server_test")
	if err != nil {
		t.Fatal("Could not create temp directory")
		t.FailNow()
		return
	}
	defer os.RemoveAll(dir)

	agg := &NodeAggregator{
		BindAddr:    "0.0.0.0:" + strconv.Itoa(testPort),
		ExpectNodes: expectedNodes,
		OutputDir:   dir,
	}

	callback(agg)
}
