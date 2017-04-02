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
	"sync"

	"github.com/viniciuschiele/tarx"
)

// Run is the main entrypoint for discovery
func Run(stopCh <-chan struct{}) []error {
	var wg sync.WaitGroup
	var m sync.Mutex
	var errlst []error
	done := make(chan struct{})

	// TODO - This will be in main and passed in.
	// 0. Load the config
	kubeClient, dc := LoadConfig()

	// 1. Get the list of namespaces and apply the regex filter on the namespace
	nslist := FilterNamespaces(kubeClient, dc.Namespaces)

	// 3. Create the directory which wil store the results
	outpath := dc.ResultsDir + "/" + dc.UUID
	err := os.MkdirAll(outpath+"/namespaces", 0755)
	if err != nil {
		panic(err.Error())
	}

	// 4. Dump the config.json we used to run our test
	if blob, err := json.Marshal(dc); err == nil {
		if err = ioutil.WriteFile(outpath+"/config.json", blob, 0644); err != nil {
			panic(err.Error())
		}
	}

	// TODO: Have consistency on error reporting.
	// Gather as many errors as possible get as far as we can, but only dump on success.
	// Errors should not be in-band.

	// 5. Launch queries concurrently
	wg.Add(len(nslist) + 2)
	spawn := func(err error) {
		defer wg.Done()
		defer m.Unlock()
		m.Lock()
		if err != nil {
			errlst = append(errlst, err)
		}
	}
	waitcomplete := func() {
		wg.Wait()
		close(done)
	}

	// TODO: Determine the level of parallelism we consider acceptable.
	// We can be throttled by the client to just let loose queries and channel back errors.
	go spawn(QueryNonNSResources(kubeClient, outpath, dc))
	for _, ns := range nslist {
		go spawn(QueryNSResources(kubeClient, outpath+"/namespaces", ns, dc))
	}
	go spawn(rune2e(outpath, dc))
	go waitcomplete()

	//6. Block until completion or kill signal
	select {
	case <-stopCh:
		// signal raised just exit
	case <-done:
		//7. tarball up results
		tb := dc.ResultsDir + "/sonobuoy_" + dc.UUID + ".tar.gz"
		err = tarx.Compress(tb, outpath, &tarx.CompressOptions{Compression: tarx.Gzip})
		if err == nil {
			err = os.RemoveAll(outpath)
		}
		if err != nil {
			errlst = append(errlst, err)
		}
	}

	return errlst
}
