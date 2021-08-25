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
	"strings"

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
	DefaultZkContainerVersion = "0.2.12"

	// DefaultZkContainerPolicy is the default container pull policy used
	DefaultZkContainerPolicy = "Always"

	// DefaultTerminationGracePeriod is the default time given before the
	// container is stopped. This gives clients time to disconnect from a
	// specific node gracefully.
	DefaultTerminationGracePeriod = 30

	// DefaultZookeeperCacheVolumeSize is the default volume size for the
	// Zookeeper cache volume
	DefaultZookeeperCacheVolumeSize = "20Gi"

	// DefaultReadinessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the readiness probe
	DefaultReadinessProbeInitialDelaySeconds = 10

	// DefaultReadinessProbePeriodSeconds is the default probe period (in seconds)
	// for the readiness probe
	DefaultReadinessProbePeriodSeconds = 10

	// DefaultReadinessProbeFailureThreshold is the default probe failure threshold
	// for the readiness probe
	DefaultReadinessProbeFailureThreshold = 3

	// DefaultReadinessProbeSuccessThreshold is the default probe success threshold
	// for the readiness probe
	DefaultReadinessProbeSuccessThreshold = 1

	// DefaultReadinessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the readiness probe
	DefaultReadinessProbeTimeoutSeconds = 10

	// DefaultLivenessProbeInitialDelaySeconds is the default initial delay (in seconds)
	// for the liveness probe
	DefaultLivenessProbeInitialDelaySeconds = 10

	// DefaultLivenessProbePeriodSeconds is the default probe period (in seconds)
	// for the liveness probe
	DefaultLivenessProbePeriodSeconds = 10

	// DefaultLivenessProbeFailureThreshold is the default probe failure threshold
	// for the liveness probe
	DefaultLivenessProbeFailureThreshold = 3

	// DefaultLivenessProbeTimeoutSeconds is the default probe timeout (in seconds)
	// for the liveness probe
	DefaultLivenessProbeTimeoutSeconds = 10
)

