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

// Package aggregator is responsible for hosting an HTTP server which
// aggregates results from all of the nodes that are running sonobuoy agent. It
// is not responsible for dispatching the nodes (see pkg/dispatch), only
// expecting their results.
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
	// ExpectResultTypes is the list of types of results we need from each host
	ExpectResultTypes []string

	results map[string]*nodeCheckin
}

// GatherAndAwaitResults begins the HTTP server and waits for all the required
// nodes to check in.
//
// The stop chan can be written to by callers to stop the HTTP server
// The ready chan is written to when the server is ready for connections
// The complete chan is written to when the ndoes are all checked in
func (n *NodeAggregator) GatherAndAwaitResults(stop <-chan bool, ready chan<- bool, complete chan<- bool) error {
	// Record results in the NodeAggregator struct
	n.results = make(map[string]*nodeCheckin)

	// Build a hash table of the nodes we expect, so we can reject results we don't care about
	nodemap := make(map[string]bool, len(n.ExpectNodes))
	for _, node := range n.ExpectNodes {
		nodemap[node] = true
	}

	successEvents := make(chan bool)

	// This is called every time the HTTP server gets a well-formed request with
	// results. This method is responsible for returning with things like a 409
	// conflict if a node has checked in twice (or a 403 forbidden if a node isn't
	// expected). It also knows when all nodes have checked in, and is responsible
	// for signaling over the "complete" channel when that happens.
	nodeCallback := func(checkin *nodeCheckin, w http.ResponseWriter) {
		// Only gather from expected nodes
		if _, ok := nodemap[checkin.NodeName]; !ok {
			glog.Warningf("Got checkin from unexpected node %v\n", checkin.NodeName)
			http.Error(
				w,
				fmt.Sprintf("Results from node %v not expected", checkin.NodeName),
				http.StatusForbidden,
			)
			return
		}

		// Nodes can't check in the same result twice
		checkinID := checkin.NodeName + "/" + checkin.ResultsType
		if _, ok := n.results[checkinID]; ok {
			glog.Warningf("Got a duplicate checkin of %v\n", checkinID)
			http.Error(
				w,
				fmt.Sprintf("Results already received for %v", checkinID),
				http.StatusConflict,
			)
			return
		}

		// Create the output directory for the results
		nodeDir := path.Join(n.OutputDir, checkin.NodeName)
		resultsFile := path.Join(nodeDir, checkin.ResultsType+".json")
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

		// Take note that we've recorded this node. Detecting when we're done is part of
		// a select loop below, to prevent race conditions with concurrent checkins.
		n.results[checkinID] = checkin

		successEvents <- true
	}

	// Start the server with the above callback
	s := &server{
		BindAddr:     n.BindAddr,
		NodeCallback: nodeCallback,
	}

	result := make(chan error, 1)
	go func() {
		result <- s.Start(stop, ready)
		glog.Infof("Results aggregation server shutting down with %v of %v nodes seen\n", len(n.results), len(n.ExpectNodes)*len(n.ExpectResultTypes))
	}()

loop:
	for {
		select {
		case err := <-result:
			return err
		case <-successEvents:
			glog.Infof("Got %v out of %v expected results\n", len(n.results), len(n.ExpectNodes)*len(n.ExpectResultTypes))
			// If that's all the nodes, let the caller know (and they can stop the server)
			if len(n.results) >= len(n.ExpectNodes)*len(n.ExpectResultTypes) {
				complete <- true
				break loop
			}
			break
		}
	}

	return nil
}
