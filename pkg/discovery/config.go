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
	"path"
	"reflect"
	"strings"

	"github.com/satori/go.uuid"
	"github.com/spf13/viper"

	"os"

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
	Description string `json:"description"`
	UUID        string `json:"UUID"`
	Version     string `json:"version"`

	// Location to store the output results
	ResultsDir string `json:"resultsdir"`

	// Configuration for ansible
	SSHRemoteUser string `json:"sshremoteuser"`

	// regex filters for e2es
	Runtests       bool   `json:"runtests"`
	TestFocusRegex string `json:"testfocusregex"`
	TestSkipRegex  string `json:"testskipregex"`
	Provider       string `json:"provider"`

	// regex to define namespace collection default=*
	Namespaces string `json:"namespaces"`
	// LabelSelector allows a selector string to selectively prune namespaced results
	// this is for use
	LabelSelector  string `json:"labelselector"`
	CollectPodLogs bool   `json:"collectpodlogs"`

	// Namespace scoped
	ConfigMaps               bool `json:"configmaps" resource:"ns"`
	CronJobs                 bool `json:"cronjobs" resource:"ns"`
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
	PodPresets               bool `json:"podpresets" resource:"ns"`
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
	// networkpolicies          bool `json:"networkpolicies"`

	// Non-NSScoped.
	CertificateSigningRequests bool `json:"certificatesigningrequests" resource:"non-ns"`
	ClusterRoleBindings        bool `json:"clusterrolebindings" resource:"non-ns"`
	ClusterRoles               bool `json:"clusterroles" resource:"non-ns"`
	ComponentStatuses          bool `json:"componentstatuses" resource:"non-ns"`
	Nodes                      bool `json:"nodes" resource:"non-ns"`
	PersistentVolumes          bool `json:"persistentvolumes" resource:"non-ns"`
	PodSecurityPolicies        bool `json:"podsecuritypolicies" resource:"non-ns"`
	StorageClasses             bool `json:"storageclasses" resource:"non-ns"`
	ThirdPartyResources        bool `json:"thirdpartyresources" resource:"non-ns"`

	// Other properties
	HostFacts     bool `json:"hostfacts"`
	ServerVersion bool `json:"serverversion"`

	// Non-serialized used for internal passing.
	kubeconfig string

	// TODOs:
	// - Master component /configz (Still unsupported atm)
	// - Other api-types?
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
	dc.SSHRemoteUser = "root"
	dc.Runtests = false
	dc.Provider = "local"
	dc.TestFocusRegex = "Conformance"
	dc.TestSkipRegex = "Alpha|Disruptive|Feature|Flaky|Kubectl"
	dc.Namespaces = ".*"
	dc.CertificateSigningRequests = true
	dc.ClusterRoleBindings = true
	dc.ClusterRoles = true
	dc.CollectPodLogs = false
	dc.ComponentStatuses = true
	dc.ConfigMaps = true
	dc.CronJobs = false
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
	dc.PodPresets = true
	dc.PodSecurityPolicies = true
	dc.PodTemplates = true
	dc.ReplicaSets = true
	dc.ReplicationControllers = true
	dc.ResourceQuotas = true
	dc.RoleBindings = true
	dc.Roles = true
	dc.Secrets = false
	dc.ServerVersion = true
	dc.ServiceAccounts = true
	dc.Services = true
	dc.StatefulSets = true
	dc.StorageClasses = true
	dc.ThirdPartyResources = true
	dc.HostFacts = false
}

// OutputDir returns the directory under the ResultsDir containing the
// UUID for this run.
func (dc *Config) OutputDir() string {
	return path.Join(dc.ResultsDir, dc.UUID)
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
	// Allow specifying a custom config file via the SONOBUOY_CONFIG env var
	if forceCfg := os.Getenv("SONOBUOY_CONFIG"); forceCfg != "" {
		viper.SetConfigFile(forceCfg)
	}

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
		dc.kubeconfig = kubeconfig
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
