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

	"k8s.io/client-go/kubernetes"
)

// Lister is something that can enumerate any array of results that can be
// dumped as json (so, any object really)
type Lister func() ([]interface{}, error)

func createresults(outpath string, file string, condition bool, f Lister) error {
	// Short-circuit early if we're not configured to gather these results
	if !condition {
		return nil
	}

	listObj, err := f()
	if err == nil && listObj != nil {
		if err = os.Mkdir(outpath, 0755); err == nil {
			if eJSONBytes, err := json.Marshal(listObj); err == nil {
				glog.V(5).Infof("%v", string(eJSONBytes))
				err = ioutil.WriteFile(outpath+"/"+file, eJSONBytes, 0644)
			}
		}
	}
	return err
}

// QueryNSResources writes out json files for namespaced resources in the cluster (pods, etc.)
func QueryNSResources(kubeClient kubernetes.Interface, outpath string, ns string, dc *DiscoveryConfig) error {
	glog.Infof("Running ns query (%v)", ns)
	return nil
}

// QueryNonNSResources writes out json files for non-namespaced resources in the cluster.
func QueryNonNSResources(kubeClient kubernetes.Interface, outpath string, dc *DiscoveryConfig) error {
	var err error
	glog.Infof("Running non-ns query")

	err = gatherNodeData(kubeClient, outpath, dc)
	return err
}
