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
	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/ansible"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

func ansibleConfig(client kubernetes.Interface, outpath string, dc *Config) (*ansible.Config, error) {
	nodelist, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hostnames := make([]string, len(nodelist.Items))
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

		hostnames[i] = addr
	}
	newcfg := ansible.Config{
		OutputDir:  outpath,
		Hosts:      hostnames,
		RemoteUser: dc.SshRemoteUser,
	}
	return &newcfg, err
}

func gatherHostFacts(client kubernetes.Interface, outpath string, dc *Config) error {
	acfg, err := ansibleConfig(client, outpath, dc)
	if err != nil {
		return err
	}

	return ansible.GatherHostData(acfg)
}
