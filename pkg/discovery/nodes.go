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

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeDC is the
type NodeDC struct {
	collectBasicNodeData bool
	// TODO: Add other params that we will switch on.
}

// CollectNodeData will call out to the api-server and collect node data
func CollectNodeData(kubeClient kubernetes.Interface, ndc *NodeDC) error {
	var err error
	if ndc.collectBasicNodeData {
		glog.Info("Collecting Node Data...")
		nodelist, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
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
	return err
}
