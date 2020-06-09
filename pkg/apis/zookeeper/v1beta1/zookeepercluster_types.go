/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultZkContainerRepository is the default docker repo for the zookeeper
	// container
	DefaultZkContainerRepository = "pravega/zookeeper"

	// DefaultZkContainerVersion is the default tag used for for the zookeeper
	// container
	DefaultZkContainerVersion = "0.2.7"

	// DefaultZkContainerPolicy is the default container pull policy used
	DefaultZkContainerPolicy = "Always"

	// DefaultTerminationGracePeriod is the default time given before the
	// container is stopped. This gives clients time to disconnect from a
	// specific node gracefully.
	DefaultTerminationGracePeriod = 180

	// DefaultZookeeperCacheVolumeSize is the default volume size for the
	// Zookeeper cache volume
	DefaultZookeeperCacheVolumeSize = "20Gi"
)

// ZookeeperClusterSpec defines the desired state of ZookeeperCluster
type ZookeeperClusterSpec struct {
	// Image is the  container image. default is zookeeper:0.2.7
	Image ContainerImage `json:"image,omitempty"`

	// Labels specifies the labels to attach to pods the operator creates for
	// the zookeeper cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// Replicas is the expected size of the zookeeper cluster.
	// The pravega-operator will eventually make the size of the running cluster
	// equal to the expected size.
	//
	// The valid range of size is from 1 to 7.
	Replicas int32 `json:"replicas"`

	Ports []v1.ContainerPort `json:"ports,omitempty"`

	// Pod defines the policy to create pod for the zookeeper cluster.
	//
	// Updating the Pod does not take effect on any existing pods.
	Pod PodPolicy `json:"pod,omitempty"`

	// Persistence is the configuration for zookeeper persistent layer.
	// PersistentVolumeClaimSpec and VolumeReclaimPolicy can be specified in here.
	Persistence *Persistence `json:"persistence,omitempty"`

	// Conf is the zookeeper configuration, which will be used to generate the
	// static zookeeper configuration. If no configuration is provided required
	// default values will be provided, and optional values will be excluded.
	Conf ZookeeperConfig `json:"config,omitempty"`

	// Domain Name to be used for DNS
	DomainName string `json:"domainName,omitempty"`
}

func (s *ZookeeperClusterSpec) withDefaults(z *ZookeeperCluster) (changed bool) {
	changed = s.Image.withDefaults()
	if s.Conf.withDefaults() {
		changed = true
	}
	if s.Replicas == 0 {
		s.Replicas = 3
		changed = true
	}
	if s.Ports == nil {
		s.Ports = []v1.ContainerPort{
			{
				Name:          "client",
				ContainerPort: 2181,
			},
			{
				Name:          "quorum",
				ContainerPort: 2888,
			},
			{
				Name:          "leader-election",
				ContainerPort: 3888,
			},
			{
				Name:          "metrics",
				ContainerPort: 7000,
			},
		}
		changed = true
	}
	if z.Spec.Labels == nil {
		z.Spec.Labels = map[string]string{}
		changed = true
	}
	if _, ok := z.Spec.Labels["app"]; !ok {
		z.Spec.Labels["app"] = z.GetName()
		changed = true
	}
	if _, ok := z.Spec.Labels["release"]; !ok {
		z.Spec.Labels["release"] = z.GetName()
		changed = true
	}
	if s.Pod.withDefaults(z) {
		changed = true
	}
	if s.Persistence == nil {
		s.Persistence = &Persistence{}
		changed = true
	}
	if s.Persistence.withDefaults() {
		changed = true
	}
	return changed
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZookeeperCluster is the Schema for the zookeeperclusters API
// +k8s:openapi-gen=true
type ZookeeperCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperClusterSpec   `json:"spec,omitempty"`
	Status ZookeeperClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (z *ZookeeperCluster) WithDefaults() bool {
	return z.Spec.withDefaults(z)
}

// ConfigMapName returns the name of the cluster config-map
func (z *ZookeeperCluster) ConfigMapName() string {
	return fmt.Sprintf("%s-configmap", z.GetName())
}

// ZookeeperPorts returns a struct of ports
func (z *ZookeeperCluster) ZookeeperPorts() Ports {
	ports := Ports{}
	for _, p := range z.Spec.Ports {
		if p.Name == "client" {
			ports.Client = p.ContainerPort
		} else if p.Name == "quorum" {
			ports.Quorum = p.ContainerPort
		} else if p.Name == "leader-election" {
			ports.Leader = p.ContainerPort
		} else if p.Name == "metrics" {
			ports.Metrics = p.ContainerPort
		}
	}
	return ports
}

// GetClientServiceName returns the name of the client service for the cluster
func (z *ZookeeperCluster) GetClientServiceName() string {
	return fmt.Sprintf("%s-client", z.GetName())
}

// Ports groups the ports for a zookeeper cluster node for easy access
type Ports struct {
	Client  int32
	Quorum  int32
	Leader  int32
	Metrics int32
}

// ContainerImage defines the fields needed for a Docker repository image. The
// format here matches the predominant format used in Helm charts.
type ContainerImage struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

func (c *ContainerImage) withDefaults() (changed bool) {
	if c.Repository == "" {
		changed = true
		c.Repository = DefaultZkContainerRepository
	}
	if c.Tag == "" {
		changed = true
		c.Tag = DefaultZkContainerVersion
	}
	if c.PullPolicy == "" {
		changed = true
		c.PullPolicy = DefaultZkContainerPolicy
	}
	return changed
}

// ToString formats a container image struct as a docker compatible repository
// string.
func (c *ContainerImage) ToString() string {
	return fmt.Sprintf("%s:%s", c.Repository, c.Tag)
}

// PodPolicy defines the common pod configuration for Pods, including when used
// in deployments, stateful-sets, etc.
type PodPolicy struct {
	// Labels specifies the labels to attach to pods the operator creates for
	// the zookeeper cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be
	// eligible to run on a node, the node must have each of the indicated
	// key-value pairs as labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// The scheduling constraints on pods.
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// Resources is the resource requirements for the container.
	// This field cannot be updated once the cluster is created.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Tolerations specifies the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// List of environment variables to set in the container.
	// This field cannot be updated.
	Env []v1.EnvVar `json:"env,omitempty"`

	// Annotations specifies the annotations to attach to pods the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`

	// SecurityContext specifies the security context for the entire pod
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// TerminationGracePeriodSeconds is the amount of time that kubernetes will
	// give for a pod instance to shutdown normally.
	// The default value is 180.
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds"`
}

func (p *PodPolicy) withDefaults(z *ZookeeperCluster) (changed bool) {
	if p.Labels == nil {
		p.Labels = map[string]string{}
		changed = true
	}
	if p.TerminationGracePeriodSeconds == 0 {
		p.TerminationGracePeriodSeconds = DefaultTerminationGracePeriod
		changed = true
	}
	if z.Spec.Pod.Labels == nil {
		p.Labels = map[string]string{}
		changed = true
	}
	if _, ok := p.Labels["app"]; !ok {
		p.Labels["app"] = z.GetName()
		changed = true
	}
	if _, ok := p.Labels["release"]; !ok {
		p.Labels["release"] = z.GetName()
		changed = true
	}
	if p.Affinity == nil {
		p.Affinity = &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
					{
						Weight: 20,
						PodAffinityTerm: v1.PodAffinityTerm{
							TopologyKey: "kubernetes.io/hostname",
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "app",
										Operator: metav1.LabelSelectorOpIn,
										Values:   []string{z.GetName()},
									},
								},
							},
						},
					},
				},
			},
		}
		changed = true
	}
	return changed
}

