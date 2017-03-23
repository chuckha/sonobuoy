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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// Lister is used for
type Lister func() (runtime.Object, error)

func createresults(outpath string, file string, test bool, err error, f Lister) error {
	if test && err == nil {
		listObj, err := f()
		if err == nil && listObj != nil {
			listPtr, err := meta.GetItemsPtr(listObj)
			if err == nil && listPtr != nil {
				if err = os.Mkdir(outpath, 0755); err == nil {
					if eJSONBytes, err := json.Marshal(listPtr); err == nil {
						glog.V(5).Infof("%v", string(eJSONBytes))
						err = ioutil.WriteFile(outpath+"/"+file, eJSONBytes, 0644)
					}
				}
			}
		}
	}
	return err
}

func QueryNSResources(kubeClient kubernetes.Interface, outpath string, ns string, dc *DiscoveryConfig) error {
	glog.Infof("Running ns query (%v)", ns)
	return nil
}

func QueryNonNSResources(kubeClient kubernetes.Interface, outpath string, dc *DiscoveryConfig) error {
	var err error

	glog.Infof("Running non-ns query")
	f := func() (runtime.Object, error) { return kubeClient.CoreV1().Nodes().List(metav1.ListOptions{}) }
	err = createresults(outpath+"/nodes", "nodes.json", dc.nodes, err, f)

	return err
}

/*
	if dc.nodes {
		glog.Info("Collecting Node Data...")
		nodelist, err :=
		if err == nil {
			for i, node := range nodelist.Items {
				// TODO: We'll need to add more analysis
				if eJSONBytes, err := json.Marshal(node); err == nil {
					// TODO: need to write output file
					glog.Infof("NODE(%v)\n%v", i, string(eJSONBytes))
				} else {
					glog.Warningf("Failed to json serialize node: %v", err)
				}
			}
		}
	}
	return nil
*/
