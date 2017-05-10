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
	"time"

	"github.com/golang/glog"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

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

// queryData captures the results of the run for post-processing
type queryData struct {
	QueryObj    string `json:"queryobj,omitempty"`
	ElapsedTime string `json:"time,omitempty"`
	Error       error  `json:"error,omitempty"`
}

// objListQuery performs a list query and serialize the results
func objListQuery(outpath string, file string, f ObjQuery) (time.Duration, error) {
	start := time.Now()
	listObj, err := f()
	duration := time.Since(start)
	if err != nil {
		return duration, err
	}
	if listObj == nil {
		return duration, fmt.Errorf("got invalid response from API server")
	}
	if listPtr, err := meta.GetItemsPtr(listObj); err == nil {
		if items, err := conversion.EnforcePtr(listPtr); err == nil {
			if items.Len() > 0 {
				err = SerializeObj(listPtr, outpath, file)
			}
		}
	}
	return duration, err
}

// untypedQuery performs a untyped query and serialize the results
func untypedQuery(outpath string, file string, f UntypedQuery) (time.Duration, error) {
	start := time.Now()
	Obj, err := f()
	duration := time.Since(start)
	if err == nil && Obj != nil {
		err = SerializeObj(Obj, outpath, file)
	}
	return duration, err
}

// untypedListQuery performs a untyped list query and serialize the results
func untypedListQuery(outpath string, file string, f UntypedListQuery) (time.Duration, error) {
	start := time.Now()
	listObj, err := f()
	duration := time.Since(start)
	if err == nil && listObj != nil {
		err = SerializeArrayObj(listObj, outpath, file)
	}
	return duration, err
}

// recordResults will write out the execution results of a query.
func recordResults(f *os.File, name string, duration time.Duration, recerr error) error {
	summary := &queryData{
		QueryObj:    name,
		ElapsedTime: duration.String(),
		Error:       recerr,
	}
	if err := SerializeObjAppend(f, summary); err != nil {
		return err
	}
	return nil
}

// Wraps the execution of the function with a recorded timed snapshot
func timedQuery(f *os.File, name string, errs []error, fn func() (time.Duration, error)) {
	duration, err := fn()
	if err != nil {
		glog.Warningf("Failed query on resource: %v, error:%v", name, err)
		errs = append(errs, err)
	}
	if err = recordResults(f, name, duration, err); err != nil {
		errs = append(errs, err)
	}
}

// queryNsResource performs the appropriate namespace-scoped query according to its input args
func queryNsResource(ns string, resourceKind string, opts metav1.ListOptions, kubeClient kubernetes.Interface) (runtime.Object, error) {
	switch resourceKind {
	case "configmaps":
		return kubeClient.CoreV1().ConfigMaps(ns).List(opts)
	case "cronjobs":
		return kubeClient.BatchV2alpha1().CronJobs(ns).List(opts)
	case "daemonsets":
		return kubeClient.Extensions().DaemonSets(ns).List(opts)
	case "deployments":
		return kubeClient.Apps().Deployments(ns).List(opts)
	case "endpoints":
		return kubeClient.CoreV1().Endpoints(ns).List(opts)
	case "events":
		return kubeClient.CoreV1().Events(ns).List(opts)
	case "horizontalpodautoscalers":
		return kubeClient.Autoscaling().HorizontalPodAutoscalers(ns).List(opts)
	case "ingresses":
		return kubeClient.Extensions().Ingresses(ns).List(opts)
	case "jobs":
		return kubeClient.Batch().Jobs(ns).List(opts)
	case "limitranges":
		return kubeClient.CoreV1().LimitRanges(ns).List(opts)
	case "persistentvolumeclaims":
		return kubeClient.CoreV1().PersistentVolumeClaims(ns).List(opts)
	case "pods":
		return kubeClient.CoreV1().Pods(ns).List(opts)
	case "poddisruptionbudgets":
		return kubeClient.Policy().PodDisruptionBudgets(ns).List(opts)
	case "podpresets":
		return kubeClient.Settings().PodPresets(ns).List(opts)
	case "podtemplates":
		return kubeClient.CoreV1().PodTemplates(ns).List(opts)
	case "replicasets":
		return kubeClient.Extensions().ReplicaSets(ns).List(opts)
	case "replicationcontrollers":
		return kubeClient.CoreV1().ReplicationControllers(ns).List(opts)
	case "resourcequotas":
		return kubeClient.CoreV1().ResourceQuotas(ns).List(opts)
	case "rolebindings":
		return kubeClient.Rbac().RoleBindings(ns).List(opts)
	case "roles":
		return kubeClient.Rbac().Roles(ns).List(opts)
	case "secrets":
		return kubeClient.CoreV1().Secrets(ns).List(opts)
	case "serviceaccounts":
		return kubeClient.CoreV1().ServiceAccounts(ns).List(opts)
	case "services":
		return kubeClient.CoreV1().Services(ns).List(opts)
	case "statefulsets":
		return kubeClient.Apps().StatefulSets(ns).List(opts)
	default:
		return nil, fmt.Errorf("don't know how to handle namespaced resource %v", resourceKind)
	}
}

