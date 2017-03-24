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
	"io/ioutil"
	"os"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// TODO:
// 1. This will change quite a bit once we have conversion from bool <> string
// 2. Pass back errors through channel
// 3. map of name<>function.

// Lister is something that can enumerate any array of results that can be
// dumped as json (so, any object really)
type Lister func() (runtime.Object, error)

func listquery(outpath string, file string, test bool, err error, f Lister) error {
	// Short-circuit early if we're not configured to gather these results
	if test && err == nil {
		listObj, err := f()
		if err == nil && listObj != nil {
			if listPtr, err := meta.GetItemsPtr(listObj); err == nil {
				if items, err := conversion.EnforcePtr(listPtr); err == nil {
					if items.Len() > 0 {
						if err = os.Mkdir(outpath, 0755); err == nil {
							if eJSONBytes, err := json.Marshal(listPtr); err == nil {
								glog.V(5).Infof("%v", string(eJSONBytes))
								err = ioutil.WriteFile(outpath+"/"+file, eJSONBytes, 0644)
							}
						}
					}
				}
			}
		}
	}
	return err
}

// QueryNSResources will query namespace specific
func QueryNSResources(kubeClient kubernetes.Interface, outpath string, ns string, dc *Config) error {
	var err error
	glog.Infof("Running ns query (%v)", ns)

	outdir := outpath + "/" + ns
	err = os.Mkdir(outdir, 0755)

	// grab configmaps
	f := func() (runtime.Object, error) {
		return kubeClient.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/configmaps", "configmaps.json", dc.Configmaps, err, f)

	// grab daemonsets
	f = func() (runtime.Object, error) {
		return kubeClient.Extensions().DaemonSets(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/daemonsets", "daemonsets.json", dc.Daemonsets, err, f)

	// grab deployments
	f = func() (runtime.Object, error) {
		return kubeClient.Apps().Deployments(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/deployments", "deployments.json", dc.Deployments, err, f)

	// grab endpoints
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().Endpoints(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/endpoints", "endpoints.json", dc.Endpoints, err, f)

	// grab events
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().Events(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/events", "events.json", dc.Events, err, f)

	// grab horizontalpodautoscalers
	f = func() (runtime.Object, error) {
		return kubeClient.Autoscaling().HorizontalPodAutoscalers(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/horizontalpodautoscalers", "horizontalpodautoscalers.json", dc.Horizontalpodautoscalers, err, f)

	// grab ingresses
	f = func() (runtime.Object, error) {
		return kubeClient.Extensions().Ingresses(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/ingresses", "ingresses.json", dc.Ingresses, err, f)

	//grab jobs
	f = func() (runtime.Object, error) {
		return kubeClient.Batch().Jobs(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/jobs", "jobs.json", dc.Jobs, err, f)

	// grab limitranges
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().LimitRanges(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/limitranges", "limitranges.json", dc.Limitranges, err, f)

	/* grab networkpolicies
	f = func() (runtime.Object, error) {
		return kubeClient.Foo().Bar(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/networkpolicies", "networkpolicies.json", dc.networkpolicies, err, f)
	*/

	// grab persistentvolumeclaims
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().PersistentVolumeClaims(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/persistentvolumeclaims", "persistentvolumeclaims.json", dc.Persistentvolumeclaims, err, f)

	// grab pods
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().Pods(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/pods", "pods.json", dc.Pods, err, f)

	// grab poddisruptionbudgets
	f = func() (runtime.Object, error) {
		return kubeClient.Policy().PodDisruptionBudgets(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/poddisruptionbudgets", "poddisruptionbudgets.json", dc.Poddisruptionbudgets, err, f)

	// grab podtemplates
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().PodTemplates(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/podtemplates", "podtemplates.json", dc.Podtemplates, err, f)

	// grab replicasets
	f = func() (runtime.Object, error) {
		return kubeClient.Extensions().ReplicaSets(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/replicasets", "replicasets.json", dc.Replicasets, err, f)

	// grab replicationcontrollers
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().ReplicationControllers(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/replicationcontrollers", "replicationcontrollers.json", dc.Replicationcontrollers, err, f)

	// grab resourcequotas
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().ResourceQuotas(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/resourcequotas", "resourcequotas.json", dc.Resourcequotas, err, f)

	// grab rolebindings
	f = func() (runtime.Object, error) {
		return kubeClient.Rbac().RoleBindings(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/rolebindings", "rolebindings.json", dc.Rolebindings, err, f)

	// grab roles
	f = func() (runtime.Object, error) {
		return kubeClient.Rbac().Roles(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/roles", "roles.json", dc.Roles, err, f)

	// grab secrets
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().Secrets(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/secrets", "secrets.json", dc.Secrets, err, f)

	// grab serviceaccounts
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().ServiceAccounts(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/serviceaccounts", "serviceaccounts.json", dc.Serviceaccounts, err, f)

	// grab services
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().Secrets(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/services", "services.json", dc.Services, err, f)

	// grab statefulsets
	f = func() (runtime.Object, error) {
		return kubeClient.Apps().StatefulSets(ns).List(metav1.ListOptions{})
	}
	err = listquery(outdir+"/statefulsets", "statefulsets.json", dc.Deployments, err, f)

	return err
}

// QueryNonNSResources queries non-namespace aware components
func QueryNonNSResources(kubeClient kubernetes.Interface, outpath string, dc *Config) error {
	var err error
	glog.Infof("Running non-ns query")

	// grab clusterrolebindings
	f := func() (runtime.Object, error) {
		return kubeClient.Rbac().ClusterRoleBindings().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/clusterrolebindings", "clusterrolebindings.json", dc.Clusterrolebindings, err, f)

	// grab clusterroles
	f = func() (runtime.Object, error) {
		return kubeClient.Rbac().ClusterRoles().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/clusterroles", "clusterroles.json", dc.Clusterrolebindings, err, f)

	// grab componentstatus
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/componentstatuses", "componentstatuses.json", dc.Componentstatuses, err, f)

	// grab nodes
	if err == nil {
		err = gatherNodeData(kubeClient, outpath, dc)
	}
	/*f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/nodes", "nodes.json", dc.Nodes, err, f)*/

	// grab persistentvolumes
	f = func() (runtime.Object, error) {
		return kubeClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/persistentvolumes", "persistentvolumes.json", dc.Persistentvolumes, err, f)

	// grab podsecuritypolicies
	f = func() (runtime.Object, error) {
		return kubeClient.Extensions().PodSecurityPolicies().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/podsecuritypolicies", "podsecuritypolicies.json", dc.Podsecuritypolicies, err, f)

	// grab storageclasses
	f = func() (runtime.Object, error) {
		return kubeClient.Storage().StorageClasses().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/storageclasses", "storageclasses.json", dc.Storageclasses, err, f)

	// grab thirdpartyresources
	f = func() (runtime.Object, error) {
		return kubeClient.Extensions().ThirdPartyResources().List(metav1.ListOptions{})
	}
	err = listquery(outpath+"/thirdpartyresources", "thirdpartyresources.json", dc.Thirdpartyresources, err, f)

	return err
}
