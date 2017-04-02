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
	"fmt"
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

func listquery(outpath string, file string, f Lister) error {
	listObj, err := f()
	if err != nil {
		return err
	}
	if listObj == nil {
		return fmt.Errorf("got invalid response from API server")
	}
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
	return err
}

func queryNsResource(ns string, resourceKind string, kubeClient kubernetes.Interface) (runtime.Object, error) {
	switch resourceKind {
	case "configmaps":
		return kubeClient.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
	case "daemonsets":
		return kubeClient.Extensions().DaemonSets(ns).List(metav1.ListOptions{})
	case "deployments":
		return kubeClient.Apps().Deployments(ns).List(metav1.ListOptions{})
	case "endpoints":
		return kubeClient.CoreV1().Endpoints(ns).List(metav1.ListOptions{})
	case "events":
		return kubeClient.CoreV1().Events(ns).List(metav1.ListOptions{})
	case "horizontalpodautoscalers":
		return kubeClient.Autoscaling().HorizontalPodAutoscalers(ns).List(metav1.ListOptions{})
	case "ingresses":
		return kubeClient.Extensions().Ingresses(ns).List(metav1.ListOptions{})
	case "jobs":
		return kubeClient.Batch().Jobs(ns).List(metav1.ListOptions{})
	case "limitranges":
		return kubeClient.CoreV1().LimitRanges(ns).List(metav1.ListOptions{})
	case "persistentvolumeclaims":
		return kubeClient.CoreV1().PersistentVolumeClaims(ns).List(metav1.ListOptions{})
	case "pods":
		return kubeClient.CoreV1().Pods(ns).List(metav1.ListOptions{})
	case "poddisruptionbudgets":
		return kubeClient.Policy().PodDisruptionBudgets(ns).List(metav1.ListOptions{})
	case "podtemplates":
		return kubeClient.CoreV1().PodTemplates(ns).List(metav1.ListOptions{})
	case "replicasets":
		return kubeClient.Extensions().ReplicaSets(ns).List(metav1.ListOptions{})
	case "replicationcontrollers":
		return kubeClient.CoreV1().ReplicationControllers(ns).List(metav1.ListOptions{})
	case "resourcequotas":
		return kubeClient.CoreV1().ResourceQuotas(ns).List(metav1.ListOptions{})
	case "rolebindings":
		return kubeClient.Rbac().RoleBindings(ns).List(metav1.ListOptions{})
	case "roles":
		return kubeClient.Rbac().Roles(ns).List(metav1.ListOptions{})
	case "secrets":
		return kubeClient.CoreV1().Secrets(ns).List(metav1.ListOptions{})
	case "serviceaccounts":
		return kubeClient.CoreV1().ServiceAccounts(ns).List(metav1.ListOptions{})
	case "services":
		return kubeClient.CoreV1().Services(ns).List(metav1.ListOptions{})
	case "statefulsets":
		return kubeClient.Apps().StatefulSets(ns).List(metav1.ListOptions{})
	default:
		return nil, fmt.Errorf("don't know how to handle namespaced resource %v", resourceKind)
	}

}

func queryNonNsResource(resourceKind string, kubeClient kubernetes.Interface) (runtime.Object, error) {
	switch resourceKind {
	case "clusterrolebindings":
		return kubeClient.Rbac().ClusterRoleBindings().List(metav1.ListOptions{})
	case "clusterroles":
		return kubeClient.Rbac().ClusterRoles().List(metav1.ListOptions{})
	case "componentstatuses":
		return kubeClient.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
	case "nodes":
		return kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	case "persistentvolumes":
		return kubeClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	case "podsecuritypolicies":
		return kubeClient.Extensions().PodSecurityPolicies().List(metav1.ListOptions{})
	case "storageclasses":
		return kubeClient.Storage().StorageClasses().List(metav1.ListOptions{})
	case "thirdpartyresources":
		return kubeClient.Extensions().ThirdPartyResources().List(metav1.ListOptions{})
	default:
		return nil, fmt.Errorf("don't know how to handle non-namespaced resource %v", resourceKind)
	}
}

// QueryNSResources will query namespace specific
func QueryNSResources(kubeClient kubernetes.Interface, outpath string, ns string, dc *Config) error {
	var err error
	glog.Infof("Running ns query (%v)", ns)

	outdir := outpath + "/" + ns
	err = os.Mkdir(outdir, 0755)
	if err != nil {
		return fmt.Errorf("could not create output directory %s: %v", outdir, err)
	}

	for resourceKind, resourceScope := range dc.ResourcesToQuery() {
		// We use annotations to tag resources as being namespaced vs not, skip any
		// that aren't "ns"
		if resourceScope == "ns" {
			lister := func() (runtime.Object, error) { return queryNsResource(ns, resourceKind, kubeClient) }
			err = listquery(outdir+"/"+resourceKind, resourceKind+".json", lister)
			if err != nil {
				return fmt.Errorf("Error querying %v: %v", resourceKind, err)
			}
		}
	}

	return err
}

// QueryNonNSResources queries non-namespace aware components
func QueryNonNSResources(kubeClient kubernetes.Interface, outpath string, dc *Config) error {
	var err error
	glog.Infof("Running non-ns query")

	for resourceKind, resourceScope := range dc.ResourcesToQuery() {
		// We use annotations to tag resources as being namespaced vs not, skip any
		// that aren't "non-ns"
		if resourceScope == "non-ns" {
			lister := func() (runtime.Object, error) { return queryNonNsResource(resourceKind, kubeClient) }
			err = listquery(outpath+"/"+resourceKind, resourceKind+".json", lister)
			if err != nil {
				return fmt.Errorf("Error querying %v: %v", resourceKind, err)
			}
		}
	}

	if dc.Nodes {
		if err = gatherNodeData(kubeClient, outpath, dc); err != nil {
			return err
		}
	}
	return nil
}