// queryNonNsResource performs the appropriate non-namespace-scoped query according to its input args
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

	// 1. Create the parent directory we will use to store the results
	outdir := path.Join(dc.OutputDir(), NSResourceLocation, ns)
	if err := os.MkdirAll(outdir, 0755); err != nil {
		errs = append(errs, err)
		return errs
	}

	// 2. Create the results output file.
	f, err := os.Create(outdir + "/results.json")
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	defer func() {
		f.WriteString("{}]")
		f.Close()
	}()

	_, err = f.WriteString("[")
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	// 3. Setup label filter if there is one.
	opts := metav1.ListOptions{}
	if len(dc.LabelSelector) > 0 {
		if _, err := labels.Parse(dc.LabelSelector); err != nil {
			glog.Warningf("Labelselector %v failed to parse with error %v", dc.LabelSelector, err)
		} else {
			opts.LabelSelector = dc.LabelSelector
		}
	}

	// 4. Execute the ns-query
	for resourceKind, resourceScope := range dc.ResourcesToQuery() {
		// We use annotations to tag resources as being namespaced vs not, skip any
		// that aren't "ns"
		if resourceScope == "ns" {
			lister := func() (runtime.Object, error) { return queryNsResource(ns, resourceKind, opts, kubeClient) }
			query := func() (time.Duration, error) { return objListQuery(outdir+"/", resourceKind+".json", lister) }
			timedQuery(f, resourceKind, errs, query)
		}
	}

	if dc.CollectPodLogs {
		// NOTE: pod log collection is an aggregated time b/c propagating that detail back up
		// is odd and would pollute some of the output.
		start := time.Now()
		if errlst := gatherPodLogs(kubeClient, ns, opts, dc); errlst != nil {
			err = errlst[0]
			errs = append(errs, errlst...)
		}
		duration := time.Since(start)
		recordResults(f, "podlogs", duration, err)
	}

	return errs
}

// QueryNonNSResources queries non-namespace resources in the cluster, writing
// them out to <resultsdir>/resources/non-ns/*.json
func QueryNonNSResources(kubeClient kubernetes.Interface, dc *Config) []error {
	var errs []error
	glog.Infof("Running non-ns query")

	resources := dc.ResourcesToQuery()

	// 1. Create the parent directory we will use to store the results
	outdir := path.Join(dc.OutputDir(), NonNSResourceLocation)
	if len(resources) > 0 {
		if err := os.MkdirAll(outdir, 0755); err != nil {
			errs = append(errs, err)
			return errs
		}
	}

	// 2. Create the results output file.
	f, err := os.Create(outdir + "/results.json")
	if err != nil {
		errs = append(errs, err)
		return errs
	}
	defer func() {
		f.WriteString("{}]")
		f.Close()
	}()

	_, err = f.WriteString("[")
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	// 3. Execute the non-ns-query
	for resourceKind, resourceScope := range resources {
		// We use annotations to tag resources as being namespaced vs not, skip any
		// that aren't "non-ns"
		if resourceScope == "non-ns" {
			lister := func() (runtime.Object, error) { return queryNonNsResource(resourceKind, kubeClient) }
			query := func() (time.Duration, error) { return objListQuery(outdir+"/", resourceKind+".json", lister) }
			timedQuery(f, resourceKind, errs, query)
		}
	}

	// dc.Nodes configures whether users want to gather the Nodes resource in the
	// cluster, but we also use that option to guide whether we get node data such
	// as configz and healthz endpoints.
	if dc.Nodes {
		// NOTE: Node data collection is an aggregated time b/c propagating that detail back up
		// is odd and would pollute some of the output.
		start := time.Now()
		if err = gatherNodeData(kubeClient, dc); err != nil {
			errs = append(errs, err)
		}
		duration := time.Since(start)
		recordResults(f, "podlogs", duration, err)
	}

	if dc.ServerVersion {
		objqry := func() (interface{}, error) { return kubeClient.Discovery().ServerVersion() }
		query := func() (time.Duration, error) {
			return untypedQuery(dc.OutputDir()+"/serverversion", "serverversion.json", objqry)
		}
		timedQuery(f, "serverversion", errs, query)
	}

	return errs
}
