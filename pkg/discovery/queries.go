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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// TODO:
// 1. Pass back errors through channel
// 2. map of name<>function. (debatable if we want to do this)

// ObjQuery is a query function that returns a kubernetes object
type ObjQuery func() (runtime.Object, error)

// UntypedQuery is a query function that return an untyped array of objs
type UntypedQuery func() (interface{}, error)

// UntypedListQuery is a query function that return an untyped array of objs
type UntypedListQuery func() ([]interface{}, error)

const (
	// NSResourceLocation is the place under which namespaced API resources (pods, etc) are stored
	NSResourceLocation = "resources/ns"
	// NonNSResourceLocation is the place under which non-namespaced API resources (nodes, etc) are stored
	NonNSResourceLocation = "resources/non-ns"
	// HostsLocation is the place under which host information (configz, ansible JSON, etc) is stored
	HostsLocation = "hosts"
)

func objListQuery(outpath string, file string, f ObjQuery) error {
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
				err = SerializeObj(listPtr, outpath, file)
			}
		}
	}
	return err
}

func untypedQuery(outpath string, file string, f UntypedQuery) error {
	Obj, err := f()
	if err == nil && Obj != nil {
		err = SerializeObj(Obj, outpath, file)
	}
	return err
}

func untypedListQuery(outpath string, file string, f UntypedListQuery) error {
	listObj, err := f()
	if err == nil && listObj != nil {
		err = SerializeArrayObj(listObj, outpath, file)
	}
	return err
}

func queryNsResource(ns string, resourceKind string, kubeClient kubernetes.Interface) (runtime.Object, error) {
	switch resourceKind {
	case "configmaps":
		return kubeClient.CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
	case "cronjobs":
		return kubeClient.BatchV2alpha1().CronJobs(ns).List(metav1.ListOptions{})
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
	case "podpresets":
		return kubeClient.Settings().PodPresets(ns).List(metav1.ListOptions{})
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
	case "certificatesigningrequests":
		return kubeClient.Certificates().CertificateSigningRequests().List(metav1.ListOptions{})
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

// QueryNSResources will query namespace-specific resources in the cluster,
// writing them out to <resultsdir>/resources/ns/<ns>/*.json
func QueryNSResources(kubeClient kubernetes.Interface, ns string, dc *Config) []error {
	var errs []error
	glog.Infof("Running ns query (%v)", ns)

	outdir := path.Join(dc.OutputDir(), NSResourceLocation, ns)
	if err := os.MkdirAll(outdir, 0755); err != nil {
		errs = append(errs, err)
		return errs
	}

	for resourceKind, resourceScope := range dc.ResourcesToQuery() {
		// We use annotations to tag resources as being namespaced vs not, skip any
		// that aren't "ns"
		if resourceScope == "ns" {
			lister := func() (runtime.Object, error) { return queryNsResource(ns, resourceKind, kubeClient) }
			if err := objListQuery(outdir+"/", resourceKind+".json", lister); err != nil {
				glog.Warningf("Failed query on resource: %v, ns: %v, error:%v", resourceKind, ns, err)
				errs = append(errs, err)
			}
		}
	}

	return errs
}

// QueryNonNSResources queries non-namespace resources in the cluster, writing
// them out to <resultsdir>/resources/non-ns/*.json
func QueryNonNSResources(kubeClient kubernetes.Interface, dc *Config) []error {
	var errs []error
	glog.Infof("Running non-ns query")

	resources := dc.ResourcesToQuery()
	resourcesOutdir := path.Join(dc.OutputDir(), NonNSResourceLocation)

	if len(resources) > 0 {
		if err := os.MkdirAll(resourcesOutdir, 0755); err != nil {
			errs = append(errs, err)
			return errs
		}
	}

	for resourceKind, resourceScope := range resources {
		// We use annotations to tag resources as being namespaced vs not, skip any
		// that aren't "non-ns"
		if resourceScope == "non-ns" {
			lister := func() (runtime.Object, error) { return queryNonNsResource(resourceKind, kubeClient) }
			if err := objListQuery(resourcesOutdir, resourceKind+".json", lister); err != nil {
				glog.Warningf("Failed query on resource: %v, error:%v", resourceKind, err)
				errs = append(errs, err)
			}
		}
	}

	// dc.Nodes configures whether users want to gather the Nodes resource in the
	// cluster, but we also use that option to guide whether we get node data such
	// as configz and healthz endpoints.
	if dc.Nodes {
		if err := gatherNodeData(kubeClient, dc); err != nil {
			errs = append(errs, err)
		}
	}

	if dc.HostFacts {
		if err := gatherHostFacts(kubeClient, dc); err != nil {
			errs = append(errs, err)
		}
	}

	if dc.ServerVersion {
		objqry := func() (interface{}, error) { return kubeClient.Discovery().ServerVersion() }
		if err := untypedQuery(dc.OutputDir()+"/serverversion", "serverversion.json", objqry); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