// ZookeeperConfig is the current configuration of each Zookeeper node, which
// sets these values in the config-map
type ZookeeperConfig struct {
	// InitLimit is the amount of time, in ticks, to allow followers to connect
	// and sync to a leader.
	//
	// Default value is 10.
	InitLimit int `json:"initLimit"`

	// TickTime is the length of a single tick, which is the basic time unit used
	// by Zookeeper, as measured in milliseconds
	//
	// The default value is 2000.
	TickTime int `json:"tickTime"`

	// SyncLimit is the amount of time, in ticks, to allow followers to sync with
	// Zookeeper.
	//
	// The default value is 2.
	SyncLimit int `json:"syncLimit"`

	// QuorumListenOnAllIPs when set to true the ZooKeeper server will listen for
	// connections from its peers on all available IP addresses, and not only the
	// address configured in the server list of the configuration file. It affects
	// the connections handling the ZAB protocol and the Fast Leader Election protocol.
	//
	// The default value is false.
	QuorumListenOnAllIPs bool `json:"quorumListenOnAllIPs"`
}

func (c *ZookeeperConfig) withDefaults() (changed bool) {
	if c.InitLimit == 0 {
		changed = true
		c.InitLimit = 10
	}
	if c.TickTime == 0 {
		changed = true
		c.TickTime = 2000
	}
	if c.SyncLimit == 0 {
		changed = true
		c.SyncLimit = 2
	}
	return changed
}

type Persistence struct {
	// VolumeReclaimPolicy is a zookeeper operator configuration. If it's set to Delete,
	// the corresponding PVCs will be deleted by the operator when zookeeper cluster is deleted.
	// The default value is Retain.
	VolumeReclaimPolicy VolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
	// PersistentVolumeClaimSpec is the spec to describe PVC for the container
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume.
	PersistentVolumeClaimSpec v1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
}

func (p *Persistence) withDefaults() (changed bool) {
	if !p.VolumeReclaimPolicy.isValid() {
		changed = true
		p.VolumeReclaimPolicy = VolumeReclaimPolicyRetain
	}

	p.PersistentVolumeClaimSpec.AccessModes = []v1.PersistentVolumeAccessMode{
		v1.ReadWriteOnce,
	}

	storage, _ := p.PersistentVolumeClaimSpec.Resources.Requests["storage"]
	if storage.IsZero() {
		p.PersistentVolumeClaimSpec.Resources.Requests = v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(DefaultZookeeperCacheVolumeSize),
		}
		changed = true
	}
	return changed
}

func (v VolumeReclaimPolicy) isValid() bool {
	if v != VolumeReclaimPolicyDelete && v != VolumeReclaimPolicyRetain {
		return false
	}
	return true
}

type VolumeReclaimPolicy string

const (
	VolumeReclaimPolicyRetain VolumeReclaimPolicy = "Retain"
	VolumeReclaimPolicyDelete VolumeReclaimPolicy = "Delete"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZookeeperClusterList contains a list of ZookeeperCluster
type ZookeeperClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZookeeperCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZookeeperCluster{}, &ZookeeperClusterList{})
}
