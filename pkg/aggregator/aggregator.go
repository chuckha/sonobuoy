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
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/golang/glog"
)

// NodeAggregator is responsible for taking a list of nodes to expect results
// from, and starting an HTTP server which writes those out to a directory.
type NodeAggregator struct {
	// BindAddr is the address to bind the HTTP server to (eg. "0.0.0.0:8080")
	BindAddr string
	// OutputDir is the directory to write the node results
	OutputDir string
	// ExpectNodes is the list of nodes that we expect results for
	ExpectNodes []string

	results map[string]*nodeCheckin
}

// GatherAndAwaitResults begins the HTTP server and waits for all the required
// nodes to check in.  It returns nil when all results have finished as expected.
func (n *NodeAggregator) GatherAndAwaitResults(stop, ready chan bool) error {
	// Record results in the NodeAggregator struct
	n.results = make(map[string]*nodeCheckin)

	// Build a hash table of the nodes we expect, so we can reject results we don't care about
	nodemap := make(map[string]bool, len(n.ExpectNodes))
	for _, node := range n.ExpectNodes {
		nodemap[node] = true
	}

	nodeCallback := func(checkin *nodeCheckin, w http.ResponseWriter) {
		if _, ok := nodemap[checkin.NodeName]; !ok {
			glog.Warningf("Got checkin from unexpected node %v\n", checkin.NodeName)
			http.Error(
				w,
				fmt.Sprintf("Results from node %v not expected", checkin.NodeName),
				http.StatusForbidden,
			)
			return
		}

		if _, ok := n.results[checkin.NodeName]; ok {
			glog.Warningf("Got a duplicate checkin from from node %v\n", checkin.NodeName)
			http.Error(
				w,
				fmt.Sprintf("Results already received from node %v", checkin.NodeName),
				http.StatusConflict,
			)
			return
		}

		nodeDir := path.Join(n.OutputDir, checkin.NodeName)
		resultsFile := path.Join(nodeDir, checkin.ResultsType+".json")

		// Create the output directory for the results
		if err := os.MkdirAll(nodeDir, 0755); err != nil {
			glog.Errorf("Could not make directory %v: %v", nodeDir, err)
			serverError(w)
			return
		}

		// Open the results file for writing
		f, err := os.Create(resultsFile)
		if err != nil {
			glog.Errorf("Could not open output file %v for writing: %v", resultsFile, err)
			serverError(w)
			return
		}
		defer f.Close()

		// Copy the request body into the file
		io.Copy(f, checkin.Body)
		glog.Infof("wrote results to %v\n", resultsFile)
		checkin.Path = resultsFile

		// Take note that we've recorded this node
		n.results[checkin.NodeName] = checkin
		glog.Infof("Got results from %v out of %v nodes\n", len(n.results), len(n.ExpectNodes))

		// If that's all the nodes, stop serving now
		if len(n.results) == len(n.ExpectNodes) {
			stop <- true
		}

	}

	s := &server{
		BindAddr:     n.BindAddr,
		NodeCallback: nodeCallback,
	}

	err := s.Start(stop, ready)
	glog.Infof("Results aggregation server shutting down with %v of %v nodes seen\n", len(n.results), len(n.ExpectNodes))

	return err
}