// ZookeeperClusterSpec defines the desired state of ZookeeperCluster
type ZookeeperClusterSpec struct {
	// Image is the  container image. default is zookeeper:0.2.10
	Image ContainerImage `json:"image,omitempty"`

	// Labels specifies the labels to attach to pods the operator creates for
	// the zookeeper cluster.
	Labels map[string]string `json:"labels,omitempty"`

	// Replicas is the expected size of the zookeeper cluster.
	// The pravega-operator will eventually make the size of the running cluster
	// equal to the expected size.
	//
	// The valid range of size is from 1 to 7.
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	Ports []v1.ContainerPort `json:"ports,omitempty"`

	// Pod defines the policy to create pod for the zookeeper cluster.
	// Updating the Pod does not take effect on any existing pods.
	Pod PodPolicy `json:"pod,omitempty"`

	// AdminServerService defines the policy to create AdminServer Service
	// for the zookeeper cluster.
	AdminServerService AdminServerServicePolicy `json:"adminServerService,omitempty"`

	// ClientService defines the policy to create client Service
	// for the zookeeper cluster.
	ClientService ClientServicePolicy `json:"clientService,omitempty"`

	// TriggerRollingRestart if set to true will instruct operator to restart all
	// the pods in the zookeeper cluster, after which this value will be set to false
	TriggerRollingRestart bool `json:"triggerRollingRestart,omitempty"`

	// HeadlessService defines the policy to create headless Service
	// for the zookeeper cluster.
	HeadlessService HeadlessServicePolicy `json:"headlessService,omitempty"`

	//StorageType is used to tell which type of storage we will be using
	//It can take either Ephemeral or persistence
	//Default StorageType is Persistence storage
	StorageType string `json:"storageType,omitempty"`

	// Persistence is the configuration for zookeeper persistent layer.
	// PersistentVolumeClaimSpec and VolumeReclaimPolicy can be specified in here.
	Persistence *Persistence `json:"persistence,omitempty"`

	// Ephemeral is the configuration which helps create ephemeral storage
	// At anypoint only one of Persistence or Ephemeral should be present in the manifest
	Ephemeral *Ephemeral `json:"ephemeral,omitempty"`

	// Conf is the zookeeper configuration, which will be used to generate the
	// static zookeeper configuration. If no configuration is provided required
	// default values will be provided, and optional values will be excluded.
	Conf ZookeeperConfig `json:"config,omitempty"`

	// External host name appended for dns annotation
	DomainName string `json:"domainName,omitempty"`

	// Domain of the kubernetes cluster, defaults to cluster.local
	KubernetesClusterDomain string `json:"kubernetesClusterDomain,omitempty"`

	// Containers defines to support multi containers
	Containers []v1.Container `json:"containers,omitempty"`

	// Init containers to support initialization
	InitContainers []v1.Container `json:"initContainers,omitempty"`

	// Volumes defines to support customized volumes
	Volumes []v1.Volume `json:"volumes,omitempty"`

	// VolumeMounts defines to support customized volumeMounts
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`

	// Probes specifies the timeout values for the Readiness and Liveness Probes
	// for the zookeeper pods.
	// +optional
	Probes *Probes `json:"probes,omitempty"`
}

type Probes struct {
	// +optional
	ReadinessProbe *Probe `json:"readinessProbe,omitempty"`
	// +optional
	LivenessProbe *Probe `json:"livenessProbe,omitempty"`
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
	if s.Probes == nil {
		changed = true
		s.Probes = &Probes{}
	}
	if s.Probes.withDefaults() {
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
			{
				Name:          "admin-server",
				ContainerPort: 8080,
			},
		}
		changed = true
	} else {
		var (
			foundClient, foundQuorum, foundLeader, foundMetrics, foundAdmin bool
		)
		for i := 0; i < len(s.Ports); i++ {
			if s.Ports[i].Name == "client" {
				foundClient = true
			} else if s.Ports[i].Name == "quorum" {
				foundQuorum = true
			} else if s.Ports[i].Name == "leader-election" {
				foundLeader = true
			} else if s.Ports[i].Name == "metrics" {
				foundMetrics = true
			} else if s.Ports[i].Name == "admin-server" {
				foundAdmin = true
			}
		}
		if !foundClient {
			ports := v1.ContainerPort{Name: "client", ContainerPort: 2181}
			s.Ports = append(s.Ports, ports)
			changed = true
		}
		if !foundQuorum {
			ports := v1.ContainerPort{Name: "quorum", ContainerPort: 2888}
			s.Ports = append(s.Ports, ports)
			changed = true
		}
		if !foundLeader {
			ports := v1.ContainerPort{Name: "leader-election", ContainerPort: 3888}
			s.Ports = append(s.Ports, ports)
			changed = true
		}
		if !foundMetrics {
			ports := v1.ContainerPort{Name: "metrics", ContainerPort: 7000}
			s.Ports = append(s.Ports, ports)
			changed = true
		}
		if !foundAdmin {
			ports := v1.ContainerPort{Name: "admin-server", ContainerPort: 8080}
			s.Ports = append(s.Ports, ports)
			changed = true
		}
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
	if strings.EqualFold(s.StorageType, "ephemeral") {
		if s.Ephemeral == nil {
			s.Ephemeral = &Ephemeral{}
			s.Ephemeral.EmptyDirVolumeSource = v1.EmptyDirVolumeSource{}
			changed = true
		}
	} else {
		if s.Persistence == nil {
			s.StorageType = "persistence"
			s.Persistence = &Persistence{}
			changed = true
		}
		if s.Persistence.withDefaults() {
			s.StorageType = "persistence"
			changed = true
		}
	}
	return changed
}

type Probe struct {
	// +kubebuilder:validation:Minimum=0
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	PeriodSeconds int32 `json:"periodSeconds"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	FailureThreshold int32 `json:"failureThreshold"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	SuccessThreshold int32 `json:"successThreshold"`
	// +kubebuilder:validation:Minimum=0
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds"`
}

// Generate CRD using kubebuilder
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=zk
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`,description="The number of ZooKeeper servers in the ensemble"
// +kubebuilder:printcolumn:name="Ready Replicas",type=integer,JSONPath=`.status.readyReplicas`,description="The number of ZooKeeper servers in the ensemble that are in a Ready state"
// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.currentVersion`,description="The current Zookeeper version"
// +kubebuilder:printcolumn:name="Desired Version",type=string,JSONPath=`.spec.image.tag`,description="The desired Zookeeper version"
// +kubebuilder:printcolumn:name="Internal Endpoint",type=string,JSONPath=`.status.internalClientEndpoint`,description="Client endpoint internal to cluster network"
// +kubebuilder:printcolumn:name="External Endpoint",type=string,JSONPath=`.status.externalClientEndpoint`,description="Client endpoint external to cluster network via LoadBalancer"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true

// ZookeeperCluster is the Schema for the zookeeperclusters API
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

// GetKubernetesClusterDomain returns the cluster domain of kubernetes
func (z *ZookeeperCluster) GetKubernetesClusterDomain() string {
	if z.Spec.KubernetesClusterDomain == "" {
		return "cluster.local"
	}
	return z.Spec.KubernetesClusterDomain
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
		} else if p.Name == "admin-server" {
			ports.AdminServer = p.ContainerPort
		}
	}
	return ports
}

