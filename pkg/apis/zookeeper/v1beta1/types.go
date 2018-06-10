package v1beta1

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const DefaultZkContainerRepository = "zookeeper"
const DefaultZkContainerVersion = "3.5"
const DefaultZkContainerPolicy = "IfNotPresent"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ZookeeperClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ZookeeperCluster `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ZookeeperCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterSpec   `json:"spec"`
	Status            ClusterStatus `json:"status,omitempty"`
}

func (z *ZookeeperCluster) WithDefaults() {
	z.Spec.withDefaults(z)
}

type ClusterSpec struct {
	// Zookeeper container image. default is zookeeper:latest
	Image ContainerImage `json:"image"`

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
}

func (s *ClusterSpec) withDefaults(z *ZookeeperCluster) {
	s.Image.withDefaults()
	if s.Size == 0 {
		s.Size = 1
	}
	if s.Ports == nil {
		s.Ports = []v1.ContainerPort{
			{
				Name:          "client",
				HostPort:      2181,
				ContainerPort: 2181,
			},
			{
				Name:          "server",
				HostPort:      2888,
				ContainerPort: 2888,
			},
			{
				Name:          "election",
				HostPort:      3888,
				ContainerPort: 3888,
			},
		}
	}
	if s.Pod == nil {
		pod := PodPolicy{}
		pod.withDefaults(z)
		s.Pod = &pod
	}
}

type ContainerImage struct {
	Repository string        `json:"repository"`
	Tag        string        `json:"tag"`
	PullPolicy v1.PullPolicy `json:"imagePullPolicy"`
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

func (c *ContainerImage) ToString() string {
	return fmt.Sprintf("%s:%s", c.Repository, c.Tag)
}

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
	Env []v1.EnvVar `json:"Env,omitempty"`

	// PersistentVolumeClaimSpec is the spec to describe PVC for the container
	// This field is optional. If no PVC spec, stateful containers will use
	// emptyDir as volume.
	PersistentVolumeClaimSpec *v1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`

	// Annotations specifies the annotations to attach to pods the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`

	// SecurityContext specifies the security context for the entire pod
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// DNSTimeoutInSecond is the maximum allowed time for the init container of the pod to
	// reverse DNS lookup its IP given the hostname.
	// The default is to wait indefinitely and has a vaule of 0.
	DNSTimeoutInSecond int `json:"DNSTimeoutInSecond,omitempty"`
}

func (p *PodPolicy) withDefaults(z *ZookeeperCluster) {
	if p.Labels == nil {
		p.Labels = map[string]string{"app": z.GetObjectMeta().GetName()}
	}
}

type ClusterStatus struct {
}
