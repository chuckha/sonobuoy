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
	"path"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/aggregator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

func gatherHostData(client kubernetes.Interface, dc *Config) error {
	// TODO: there are other places that iterate through the CoreV1.Nodes API
	// call, we should only do this in one place and cache it.
	nodelist, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// TODO: make this a little more DRY
	var resultTypes []string
	if dc.HostFacts {
		resultTypes = append(resultTypes, "ansible")
	}
	if dc.HostLogs {
		resultTypes = append(resultTypes, "systemd_logs")
	}

	hosts := make(map[string]string, len(nodelist.Items))
	nodeNames := make([]string, len(nodelist.Items))
	for i, node := range nodelist.Items {
		addrs := node.Status.Addresses
		var addr string

		if len(addrs) < 1 {
			// sanity check
			continue
		}

		// We prefer the internal IP of each node
		for _, a := range addrs {
			if a.Type == v1.NodeInternalIP {
				addr = a.Address
			}
		}

		// Pick the first one as a fallback
		if addr == "" {
			glog.Warningf("Could not determine internal address for %v, defaulting to first adddress found (%v)\n", node.Name, addrs[0].Address)
			addr = addrs[0].Address
		}

		nodeNames[i] = node.Name
		hosts[node.Name] = addr
	}

	if len(nodeNames) == 0 || len(resultTypes) == 0 {
		glog.Warningf("Skipping host data gathering: no data to gather (%n hosts, %n result types)", len(nodeNames), len(resultTypes))
		return nil
	}

	aggr := &aggregator.NodeAggregator{
		BindAddr:          dc.AggregationBindAddress + ":" + strconv.Itoa(dc.AggregationBindPort),
		ExpectNodes:       nodeNames,
		ExpectResultTypes: resultTypes,
		OutputDir:         path.Join(dc.OutputDir(), "hosts"),
	}

	// Ensure we only wait for results for a certain time
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(dc.AggregationTimeoutSeconds) * time.Second)
		timeout <- true
	}()

	stop := make(chan bool)
	result := make(chan error)
	ready := make(chan bool, 1)
	done := make(chan bool, 1)
	go func() {
		result <- aggr.GatherAndAwaitResults(stop, ready, done)
	}()
	<-ready

	select {
	case err = <-result:
		break
	case <-done:
		stop <- true
		<-result
	case <-timeout:
		glog.Errorf("Timed out waiting for results, shutting down HTTP server\n")
		stop <- true
		<-result
	}

	return err
}
