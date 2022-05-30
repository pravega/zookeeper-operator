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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ZookeeperBackupSpec defines the desired state of ZookeeperBackup
type ZookeeperBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the ZookeeperCluster to backup
	// +kubebuilder:validation:Required
	ZookeeperCluster string `json:"zookeeperCluster"`
	// Schedule in Cron format
	// +kubebuilder:default:="0 0 */1 * *"
	// +optional
	Schedule string `json:"schedule,omitempty"`
	// Number of backups to store
	// +kubebuilder:default:="7"
	// +optional
	BackupsToKeep string `json:"backupsToKeep,omitempty"`
	// Data Storage Capacity
	// +kubebuilder:default:="1Gi"
	// +optional
	DataCapacity string `json:"dataCapacity,omitempty"`
	// Data Storage Class name
	// +kubebuilder:validation:Required
	DataStorageClass string `json:"dataStorageClass,omitempty"`

	// Image for backup procedure
	Image ContainerImage `json:"image,omitempty"`
}

func (s *ZookeeperBackupSpec) withDefaults() (changed bool) {
	if s.Schedule == "" {
		s.Schedule = "0 0 */1 * *"
		changed = true
	}
	if s.BackupsToKeep == "" {
		s.BackupsToKeep = "7"
		changed = true
	}
	if s.DataCapacity == "" {
		s.DataCapacity = "1Gi"
		changed = true
	}
	if s.Image.Repository == "" {
		s.Image.Repository = "pravega/zkbackup"
		changed = true
	}
	if s.Image.Tag == "" {
		s.Image.Tag = "latest"
		changed = true
	}
	return changed
}

// ZookeeperBackupStatus defines the observed state of ZookeeperBackup
type ZookeeperBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ZookeeperBackup is the Schema for the zookeeperbackups API
type ZookeeperBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperBackupSpec   `json:"spec,omitempty"`
	Status ZookeeperBackupStatus `json:"status,omitempty"`
}

func (z *ZookeeperBackup) WithDefaults() bool {
	return z.Spec.withDefaults()
}

//+kubebuilder:object:root=true

// ZookeeperBackupList contains a list of ZookeeperBackup
type ZookeeperBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ZookeeperBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ZookeeperBackup{}, &ZookeeperBackupList{})
}