// GetClientServiceName returns the name of the client service for the cluster
func (z *ZookeeperCluster) GetClientServiceName() string {
	return fmt.Sprintf("%s-client", z.GetName())
}

// GetAdminServerServiceName returns the name of the admin server service for the cluster
func (z *ZookeeperCluster) GetAdminServerServiceName() string {
	return fmt.Sprintf("%s-admin-server", z.GetName())
}

func (z *ZookeeperCluster) GetTriggerRollingRestart() bool {
	return z.Spec.TriggerRollingRestart
}

// set the value of triggerRollingRestart function
func (z *ZookeeperCluster) SetTriggerRollingRestart(val bool) {
	z.Spec.TriggerRollingRestart = val
}

// Ports groups the ports for a zookeeper cluster node for easy access
type Ports struct {
	Client      int32
	Quorum      int32
	Leader      int32
	Metrics     int32
	AdminServer int32
}

// ContainerImage defines the fields needed for a Docker repository image. The
// format here matches the predominant format used in Helm charts.
type ContainerImage struct {
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag,omitempty"`
	// +kubebuilder:validation:Enum="Always";"Never";"IfNotPresent"
	PullPolicy v1.PullPolicy `json:"pullPolicy,omitempty"`
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

	// +kubebuilder:validation:Minimum=0
	// TerminationGracePeriodSeconds is the amount of time that kubernetes will
	// give for a pod instance to shutdown normally.
	// The default value is 30.
	TerminationGracePeriodSeconds int64 `json:"terminationGracePeriodSeconds,omitempty"`
	// Service Account to be used in pods
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// ImagePullSecrets is a list of references to secrets in the same namespace to use for pulling any images
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
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
	if p.ServiceAccountName == "" {
		p.ServiceAccountName = "default"
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

type AdminServerServicePolicy struct {
	// Annotations specifies the annotations to attach to AdminServer service the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`

	External bool `json:"external,omitempty"`
}

type ClientServicePolicy struct {
	// Annotations specifies the annotations to attach to client service the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`
}

type HeadlessServicePolicy struct {
	// Annotations specifies the annotations to attach to headless service the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`
}

func (s *Probes) withDefaults() (changed bool) {
	if s.ReadinessProbe == nil {
		changed = true
		s.ReadinessProbe = &Probe{}
		s.ReadinessProbe.InitialDelaySeconds = DefaultReadinessProbeInitialDelaySeconds
		s.ReadinessProbe.PeriodSeconds = DefaultReadinessProbePeriodSeconds
		s.ReadinessProbe.FailureThreshold = DefaultReadinessProbeFailureThreshold
		s.ReadinessProbe.SuccessThreshold = DefaultReadinessProbeSuccessThreshold
		s.ReadinessProbe.TimeoutSeconds = DefaultReadinessProbeTimeoutSeconds
	}

	if s.LivenessProbe == nil {
		changed = true
		s.LivenessProbe = &Probe{}
		s.LivenessProbe.InitialDelaySeconds = DefaultLivenessProbeInitialDelaySeconds
		s.LivenessProbe.PeriodSeconds = DefaultLivenessProbePeriodSeconds
		s.LivenessProbe.FailureThreshold = DefaultLivenessProbeFailureThreshold
		s.LivenessProbe.TimeoutSeconds = DefaultLivenessProbeTimeoutSeconds
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
	InitLimit int `json:"initLimit,omitempty"`

	// TickTime is the length of a single tick, which is the basic time unit used
	// by Zookeeper, as measured in milliseconds
	//
	// The default value is 2000.
	TickTime int `json:"tickTime,omitempty"`

	// SyncLimit is the amount of time, in ticks, to allow followers to sync with
	// Zookeeper.
	//
	// The default value is 2.
	SyncLimit int `json:"syncLimit,omitempty"`

	// Clients can submit requests faster than ZooKeeper can process them, especially
	// if there are a lot of clients. Zookeeper will throttle Clients so that requests
	// won't exceed global outstanding limit.
	//
	// The default value is 1000
	GlobalOutstandingLimit int `json:"globalOutstandingLimit,omitempty"`

	// To avoid seeks ZooKeeper allocates space in the transaction log file in
	// blocks of preAllocSize kilobytes
	//
	// The default value is 64M
	PreAllocSize int `json:"preAllocSize,omitempty"`

	// ZooKeeper records its transactions using snapshots and a transaction log
	// The number of transactions recorded in the transaction log before a snapshot
	// can be taken is determined by snapCount
	//
	// The default value is 100,000
	SnapCount int `json:"snapCount,omitempty"`

	// Zookeeper maintains an in-memory list of last committed requests for fast
	// synchronization with followers
	//
	// The default value is 500
	CommitLogCount int `json:"commitLogCount,omitempty"`

	// Snapshot size limit in Kb
	//
	// The defult value is 4GB
	SnapSizeLimitInKb int `json:"snapSizeLimitInKb,omitempty"`

	// Limits the total number of concurrent connections that can be made to a
	//zookeeper server
	//
	// The defult value is 0, indicating no limit
	MaxCnxns int `json:"maxCnxns,omitempty"`

	// Limits the number of concurrent connections that a single client, identified
	// by IP address, may make to a single member of the ZooKeeper ensemble.
	//
	// The default value is 60
	MaxClientCnxns int `json:"maxClientCnxns,omitempty"`

	// The minimum session timeout in milliseconds that the server will allow the
	// client to negotiate
	//
	// The default value is 4000
	MinSessionTimeout int `json:"minSessionTimeout,omitempty"`

	// The maximum session timeout in milliseconds that the server will allow the
	// client to negotiate.
	//
	// The default value is 40000
	MaxSessionTimeout int `json:"maxSessionTimeout,omitempty"`

	// Retain the snapshots according to retain count
	//
	// The default value is 3
	AutoPurgeSnapRetainCount int `json:"autoPurgeSnapRetainCount,omitempty"`

	// The time interval in hours for which the purge task has to be triggered
	//
	// Disabled by default
	AutoPurgePurgeInterval int `json:"autoPurgePurgeInterval,omitempty"`

	// QuorumListenOnAllIPs when set to true the ZooKeeper server will listen for
	// connections from its peers on all available IP addresses, and not only the
	// address configured in the server list of the configuration file. It affects
	// the connections handling the ZAB protocol and the Fast Leader Election protocol.
	//
	// The default value is false.
	QuorumListenOnAllIPs bool `json:"quorumListenOnAllIPs,omitempty"`

	// key-value map of additional zookeeper configuration parameters
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	AdditionalConfig map[string]string `json:"additionalConfig,omitempty"`
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
	if c.GlobalOutstandingLimit == 0 {
		changed = true
		c.GlobalOutstandingLimit = 1000
	}
	if c.PreAllocSize == 0 {
		changed = true
		c.PreAllocSize = 65536
	}
	if c.SnapCount == 0 {
		changed = true
		c.SnapCount = 10000
	}
	if c.CommitLogCount == 0 {
		changed = true
		c.CommitLogCount = 500
	}
	if c.SnapSizeLimitInKb == 0 {
		changed = true
		c.SnapSizeLimitInKb = 4194304
	}
	if c.MaxClientCnxns == 0 {
		changed = true
		c.MaxClientCnxns = 60
	}
	if c.MinSessionTimeout == 0 {
		changed = true
		c.MinSessionTimeout = 2 * c.TickTime
	}
	if c.MaxSessionTimeout == 0 {
		changed = true
		c.MaxSessionTimeout = 20 * c.TickTime
	}
	if c.AutoPurgeSnapRetainCount == 0 {
		changed = true
		c.AutoPurgeSnapRetainCount = 3
	}
	if c.AutoPurgePurgeInterval == 0 {
		changed = true
		c.AutoPurgePurgeInterval = 1
	}

	return changed
}

type Persistence struct {
	// VolumeReclaimPolicy is a zookeeper operator configuration. If it's set to Delete,
	// the corresponding PVCs will be deleted by the operator when zookeeper cluster is deleted.
	// The default value is Retain.
	// +kubebuilder:validation:Enum="Delete";"Retain"
	VolumeReclaimPolicy VolumeReclaimPolicy `json:"reclaimPolicy,omitempty"`
	// PersistentVolumeClaimSpec is the spec to describe PVC for the container
	// This field is optional. If no PVC is specified default persistentvolume
	// will get created.
	PersistentVolumeClaimSpec v1.PersistentVolumeClaimSpec `json:"spec,omitempty"`
	// Annotations specifies the annotations to attach to pvc the operator
	// creates.
	Annotations map[string]string `json:"annotations,omitempty"`
}

type Ephemeral struct {
	//EmptyDirVolumeSource is optional and this will create the emptydir volume
	//It has two parameters Medium and SizeLimit which are optional as well
	//Medium specifies What type of storage medium should back this directory.
	//SizeLimit specifies Total amount of local storage required for this EmptyDir volume.
	EmptyDirVolumeSource v1.EmptyDirVolumeSource `json:"emptydirvolumesource,omitempty"`
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
