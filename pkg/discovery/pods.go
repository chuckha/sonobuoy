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
	"io/ioutil"
	"os"
	"path"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

const (
	PodsLocation = "pods"
)

func gatherPodLogs(kubeClient kubernetes.Interface, ns string, opts metav1.ListOptions, dc *Config) []error {
	var errs []error
	podlist, err := kubeClient.CoreV1().Pods(ns).List(opts)
	if err != nil {
		errs = append(errs, err)
		return errs
	}

	glog.Info("Collecting Pod Logs...")
	for _, pod := range podlist.Items {
		body, err := kubeClient.CoreV1().Pods(ns).GetLogs(pod.Name, &v1.PodLogOptions{}).Do().Raw()
		if err == nil {
			outdir := path.Join(dc.OutputDir(), NSResourceLocation, ns, PodsLocation, pod.Name)
			if err = os.MkdirAll(outdir, 0755); err != nil {
				errs = append(errs, err)
			}
			if err = ioutil.WriteFile(outdir+"/logs.txt", body, 0644); err != nil {
				errs = append(errs, err)
			}
		} else {
			errs = append(errs, err)
		}
	}

	return errs
}
