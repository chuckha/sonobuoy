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
	"k8s.io/client-go/pkg/api/v1"
)

type nodeData struct {
	APIResource   v1.Node                `json:"apiResource,omitempty"`
	ConfigzOutput map[string]interface{} `json:"configzOutput,omitempty"`
	HealthzStatus int                    `json:"healthzStatus,omitempty"`
}

func gatherNodeData(kubeClient kubernetes.Interface, outpath string, dc *Config) error {
	f := func() ([]interface{}, error) {
		glog.Info("Collecting Node Data...")
		nodelist, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		results := make([]interface{}, len(nodelist.Items))
		for i, node := range nodelist.Items {
			var configz map[string]interface{}
			var healthstatus int

			// We hit the master on /api/v1/proxy/nodes/<node> to gather node
			// information without having to reinvent auth
			proxypath := "/api/v1/proxy/nodes/" + node.Name
			restclient := kubeClient.CoreV1().RESTClient()

			// Get the configz endpoint, put the result in the nodeData
			request := restclient.Get().RequestURI(proxypath + "/configz")
			if result, err := request.Do().Raw(); err == nil {
				json.Unmarshal(result, &configz)
			} else {
				glog.Warningf("Could not get configz endpoint for node %v: %v", node.Name, err)
			}

			// Get the healthz endpoint too. We care about the response code in this
			// case, not the body.
			request = restclient.Get().RequestURI(proxypath + "/healthz")
			if result := request.Do(); result.Error() == nil {
				result.StatusCode(&healthstatus)
			} else {
				glog.Warningf("Could not get healthz endpoint for node %v: %v", node.Name, result.Error())
			}

			results[i] = nodeData{
				APIResource:   node,
				ConfigzOutput: configz,
				HealthzStatus: healthstatus,
			}
		}

		return results, nil
	}

	return untypedListQuery(outpath+"/nodes", "nodes.json", f)
}
