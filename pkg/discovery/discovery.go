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
	"github.com/spf13/viper"

	"k8s.io/client-go/kubernetes"
)

// Run is the main entrypoint for discovery
func Run(kubeClient kubernetes.Interface, stopCh <-chan struct{}) error {
	// TODO if-check input collection options
	glog.Info("Collecting Node Data...")
	err := CollectNodeData(kubeClient)
	return err
}

// SetDefaults sets the defaults for discovery
func SetDefaults() {
	// TODO - We will need to breakdown the hierarchical structure for data that we intent on collecting.
	viper.SetDefault("todo", "parameterize defaults")
}
