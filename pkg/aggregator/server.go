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
	"net"
	"net/http"
	"strings"

	"github.com/golang/glog"
)

// server is a net/http server that can handle API requests for aggregation of
// results from nodes, sending them back over the Results channel
type server struct {
	// BindAddr is the address for the HTTP server to listen on, eg. 0.0.0.0:8080
	BindAddr string
	// NodeCallback is the function that is called when a node checks in.  It returns a boolean of whether to keep serving data.
	NodeCallback func(*nodeCheckin, http.ResponseWriter)
}

// nodeCheckin represents an event when a node checks in with results.
type nodeCheckin struct {
	// NodeName is the name of the node
	NodeName string
	// ResultsType is the type of results, (eg. ansible)
	ResultsType string
	// Body is the io object containing the results for this node
	Body io.Reader
	// Path is where the results are, once they have been written
	Path string
}

const (
	// we're using /api/v1 right now but aren't doing anything intelligent, if we
	// have an /api/v2 later we'll figure out a good strategy for splitting up the
	// handling.

	// resultsByNode is the HTTP path under which node results are PUT
	resultsByNode = "/api/v1/results/by-node/"
)

// Start starts this HTTP server, binding it to s.BindAddr and sending results
// over the s.Results channel. The first argument is the stop channel, which
// when written to will stop the server and close the HTTP socket. The second
// argument is the "ready" channel, which this function will write to once the
// HTTP server is ready for connections.
func (s *server) Start(stop <-chan bool, ready chan<- bool) error {
	mux := http.NewServeMux()
	mux.Handle("/", http.NotFoundHandler())

	// We handle /api/v1/results/by-node here, but we strip the prefix so that the
	// handling function only has to do some simple string splitting to get the node name.
	// An example call may look like `PUT /api/v1/results/by-node/node1/ansible`,
	// which indicates that these are ansible results for node1.
	mux.Handle(resultsByNode, http.StripPrefix(resultsByNode, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.nodeResultsHandler(w, r)
	})))
	server := &http.Server{
		Addr:    s.BindAddr,
		Handler: mux,
	}

	l, err := net.Listen("tcp", s.BindAddr)

	if err != nil {
		return fmt.Errorf("could not listen on %v: %v", s.BindAddr, err)
	}
	defer l.Close()

	glog.Infof("Listening for incoming results on %v\n", s.BindAddr)

	done := make(chan error)
	go func() {
		done <- server.Serve(l)
	}()
	// Let clients know they can send requests now (it's not perfect since the
	// above goroutine may not schedule right away, but it's the best we can do
	// and helps in testing.)
	ready <- true

	select {
	case <-stop:
		// Calling l.Close should make the http.Serve() above return
		l.Close()
		<-done
	case err = <-done:
		break
	}

	return err
}

// Handle requests to post results by node. Path must be stripped of the
// /api/v1/results/by-node prefix, leaving just :nodename/:type. The only
// supported method is PUT, this does not support reading existing data.
// Example: PUT node1.cluster.local/ansible
func (s *server) nodeResultsHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path, "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	// We accept PUT because the client is specifying the resource identifier via
	// the HTTP path. (As opposed to POST, where typically the clients would post
	// to a base URL and the server picks the final resource path.)
	if r.Method != http.MethodPut {
		http.Error(
			w,
			fmt.Sprintf("Unsupported method %s.  Supported methods are: %v", r.Method, http.MethodPut),
			http.StatusMethodNotAllowed,
		)
		return
	}

	// Parse the path into the node name and the type
	node, resultType := parts[0], parts[1]

	glog.Infof("got %v result from %v\n", resultType, node)

	result := &nodeCheckin{
		NodeName:    node,
		ResultsType: resultType,
		Body:        r.Body,
	}

	// Trigger our callback with this checkin record (which should write the file
	// out.) The callback is responsible for doing a 409 conflict if results are
	// given twice for the same node, etc.
	s.NodeCallback(result, w)
	r.Body.Close()
}

// convenience method to send when something went wrong
func serverError(w http.ResponseWriter) {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}