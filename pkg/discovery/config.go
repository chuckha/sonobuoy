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
	"reflect"
	"strings"

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
	// strings to identify a test run.
	UUID        string `json:"UUID"`
	Description string `json:"description"`
	Version     string `json:"version"`

	// Location to store the output results
	ResultsDir string `json:"resultsdir"`

	// Configuration for ansible
	SshRemoteUser string `json:"sshremoteuser"`

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
	ConfigMaps               bool `json:"configmaps" resource:"ns"`
	DaemonSets               bool `json:"daemonsets" resource:"ns"`
	Deployments              bool `json:"deployments" resource:"ns"`
	Endpoints                bool `json:"endpoints" resource:"ns"`
	Events                   bool `json:"events" resource:"ns"`
	HorizontalPodAutoscalers bool `json:"horizontalpodautoscalers" resource:"ns"`
	Ingresses                bool `json:"ingresses" resource:"ns"`
	Jobs                     bool `json:"jobs" resource:"ns"`
	LimitRanges              bool `json:"limitranges" resource:"ns"`
	PersistentVolumeClaims   bool `json:"persistentvolumeclaims" resource:"ns"`
	Pods                     bool `json:"pods" resource:"ns"`
	PodDisruptionBudgets     bool `json:"poddisruptionbudgets" resource:"ns"`
	PodTemplates             bool `json:"podtemplates" resource:"ns"`
	ReplicaSets              bool `json:"replicasets" resource:"ns"`
	ReplicationControllers   bool `json:"replicationcontrollers" resource:"ns"`
	ResourceQuotas           bool `json:"resourcequotas" resource:"ns"`
	RoleBindings             bool `json:"rolebindings" resource:"ns"`
	Roles                    bool `json:"roles" resource:"ns"`
	Secrets                  bool `json:"secrets" resource:"ns"`
	ServiceAccounts          bool `json:"serviceaccounts" resource:"ns"`
	Services                 bool `json:"services" resource:"ns"`
	StatefulSets             bool `json:"statefulsets" resource:"ns"`

	// Non-NSScoped.
	PersistentVolumes   bool `json:"persistentvolumes" resource:"non-ns"`
	Nodes               bool `json:"nodes" resource:"non-ns"`
	ComponentStatuses   bool `json:"componentstatuses" resource:"non-ns"`
	PodSecurityPolicies bool `json:"podsecuritypolicies" resource:"non-ns"`
	ClusterRoleBindings bool `json:"clusterrolebindings" resource:"non-ns"`
	ClusterRoles        bool `json:"clusterroles" resource:"non-ns"`
	ThirdPartyResources bool `json:"thirdpartyresources" resource:"non-ns"`
	StorageClasses      bool `json:"storageclasses" resource:"non-ns"`

	// Other behavior
	HostFacts bool `json:"hostfacts"`

	// TODOs:
	// 1. Master component /configz
	// 2. Add support for label selection? (Whitelist, Blacklist)
	// 3. Pod RegEx query? (workload issue)
	// 4. Other api-types.
}

// SonoCfg is used to export a config
type SonoCfg struct {
	DC Config `json:"sonobuoy,omitempty"`
}

// SetConfigDefaults sets up the defaults in case input is sparse.
func SetConfigDefaults(dc *Config) {
	dc.UUID = uuid.NewV4().String()
	dc.Description = "NONE"
	dc.ResultsDir = "./results"
	dc.SshRemoteUser = "root"
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
	dc.HostFacts = true

}

// ResourcesToQuery returns the list of NS and non-NS resource types that are
// ok to query, depending on whether they have been set to "true" in the
// configuration. The keys to the map will be the JSON key name from the config
// struct (eg. "configmaps"), and the values will be the resource type (eg.
// "ns" or "non-ns"). All entries in the map are ok to query, values set to
// false will not be included.
func (dc *Config) ResourcesToQuery() map[string]string {
	cfgtype := reflect.TypeOf(*dc)
	ret := make(map[string]string, cfgtype.NumField())

	// Use the reflect package to iterate on fields in the config struct, adding
	// them to the returned map if they have the "resource" field tag and are set
	// to `true`. The map will be indexed by the JSON annotation key (ie.
	// "configmaps").
	for i := 0; i < cfgtype.NumField(); i++ {
		resourceTagVal := cfgtype.Field(i).Tag.Get("resource")

		if resourceTagVal != "" {
			field := cfgtype.Field(i)
			jsontag := field.Tag.Get("json")
			// `key` is the JSON key
			key := strings.Split(jsontag, ",")[0]

			if key == "" {
				continue
			}

			cfgval := reflect.ValueOf(*dc).FieldByName(field.Name)

			// Only add it if the user wants to query it
			if cfgval.Kind() == reflect.Bool && cfgval.Bool() {
				ret[key] = resourceTagVal
			}
		}
	}

	return ret
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
