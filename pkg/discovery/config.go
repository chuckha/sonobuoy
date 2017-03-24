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
	"github.com/satori/go.uuid"
	//"github.com/spf13/viper"
)

// Config is the input struct used to determine what data to collect
type Config struct {
	// UUID string to identify a test run.
	UUID string `json:"UUID"`

	// Location to store the output results
	ResultsDir string `json:"resultsDir"`

	// regex filters for e2es
	Runtests       bool   `json:"runtests"`
	TestFocusRegex string `json:"testFocusRegex"`
	TestSkipRegex  string `json:"testSkipRegex"`

	// regex to define namespace collection default=*
	Namespaces string `json:"namespaces"`

	// Namespace scoped
	// certificatesigningrequests bool
	// clusters				bool
	// networkpolicies          bool `json:"networkpolicies"`
	Clusterrolebindings      bool `json:"clusterrolebindings"`
	Clusterroles             bool `json:"clusterroles"`
	Componentstatuses        bool `json:"componentstatuses"`
	Configmaps               bool `json:"configmaps"`
	Daemonsets               bool `json:"daemonsets"`
	Deployments              bool `json:"deployments"`
	Endpoints                bool `json:"endpoints"`
	Events                   bool `json:"events"`
	Horizontalpodautoscalers bool `json:"horizontalpodautoscalers"`
	Ingresses                bool `json:"ingresses"`
	Jobs                     bool `json:"jobs"`
	Limitranges              bool `json:"limitranges"`
	Persistentvolumeclaims   bool `json:"persistentvolumeclaims"`
	Pods                     bool `json:"pods"`
	Poddisruptionbudgets     bool `json:"poddisruptionbudgets"`
	Podsecuritypolicies      bool `json:"podsecuritypolicies"`
	Podtemplates             bool `json:"podtemplates"`
	Replicasets              bool `json:"replicasets"`
	Replicationcontrollers   bool `json:"replicationcontrollers"`
	Resourcequotas           bool `json:"resourcequotas"`
	Rolebindings             bool `json:"rolebindings"`
	Roles                    bool `json:"roles"`
	Secrets                  bool `json:"secrets"`
	Serviceaccounts          bool `json:"serviceaccounts"`
	Services                 bool `json:"services"`
	Statefulsets             bool `json:"statefulsets"`
	Storageclasses           bool `json:"storageclasses"`
	Thirdpartyresources      bool `json:"thirdpartyresources"`

	// Non-NSScoped.
	Persistentvolumes bool `json:"persistentvolumes"`
	Nodes             bool `json:"nodes"`

	// TODOs:
	// 1. Add support for label selection?
	// 2. Pod RegEx query? (workload issue)
	// 3. Other api-types.
	// 4. Pass in []string of resources vs. bool festivus?
}

// LoadConfig unmarshals the viper config
func LoadConfig() *Config {
	glog.Infof("Loading Config...")
	dc := &Config{
		UUID:                uuid.NewV4().String(),
		ResultsDir:          "./results",
		Runtests:            true,
		TestFocusRegex:      "Conformance",
		TestSkipRegex:       "Alpha|Disruptive|Feature|Flaky|Serial",
		Namespaces:          ".*",
		Clusterrolebindings: true,
		Clusterroles:        true,
		Componentstatuses:   true,
		Configmaps:          true,
		Daemonsets:          true,
		Deployments:         true,
		Endpoints:           true,
		Events:              true,
		Horizontalpodautoscalers: true,
		Ingresses:                true,
		Jobs:                     true,
		Limitranges:              true,
		Nodes:                    true,
		Persistentvolumeclaims: true,
		Persistentvolumes:      true,
		Pods:                   true,
		Poddisruptionbudgets:   true,
		Podsecuritypolicies:    true,
		Podtemplates:           true,
		Replicasets:            true,
		Replicationcontrollers: true,
		Resourcequotas:         true,
		Rolebindings:           true,
		Roles:                  true,
		Secrets:                false,
		Serviceaccounts:        true,
		Services:               true,
		Statefulsets:           true,
		Storageclasses:         true,
		Thirdpartyresources:    true,
	}

	// TODO: Need to resolve the viper config
	return dc
}
