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

package dispatch

import (
	"encoding/hex"
	"fmt"

	"github.com/heptio/sonobuoy/pkg/agent"
	"github.com/heptio/sonobuoy/pkg/buildinfo"
	"github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	v1beta1ext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
)

// AgentDispatcher dispatches a sonobuoy agent to each node with configured settings
type AgentDispatcher struct {
	UUID        uuid.UUID
	AgentConfig *agent.Config
	KubeClient  kubernetes.Interface
	Namespace   string

	configMap *v1.ConfigMap
	daemonSet *v1beta1ext.DaemonSet
}

// NewAgentDispatcher returns a new agent dispatcher with a unique UUID and default settings
func NewAgentDispatcher(cfg *agent.Config, kubeclient kubernetes.Interface) *AgentDispatcher {
	return &AgentDispatcher{
		UUID:        uuid.NewV4(),
		AgentConfig: cfg,
		KubeClient:  kubeclient,
		Namespace:   metav1.NamespaceSystem,
	}
}

// Dispatch dispatches agent pods according to the AgentDispatcher's configuration.
func (d *AgentDispatcher) Dispatch() error {
	var err error
	var cm *v1.ConfigMap
	var ds *v1beta1ext.DaemonSet

	// We don't want to do this twice for one dispatcher
	if d.configMap != nil {
		return fmt.Errorf("requested to create an agent configMap when one already exists (%v)", d.configMap.SelfLink)
	}
	if d.daemonSet != nil {
		return fmt.Errorf("requested to create an agent daemonSet when one already exists (%v)", d.daemonSet.SelfLink)
	}

	// Build the resources in memory
	if cm, err = d.buildConfigMap(); err != nil {
		return err
	}
	if ds, err = d.buildDaemonSet(); err != nil {
		return err
	}

	// Submit them to the API server, capturing the results
	if d.configMap, err = d.KubeClient.CoreV1().ConfigMaps(d.Namespace).Create(cm); err != nil {
		return fmt.Errorf("could not create configMap for sonobuoy agents: %v", err)
	}
	if d.daemonSet, err = d.KubeClient.ExtensionsV1beta1().DaemonSets(d.Namespace).Create(ds); err != nil {
		return fmt.Errorf("could not create DaemonSet for sonobuoy agents: %v", err)
	}

	return nil
}

// Cleanup deletes resources created by Dispatch(), with the specified Grace Period
func (d *AgentDispatcher) Cleanup(gracePeriod int64) error {
	var err error

	deleteOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	if d.configMap != nil {
		err = d.KubeClient.CoreV1().ConfigMaps(d.Namespace).Delete(d.configMap.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf("could not delete ConfigMap %v: %v", d.configMap.Name, err)
		}
	}

	if d.daemonSet != nil {
		err = d.KubeClient.ExtensionsV1beta1().DaemonSets(d.Namespace).Delete(d.daemonSet.Name, deleteOptions)
		if err != nil {
			return fmt.Errorf("could not delete DaemonSet %v: %v", d.daemonSet.Name, err)
		}
	}

	return err
}

// SessionID returns a unique identifier for this dispatcher, used for tagging objects and cleaning them up later
func (d *AgentDispatcher) SessionID() string {
	ret := make([]byte, hex.EncodedLen(8))
	hex.Encode(ret, d.UUID.Bytes()[0:8])
	return string(ret)
}

func agentImage() string {
	var image, tag string
	if buildinfo.Version == "" {
		tag = "latest"
	} else {
		tag = buildinfo.Version
	}

	if buildinfo.DockerImage == "" {
		image = "gcr.io/heptio-images/sonobuoy"
	} else {
		image = buildinfo.DockerImage
	}
	return image + ":" + tag
}

func (d *AgentDispatcher) configMapName() string {
	return "sonobuoy-agent-config-" + d.SessionID()
}

func (d *AgentDispatcher) buildConfigMap() (*v1.ConfigMap, error) {
	// We get to build the agent config directly from our own data structures,
	// this is where doing this natively in golang helps a lot (as opposed to
	// shelling out to kubectl)
	cfgjson, err := json.Marshal(d.AgentConfig)
	if err != nil {
		return nil, err
	}

	cmap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   d.configMapName(),
			Labels: d.applyDefaultLabels(map[string]string{}),
		},
		Data: map[string]string{
			"agent.json": string(cfgjson),
		},
	}

	return cmap, err
}

func (d *AgentDispatcher) buildDaemonSet() (*v1beta1ext.DaemonSet, error) {
	tru := true // need to pass a pointer to a bool for v1.SecurityContext.Privileged, yay

	// Here comes lots of inline struct literals!
	ds := &v1beta1ext.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "sonobuoy-agents-" + d.SessionID(),
			Labels: d.applyDefaultLabels(map[string]string{}),
		},
		Spec: v1beta1ext.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"sonobuoy-run": d.SessionID(),
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: d.applyDefaultLabels(map[string]string{}),
				},
				Spec: v1.PodSpec{
					Tolerations: []v1.Toleration{
						v1.Toleration{
							Key:      "node-role.kubernetes.io/master",
							Operator: v1.TolerationOpExists,
							Effect:   v1.TaintEffectNoSchedule,
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
					HostNetwork:   true,
					HostIPC:       true,
					HostPID:       true,
					Containers: []v1.Container{
						v1.Container{
							Name:            "sonobuoy-agent",
							Image:           agentImage(),
							ImagePullPolicy: v1.PullIfNotPresent,
							Env: []v1.EnvVar{
								v1.EnvVar{
									Name: "NODE_NAME",
									ValueFrom: &v1.EnvVarSource{
										FieldRef: &v1.ObjectFieldSelector{
											FieldPath: "spec.nodeName",
										},
									},
								},
							},
							Command: []string{"sh", "-c", "/sonobuoy agent -v 5 --logtostderr && sleep 3600"},
							SecurityContext: &v1.SecurityContext{
								Privileged: &tru,
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      "root",
									MountPath: "/node",
								},
								v1.VolumeMount{
									Name:      "config",
									MountPath: "/etc/sonobuoy",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: "root",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/",
								},
							},
						},
						v1.Volume{
							Name: "config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: d.configMapName(),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ds, nil
}

func (d *AgentDispatcher) applyDefaultLabels(labels map[string]string) map[string]string {
	labels["component"] = "sonobuoy"
	labels["tier"] = "analysis"
	labels["sonobuoy-run"] = d.SessionID()

	return labels
}
