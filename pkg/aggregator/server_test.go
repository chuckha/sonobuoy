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
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestStart(t *testing.T) {
	checkins := make(map[string]*nodeCheckin, 0)

	expectedNode := "testnode"
	expectedJSON := `{"some": "json"}`

	tmpdir, err := ioutil.TempDir("", "sonobuoy_server_test")
	if err != nil {
		t.Fatal("Could not create temp directory")
		t.FailNow()
		return
	}
	defer os.RemoveAll(tmpdir)

	server := &server{
		BindAddr: "127.0.0.1:" + strconv.Itoa(testPort),
		NodeCallback: func(checkin *nodeCheckin, w http.ResponseWriter) {
			// Just take note of what we've received
			checkins[checkin.NodeName] = checkin
		},
	}

	// Use buffered channels for simplicity so we don't have to have every last
	// thing be async
	stop := make(chan bool, 1)
	ready := make(chan bool)
	done := make(chan error, 1)

	go func() {
		done <- server.Start(stop, ready)
	}()
	<-ready

	// Expect a 404 and no results
	response := doRequest(t, "PUT", "/not/found", expectedJSON)
	if response.StatusCode != 404 {
		t.Fatalf("Expected a 404 response, got %v", response.StatusCode)
		t.Fail()
	}
	if len(checkins) > 0 {
		t.Fatalf("Request to a wrong URL should not have resulted in a node check-in")
		t.Fail()
	}

	// PUT is all that is accepted
	response = doRequest(t, "POST", "/api/v1/results/by-node/"+expectedNode+"/ansible", expectedJSON)
	if response.StatusCode != 405 {
		t.Fatalf("Expected a 405 response, got %v", response.StatusCode)
		t.Fail()
	}
	if len(checkins) > 0 {
		t.Fatalf("Request with wrong HTTP method should not have resulted in a node check-in")
		t.Fail()
	}

	// Happy path
	response = doRequest(t, "PUT", "/api/v1/results/by-node/"+expectedNode+"/ansible", expectedJSON)
	if response.StatusCode != 200 {
		t.Fatalf("Client got non-200 status from server: %v", response.StatusCode)
		t.Fail()
	}

	if _, ok := checkins[expectedNode]; !ok {
		t.Fatalf("Valid request for %v did not get recorded", expectedNode)
		t.Fail()
	}

	stop <- true
	err = <-done
	if err != nil {
		t.Fatalf("Server returned error: %v", err)
		t.Fail()
	}
}

var testPort = 8099

func doRequest(t *testing.T, method, path, body string) *http.Response {
	// Make a new HTTP transport for every request, this avoids issues where HTTP
	// connection keep-alive leaves connections running to old server instances.
	// (We can take the performance hit since it's just tests.)
	trn := &http.Transport{DisableKeepAlives: true}
	client := &http.Client{Transport: trn}
	resultsURL := url.URL{
		Scheme: "http",
		Host:   "0.0.0.0:" + strconv.Itoa(testPort),
		Path:   path,
	}

	req, err := http.NewRequest(
		method,
		resultsURL.String(),
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("error constructing request: %v", err)
		t.Fail()
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("error performing request: %v", err)
		t.Fail()
	}
	return resp
}

func ensureNoResults(t *testing.T, results chan *nodeCheckin) {
	select {
	case r := <-results:
		t.Fatalf("Result found when none should be present: %v", r)
		t.Fail()
		break
	default:
		break
	}
}
