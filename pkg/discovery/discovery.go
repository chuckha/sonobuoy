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
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/agent"
	"github.com/heptio/sonobuoy/pkg/dispatch"
	"github.com/viniciuschiele/tarx"
)

// Run is the main entrypoint for discovery
func Run(version string) []error {
	var errlst []error

	// 0. Load the config
	kubeClient, dc := LoadConfig()
	dc.Version = version

	// 1. Get the list of namespaces and apply the regex filter on the namespace
	nslist := FilterNamespaces(kubeClient, dc.Namespaces)

	// 2. Create the directory which will store the results
	outpath := dc.ResultsDir + "/" + dc.UUID
	err := os.MkdirAll(outpath, 0755)
	if err != nil {
		panic(err.Error())
	}

	// 3. Dump the config.json we used to run our test
	if blob, err := json.Marshal(dc); err == nil {
		if err = ioutil.WriteFile(outpath+"/config.json", blob, 0644); err != nil {
			panic(err.Error())
		}
	}

	rollup := func(err []error) {
		if err != nil {
			errlst = append(errlst, err...)
		}
	}

	// 4. Start running the host query aggregator (it could take a while)
	hostFactsResult := make(chan error, 1)
	if dc.ShouldDispatchHostAgents() {
		go func() {
			agentCfg := &agent.Config{
				Ansible:      dc.HostFacts,
				SystemdLogs:  dc.HostLogs,
				ChrootDir:    "/node",
				PhoneHomeURL: "http://" + dc.AggregationAdvertiseAddress + "/api/v1/results/by-node",
			}

			agentDispatcher := dispatch.NewAgentDispatcher(agentCfg, kubeClient)
			defer agentDispatcher.Cleanup(0)

			if err = agentDispatcher.Dispatch(); err != nil {
				glog.Errorf("Error dispatching agent DaemonSet: %v\n", err)
				errlst = append(errlst, err)
				hostFactsResult <- nil
				return
			}

			hostFactsResult <- gatherHostData(kubeClient, dc)
		}()
	} else {
		// Put a nil result in the channel so reading it later won't block
		hostFactsResult <- nil
	}

	// 5. Run the queries
	rollup(QueryNonNSResources(kubeClient, dc))
	for _, ns := range nslist {
		rollup(QueryNSResources(kubeClient, ns, dc))
	}
	rollup(rune2e(dc))

	// 6. Wait for the host aggregator to finish
	if r := <-hostFactsResult; r != nil {
		errlst = append(errlst, r)
	}

	// 7. tarball up results
	tb := dc.ResultsDir + "/sonobuoy_" + dc.UUID + ".tar.gz"
	err = tarx.Compress(tb, outpath, &tarx.CompressOptions{Compression: tarx.Gzip})
	if err == nil {
		err = os.RemoveAll(outpath)
	}
	if err != nil {
		errlst = append(errlst, err)
	}

	return errlst
}
