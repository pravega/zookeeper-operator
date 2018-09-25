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

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultZkContainerRepository is the default docker repo for the zookeeper container
	DefaultZkContainerRepository = "spiegela/zookeeper"

	// DefaultZkContainerVersion is the default tag used for for the zookeeper container
	DefaultZkContainerVersion = "3.5.4-beta"

	// DefaultZkContainerPolicy is the default container pull policy used
	DefaultZkContainerPolicy = "Always"

	// DefaultTerminationGracePeriod is the default time given before the container is
	// stopped. This gives clients time to disconnect from a specific node gracefully.
	DefaultTerminationGracePeriod = 30
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZookeeperClusterList is the plural form of the Zookeeper cluster kubernetes resource
type ZookeeperClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ZookeeperCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZookeeperCluster is the type representing the zookeeper cluster kubernetes resource
type ZookeeperCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status,omitempty"`
}

// WithDefaults set default values when not defined in the spec.
func (z *ZookeeperCluster) WithDefaults() {
	z.Spec.withDefaults(z)
}

// ClusterSpec is the zookeeper cluster configuration
type ClusterSpec struct {
	// Zookeeper container image. default is zookeeper:latest
	Image ContainerImage `json:"image"`

	// Labels specifies the labels to attach to pods the operator creates for the
	// zookeeper cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// Size is the expected size of the zookeeper cluster.
	// The pravega-operator will eventually make the size of the running cluster
	// equal to the expected size.
	//
	// The valid range of size is from 1 to 7.
	Size int32 `json:"size"`

	Ports []v1.ContainerPort `json:"ports,omitempty"`

	// Pod defines the policy to create pod for the zookeeper cluster.
	//
	// Updating the Pod does not take effect on any existing pods.
	Pod *PodPolicy `json:"pod,omitempty"`

	// PersistentVolumeClaimSpec is the spec to describe PVC for the container
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume.
	PersistentVolumeClaimSpec *v1.PersistentVolumeClaimSpec `json:"persistence,omitempty"`

	// Conf is the zookeeper configuration, which will be used to generate the
	// static zookeeper configuration. If no configuration is provided required
	// default values will be provided, and optional values will be excluded.
	Conf *ZookeeperConfig `json:"config,omitempty"`
}

func (s *ClusterSpec) withDefaults(z *ZookeeperCluster) {
	s.Image.withDefaults()
	if s.Conf == nil {
		cfg := ZookeeperConfig{}
		cfg.withDefaults()
		s.Conf = &cfg
	}
	if s.Size == 0 {
		s.Size = 3
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
		}
	}
	if z.Spec.Labels == nil {
		z.Spec.Labels = map[string]string{}
	}
	if _, ok := z.Spec.Labels["app"]; !ok {
		z.Spec.Labels["app"] = z.GetName()
	}
	if s.Pod == nil {
		s.Pod = &PodPolicy{}
		s.Pod.withDefaults(z)
	}
	if s.PersistentVolumeClaimSpec == nil {
		s.PersistentVolumeClaimSpec = &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse("20Gi"),
				},
			},
		}
	}
}

// ContainerImage defines the fields needed for a Docker repository image. The format here
// matches the predominant format used in Helm charts.
type ContainerImage struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"pullPolicy"`
}

func (c *ContainerImage) withDefaults() {
	if c.Repository == "" {
		c.Repository = DefaultZkContainerRepository
	}
	if c.Tag == "" {
		c.Tag = DefaultZkContainerVersion
	}
	if c.PullPolicy == "" {
		c.PullPolicy = DefaultZkContainerPolicy
	}
}

// ToString formats a container image struct as a docker compatible repository string.
func (c *ContainerImage) ToString() string {
	return fmt.Sprintf("%s:%s", c.Repository, c.Tag)
}

// PodPolicy defines the common pod configuration for Pods, including when used in deployments,
// stateful-sets, etc.
type PodPolicy struct {
	// Labels specifies the labels to attach to pods the operator creates for the
	// zookeeper cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be eligible
	// to run on a node, the node must have each of the indicated key-value pairs as
	// labels.
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
	// The default value is 1800.
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds"`
}

func (p *PodPolicy) withDefaults(z *ZookeeperCluster) {
	headlessSvcName := fmt.Sprintf("%s-headless", z.GetName())
	if p.Labels == nil {
		p.Labels = map[string]string{"app": z.GetName()}
	}
	if p.TerminationGracePeriodSeconds == 0 {
		p.TerminationGracePeriodSeconds = DefaultTerminationGracePeriod
	}
	if z.Spec.Pod.Labels == nil {
		z.Spec.Pod.Labels = map[string]string{}
	}
	if _, ok := z.Spec.Pod.Labels["app"]; !ok {
		z.Spec.Pod.Labels["app"] = z.GetName()
	}
	if p.Affinity == nil {
		p.Affinity = &v1.Affinity{
			PodAntiAffinity: &v1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
					{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{headlessSvcName},
								},
							},
						},
					},
				},
			},
		}
	}
}

// ZookeeperConfig is the current configuration of each Zookeeper node, which sets these
// values in the config-map
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
}

func (c *ZookeeperConfig) withDefaults() {
	if c.InitLimit == 0 {
		c.InitLimit = 10
	}
	if c.TickTime == 0 {
		c.TickTime = 2000
	}
	if c.SyncLimit == 0 {
		c.SyncLimit = 2
	}
}

// ClusterStatus is the status representing the current state of a cluster.
type ClusterStatus struct {
	// Size is the current size of the cluster
	Size int `json:"size"`

	// Members is the zookeeper members in the cluster
	Members MembersStatus `json:"members"`
}

// MembersStatus is the status of the members of the cluster with both
// ready and unready node membership lists
type MembersStatus struct {
	Ready   []string `json:"ready"`
	Unready []string `json:"unready"`
}
