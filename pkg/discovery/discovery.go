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
	"os"
	"regexp"
	"sync"

	"github.com/golang/glog"
	//"github.com/spf13/viper"

	"github.com/satori/go.uuid"

	"io/ioutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DiscoveryConfig
type DiscoveryConfig struct {
	// UUID string to identify a test run.
	UUID string `json:"UUID"`

	// Location to store the output results
	resultsDir string `json:"resultsDir"`

	// regex filters for e2es
	runtests       bool   `json:"runtests"`
	testFocusRegex string `json:"testFocusRegex"`
	testSkipRegex  string `json:"testSkipRegex"`

	// regex to define namespace collection default=*
	namespaces string `json:"namespaces"`

	// Namespace scoped
	// certificatesigningrequests bool
	// clusters				bool
	clusterrolebindings      bool `json:"clusterrolebindings"`
	clusterroles             bool `json:"clusterroles"`
	componentstatuses        bool `json:"componentstatuses"`
	configmaps               bool `json:"configmaps"`
	daemonsets               bool `json:"daemonsets"`
	deployments              bool `json:"deployments"`
	endpoints                bool `json:"endpoints"`
	events                   bool `json:"events"`
	horizontalpodautoscalers bool `json:"horizontalpodautoscalers"`
	ingresses                bool `json:"ingresses"`
	jobs                     bool `json:"jobs"`
	limitranges              bool `json:"limitranges"`
	networkpolicies          bool `json:"networkpolicies"`
	persistentvolumeclaims   bool `json:"persistentvolumeclaims"`
	persistentvolumes        bool `json:"persistentvolumes"`
	pods                     bool `json:"pods"`
	poddisruptionbudgets     bool `json:"poddisruptionbudgets"`
	podsecuritypolicies      bool `json:"podsecuritypolicies"`
	podtemplates             bool `json:"podtemplates"`
	replicasets              bool `json:"replicasets"`
	replicationcontrollers   bool `json:"replicationcontrollers"`
	resourcequotas           bool `json:"resourcequotas"`
	rolebindings             bool `json:"rolebindings"`
	roles                    bool `json:"roles"`
	secrets                  bool `json:"secrets"`
	serviceaccounts          bool `json:"serviceaccounts"`
	services                 bool `json:"services"`
	statefulsets             bool `json:"statefulsets"`
	storageclasses           bool `json:"storageclasses"`
	thirdpartyresources      bool `json:"thirdpartyresources"`

	// Non-NSScoped.
	nodes bool

	// NOTE: there is more data, but we need to make an initial pass.
	// + we will need to evaluate for every release cycle.
}

// FilterNamespaces filter the list of namespaces according to the filter string
// TODO: Push this to a utils function
func FilterNamespaces(kubeClient kubernetes.Interface, filter string) []string {
	var validns []string
	re := regexp.MustCompile(filter)
	nslist, err := kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err == nil {
		for _, ns := range nslist.Items {
			glog.V(5).Infof("Namespace %v Matched=%v", ns.Name, re.MatchString(ns.Name))
			if re.MatchString(ns.Name) {
				validns = append(validns, ns.Name)
			}
		}
	} else {
		panic(err.Error())
	}

	return validns
}

// Run is the main entrypoint for discovery
func Run(kubeClient kubernetes.Interface, stopCh <-chan struct{}) []error {
	var wg sync.WaitGroup
	var m sync.Mutex
	var errlst []error
	done := make(chan struct{})

	// 0. Load the config
	dc := LoadDiscoveryConfig()

	// 1. Get the list of namespaces and apply the regex filter on the namespace
	nslist := FilterNamespaces(kubeClient, dc.namespaces)

	// 2. Create the directory which wil store the results
	outpath := dc.resultsDir + "/" + dc.UUID
	err := os.MkdirAll(outpath, 0755)
	if err != nil {
		panic(err.Error())
	}

	// 3. Dump the .json we used to run our test
	if blob, err := json.Marshal(dc); err == nil {
		if err = ioutil.WriteFile(outpath+"/config.json", blob, 0644); err != nil {
			panic(err.Error())
		}
	}

	// 4. Launch queries concurrently
	wg.Add(len(nslist) + 2)
	spawn := func(err error) {
		defer wg.Done()
		defer m.Unlock()
		m.Lock()
		if err != nil {
			errlst = append(errlst, err)
		}
	}
	waitcomplete := func() {
		wg.Wait()
		close(done)
	}
	go spawn(QueryNonNSResources(kubeClient, outpath, dc))
	for _, ns := range nslist {
		go spawn(QueryNSResources(kubeClient, outpath, ns, dc))
	}
	go spawn(rune2e(outpath, dc))
	go waitcomplete()

	// block until completion or kill signal
	select {
	case <-stopCh:
	case <-done:
	}

	return errlst
}

// LoadDiscoveryConfig unmarshals the viper config
func LoadDiscoveryConfig() *DiscoveryConfig {
	glog.Infof("Loading Config...")
	dc := &DiscoveryConfig{
		UUID:                uuid.NewV4().String(),
		resultsDir:          "./results",
		runtests:            true,
		testFocusRegex:      "Conformance",
		testSkipRegex:       "Serial|Alpha",
		namespaces:          ".*",
		clusterrolebindings: true,
		clusterroles:        true,
		componentstatuses:   true,
		configmaps:          true,
		daemonsets:          true,
		deployments:         true,
		endpoints:           true,
		events:              true,
		horizontalpodautoscalers: true,
		ingresses:                true,
		jobs:                     true,
		limitranges:              true,
		networkpolicies:          true,
		nodes:                    true,
		persistentvolumeclaims: true,
		persistentvolumes:      true,
		pods:                   true,
		poddisruptionbudgets:   true,
		podsecuritypolicies:    true,
		podtemplates:           true,
		replicasets:            true,
		replicationcontrollers: true,
		resourcequotas:         true,
		rolebindings:           true,
		roles:                  true,
		secrets:                false,
		serviceaccounts:        true,
		services:               true,
		statefulsets:           true,
		storageclasses:         true,
		thirdpartyresources:    true,
	}

	// TODO: Need to resolve the viper config
	return dc
}
