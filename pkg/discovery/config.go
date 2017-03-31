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
	"flag"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// DiscoveryConfigKey is the default key used to
	DiscoveryConfigKey = "sonobuoy"
)

// Config is the input struct used to determine what data to collect
type Config struct {
	// UUID string to identify a test run.
	UUID string `json:"UUID"`

	// Location to store the output results
	ResultsDir string `json:"resultsdir"`

	// regex filters for e2es
	Runtests       bool   `json:"runtests"`
	TestFocusRegex string `json:"testfocusregex"`
	TestSkipRegex  string `json:"testskipregex"`

	// regex to define namespace collection default=*
	Namespaces string `json:"namespaces"`

	// Namespace scoped
	// certificatesigningrequests bool
	// clusters				bool
	// networkpolicies          bool `json:"networkpolicies"`
	ClusterRoleBindings      bool `json:"clusterrolebindings"`
	ClusterRoles             bool `json:"clusterroles"`
	ComponentStatuses        bool `json:"componentstatuses"`
	ConfigMaps               bool `json:"configmaps"`
	DaemonSets               bool `json:"daemonsets"`
	Deployments              bool `json:"deployments"`
	Endpoints                bool `json:"endpoints"`
	Events                   bool `json:"events"`
	HorizontalPodAutoscalers bool `json:"horizontalpodautoscalers"`
	Ingresses                bool `json:"ingresses"`
	Jobs                     bool `json:"jobs"`
	LimitRanges              bool `json:"limitranges"`
	PersistentVolumeClaims   bool `json:"persistentvolumeclaims"`
	Pods                     bool `json:"pods"`
	PodDisruptionBudgets     bool `json:"poddisruptionbudgets"`
	PodSecurityPolicies      bool `json:"podsecuritypolicies"`
	PodTemplates             bool `json:"podtemplates"`
	ReplicaSets              bool `json:"replicasets"`
	ReplicationControllers   bool `json:"replicationcontrollers"`
	ResourceQuotas           bool `json:"resourcequotas"`
	RoleBindings             bool `json:"rolebindings"`
	Roles                    bool `json:"roles"`
	Secrets                  bool `json:"secrets"`
	ServiceAccounts          bool `json:"serviceaccounts"`
	Services                 bool `json:"services"`
	StatefulSets             bool `json:"statefulsets"`
	StorageClasses           bool `json:"storageclasses"`
	ThirdPartyResources      bool `json:"thirdpartyresources"`

	// Non-NSScoped.
	PersistentVolumes bool `json:"persistentvolumes"`
	Nodes             bool `json:"nodes"`

	// TODOs:
	// 1. Convert to []string of resources vs. for easier processing.
	// 2. Master component /configz
	// 3. Add support for label selection? (Whitelist, Blacklist)
	// 4. Pod RegEx query? (workload issue)
	// 5. Other api-types.
}

// SetConfigDefaults sets up the defaults in case input is sparse.
func SetConfigDefaults(dc *Config) {
	dc.UUID = uuid.NewV4().String()
	dc.ResultsDir = "./results"
	dc.Runtests = false
	dc.TestFocusRegex = "Conformance"
	dc.TestSkipRegex = "Alpha|Disruptive|Feature|Flaky|Serial"
	dc.Namespaces = ".*"
	dc.ClusterRoleBindings = true
	dc.ClusterRoles = true
	dc.ComponentStatuses = true
	dc.ConfigMaps = true
	dc.DaemonSets = true
	dc.Deployments = true
	dc.Endpoints = true
	dc.Events = true
	dc.HorizontalPodAutoscalers = true
	dc.Ingresses = true
	dc.Jobs = true
	dc.LimitRanges = true
	dc.Nodes = true
	dc.PersistentVolumeClaims = true
	dc.PersistentVolumes = true
	dc.Pods = true
	dc.PodDisruptionBudgets = true
	dc.PodSecurityPolicies = true
	dc.PodTemplates = true
	dc.ReplicaSets = true
	dc.ReplicationControllers = true
	dc.ResourceQuotas = true
	dc.RoleBindings = true
	dc.Roles = true
	dc.Secrets = false
	dc.ServiceAccounts = true
	dc.Services = true
	dc.StatefulSets = true
	dc.StorageClasses = true
	dc.ThirdPartyResources = true
}

// LoadConfig will parse input + config file and return a clientset, and config
func LoadConfig() (kubernetes.Interface, *Config) {
	var config *rest.Config
	var err error
	var dc Config

	// 0 - load defaults
	flag.Parse()
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + DiscoveryConfigKey + "/")
	viper.AddConfigPath(".")
	viper.SetDefault("kubeconfig", "")
	viper.BindEnv("kubeconfig")
	SetConfigDefaults(&dc)

	// 1 - Read in the config file.
	if err = viper.ReadInConfig(); err != nil {
		panic(err.Error())
	}

	// 2 - Unmarshal the Config struct
	if err = viper.UnmarshalKey(DiscoveryConfigKey, &dc); err != nil {
		panic(err.Error())
	}

	// 3 - gather config information used to initialize
	kubeconfig := viper.GetString("kubeconfig")
	if len(kubeconfig) > 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	// 4 - creates the clientset from kubeconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, &dc
}
