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

// Package aggregation is responsible for hosting an HTTP server which
// aggregates results from all of the nodes that are running sonobuoy agent. It
// is not responsible for dispatching the nodes (see pkg/dispatch), only
// expecting their results.
package aggregation

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/plugin"
	"github.com/viniciuschiele/tarx"
)

// Aggregator is responsible for taking results from an HTTP server (configured
// elsewhere), saving them to the filesystem, and keeping track of what has
// been seen so far, so that we can return when all expected results are
// present and accounted for.
type Aggregator struct {
	// OutputDir is the directory to write the node results
	OutputDir string
	// Results stores a map of check-in results the server has seen
	Results map[string]*plugin.Result
	// ExpectedResults stores a map of results the server should expect
	ExpectedResults map[string]*plugin.ExpectedResult

	// resultEvents is a channel that is written to when results are seen
	// by the server, so we can block until we're done.
	resultEvents chan *plugin.Result
	// resultsMutex prevents race conditions if two identical results
	// come in at the same time.
	resultsMutex sync.Mutex
}

// NewAggregator constructs a new Aggregator object to write the given result
// set out to the given output directory.
func NewAggregator(outputDir string, expected []plugin.ExpectedResult) *Aggregator {
	aggr := &Aggregator{
		OutputDir:       outputDir,
		Results:         make(map[string]*plugin.Result, len(expected)),
		ExpectedResults: make(map[string]*plugin.ExpectedResult, len(expected)),
		resultEvents:    make(chan *plugin.Result, len(expected)),
	}

	for i, expResult := range expected {
		aggr.ExpectedResults[expResult.ID()] = &expected[i]
	}

	return aggr
}

// Wait blocks until all expected results have come in.
func (a *Aggregator) Wait(stop chan bool) {
	for !a.isComplete() {
		select {
		case <-a.resultEvents:
			break
		case <-stop:
			return
		}
	}
}

// isComplete returns true if sure all expected results have checked in.
func (a *Aggregator) isComplete() bool {
	a.resultsMutex.Lock()
	defer a.resultsMutex.Unlock()

	for _, result := range a.ExpectedResults {
		if _, ok := a.Results[result.ID()]; !ok {
			return false
		}
	}

	return true
}

func (a *Aggregator) isResultExpected(result *plugin.Result) bool {
	_, ok := a.ExpectedResults[result.ExpectedResultID()]
	return ok
}

func (a *Aggregator) isResultDuplicate(result *plugin.Result) bool {
	_, ok := a.Results[result.ExpectedResultID()]
	return ok
}

// HandleHTTPResult is called every time the HTTP server gets a well-formed
// request with results. This method is responsible for returning with things
// like a 409 conflict if a node has checked in twice (or a 403 forbidden if a
// node isn't expected), as well as actually calling handleResult to write the
// results to OutputDir.
func (a *Aggregator) HandleHTTPResult(result *plugin.Result, w http.ResponseWriter) {
	a.resultsMutex.Lock()
	defer a.resultsMutex.Unlock()

	resultID := result.ExpectedResultID()

	// Make sure we were expecting this result
	if !a.isResultExpected(result) {
		glog.Warningf("Got unexpected result %v", resultID)
		http.Error(
			w,
			fmt.Sprintf("Result %v unexpected", resultID),
			http.StatusForbidden,
		)
		return
	}

	// Don't allow duplicates
	if a.isResultDuplicate(result) {
		glog.Warningf("Got a duplicate result %v", resultID)
		http.Error(
			w,
			fmt.Sprintf("Result %v already received", resultID),
			http.StatusConflict,
		)
		return
	}

	if err := a.handleResult(result); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Error handling result %v: %v", resultID, err),
			http.StatusInternalServerError,
		)
		return
	}
}

// IngestResults takes a channel of results and handles them as they come in.
// Since most plugins submit over HTTP, this method is currently only used to
// consume an error stream from each plugin's Monitor() function.
//
// If we support plugins that are just simple commands that the sonobuoy master
// runs, those plugins can submit results through the same channel.
func (a *Aggregator) IngestResults(resultsCh <-chan *plugin.Result) {
	for {
		result, more := <-resultsCh
		if !more {
			break
		}
		// Don't consume results we're not expecting, unless they're
		// errors (see below.)
		if !a.isResultExpected(result) {
			glog.Warningf("Result unexpected: %v", result)
			continue
		}

		func() {
			a.resultsMutex.Lock()
			defer a.resultsMutex.Unlock()

			// Don't consume results we've already seen
			if a.isResultDuplicate(result) {
				glog.Warningf("Duplicate result: %v", result)
				return
			}

			a.handleResult(result)
		}()

	}
}

// handleResult takes a given plugin Result and writes it out to the
// filesystem, signaling to the resultEvents channel when complete.
func (a *Aggregator) handleResult(result *plugin.Result) error {
	// Send an event that we got this result even if we get an error, so
	// that Wait() doesn't hang forever on problems.
	defer func() {
		a.Results[result.ExpectedResultID()] = result
		a.resultEvents <- result
	}()

	// Create the output directory for the result.  Will be of the
	// form .../plugins/:results_type/:node.json (for DaemonSet plugins) or
	// .../plugins/:results_type.json (for Job plugins)
	resultsFile := path.Join(a.OutputDir, result.Path()+result.Extension)
	resultsDir := path.Dir(resultsFile)
	glog.Infof("Creating directory %v", resultsDir)
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		glog.Errorf("Could not make directory %v: %v", resultsDir, err)
		return err
	}

	// Write the results file out and close it
	err := func() error {
		f, err := os.Create(resultsFile)
		if err != nil {
			glog.Errorf("Could not open output file %v for writing: %v", resultsFile, err)
			return err
		}
		defer f.Close()

		// Copy the request body into the file
		_, err = io.Copy(f, result.Body)
		if err != nil {
			glog.Errorf("Error writing plugin result: %v", err)
			return err
		}

		return nil
	}()
	if err != nil {
		return err
	}

	// If it's a tarball, extract it
	if result.Extension == ".tar.gz" {
		resultsDir := path.Join(a.OutputDir, result.Path())

		err = tarx.Extract(resultsFile, resultsDir, &tarx.ExtractOptions{})
		if err != nil {
			glog.Errorf("Could not extract tar file %v: %v", resultsFile, err)
			return err
		}

		err = os.Remove(resultsFile)
		if err != nil {
			return err
		}

		glog.Infof("extracted results tarball into %v", resultsDir)
	} else {
		glog.Infof("wrote results to %v", resultsFile)
	}

	return nil
}
