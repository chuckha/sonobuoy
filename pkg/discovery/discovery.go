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
	"sync"

	"github.com/golang/glog"
	//"github.com/spf13/viper"

	"k8s.io/client-go/kubernetes"
)

// DiscoveryConfig
type DiscoveryConfig struct {
	nodeConfig *NodeDC
	// TODO: add other config options
	// from other analysis tools
}

// Run is the main entrypoint for discovery
func Run(kubeClient kubernetes.Interface, stopCh <-chan struct{}) error {
	var wg sync.WaitGroup
	var err error
	done := make(chan struct{})

	dc := LoadDiscoveryConfig()
	// Update as we add more tools
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: Need to resolve the many error returns
		err = CollectNodeData(kubeClient, dc.nodeConfig)
	}()

	// TODO: Here is where we add in the other collectors
	// masters
	// addons
	// providers
	// workloads
	// e2es

	// Only exists to have a signal when all tools are done.
	go func() {
		wg.Wait()
		close(done)
	}()

	// block until completion or kill signal
	select {
	case <-stopCh:
	case <-done:
	}

	return err
}

// LoadDiscoveryConfig unmarshals the viper config
func LoadDiscoveryConfig() *DiscoveryConfig {
	glog.Infof("Loading Config...")
	dc := &DiscoveryConfig{
		nodeConfig: &NodeDC{
			collectBasicNodeData: true,
		},
	}

	// TODO: Need to resolve the viper config
	return dc
}
