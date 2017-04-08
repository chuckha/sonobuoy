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
	"fmt"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/heptio/sonobuoy/pkg/ansible"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

func ansibleConfig(client kubernetes.Interface, dc *Config) (*ansible.Config, error) {
	nodelist, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]string, len(nodelist.Items))
	for _, node := range nodelist.Items {
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

		hosts[node.Name] = addr
	}

	outpath := path.Join(dc.OutputDir(), ".ansible")
	if err = os.MkdirAll(outpath, 0755); err != nil {
		return nil, err
	}

	newcfg := ansible.Config{
		OutputDir:  outpath,
		Hosts:      hosts,
		RemoteUser: dc.SSHRemoteUser,
	}
	return &newcfg, err
}

func moveAnsibleResults(acfg *ansible.Config, dc *Config) error {
	// Where ansible dropped the files for all hosts (eg. results/hosts/.ansible-results)
	sourcedir := path.Join(acfg.OutputDir, ansible.ResultsLocation)
	var err error

	for nodeName, addr := range acfg.Hosts {
		// eg. results/hosts/host1.local/
		hostdir := path.Join(dc.OutputDir(), HostsLocation, nodeName)
		dest := path.Join(hostdir, "ansible.json")
		if err = os.MkdirAll(hostdir, 0755); err != nil {
			return err
		}

		// Ansible names the results files after their address, not the node's
		// logical name. (eg. results/hosts/)
		ansibleResult := path.Join(sourcedir, addr)
		err = os.Rename(ansibleResult, dest)
		if err != nil {
			return err
		}
	}

	// There should really be no other contents in the ansible results directory,
	// we should be able to remove it safely.
	if err = os.Remove(sourcedir); err != nil {
		return fmt.Errorf("could not remove temporary ansible results directory: %v", err)
	}

	return nil
}

func gatherHostFacts(client kubernetes.Interface, dc *Config) error {
	acfg, err := ansibleConfig(client, dc)
	if err != nil {
		return err
	}

	if err = ansible.GatherHostData(acfg); err != nil {
		return err
	}

	if err = moveAnsibleResults(acfg, dc); err != nil {
		return err
	}

	return nil
}
