/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/api/v1beta1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ZkClusterName = "zk-cluster"
	ZkBackupName  = "zk-backup"
	Namespace     = "default"
	Hostname      = "node-0"
)

func MockGetLeader(_ string, _ int32) (string, error) {
	return Hostname, nil
}

var _ = Describe("ZookeeperBackup controller", func() {

	var (
		s            = scheme.Scheme
		mockZkClient = new(MockZookeeperClient)
		rCl          *ZookeeperClusterReconciler
		rBk          *ZookeeperBackupReconciler
	)

	Context("Reconcile", func() {
		var (
			cl        client.Client
			res       reconcile.Result
			req       reconcile.Request
			zkBk      *v1beta1.ZookeeperBackup
			zkCl      *v1beta1.ZookeeperCluster
			leaderPod *corev1.Pod
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      ZkBackupName,
					Namespace: Namespace,
				},
			}
			zkCl = &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ZkClusterName,
					Namespace: Namespace,
				},
			}
			zkBk = &v1beta1.ZookeeperBackup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ZkBackupName,
					Namespace: Namespace,
				},
				Spec: v1beta1.ZookeeperBackupSpec{
					ZookeeperCluster: ZkClusterName,
				},
			}
			leaderPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ZkClusterName + "-0",
					Namespace: Namespace,
					Labels: map[string]string{
						"app": ZkClusterName,
					},
				},
				Spec: corev1.PodSpec{
					Hostname: Hostname,
				},
			}
			s.AddKnownTypes(v1beta1.GroupVersion, zkCl, zkBk)
		})

		When("ZK cluster isn't deployed", func() {
			var (
				errBk error
			)

			BeforeEach(func() {
				zkBk.WithDefaults()
				cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(zkBk).Build()
				rBk = &ZookeeperBackupReconciler{Client: cl, Scheme: s, LeaderGetter: MockGetLeader}
				_, errBk = rBk.Reconcile(context.TODO(), req)
			})

			It("should raise an error", func() {
				Ω(errBk).To(HaveOccurred())
			})

		})

		When("ZK Cluster is deployed", func() {
			var (
				resCl reconcile.Result
				reqCl reconcile.Request
				errCl error
			)

			Context("and empty backup spec", func() {
				var (
					errBk error
					err   error
				)

				BeforeEach(func() {
					zkCl.WithDefaults()
					cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(zkCl, zkBk).Build()
					rCl = &ZookeeperClusterReconciler{Client: cl, Scheme: s, ZkClient: mockZkClient}
					reqCl = reconcile.Request{NamespacedName: types.NamespacedName{Name: ZkClusterName, Namespace: Namespace}}
					resCl, errCl = rCl.Reconcile(context.TODO(), reqCl)
					Ω(errCl).To(BeNil())
					Ω(resCl.RequeueAfter).To(Equal(ReconcileTime))
					rBk = &ZookeeperBackupReconciler{Client: cl, Scheme: s, LeaderGetter: MockGetLeader}
					res, errBk = rBk.Reconcile(context.TODO(), req)
				})

				It("shouldn't error", func() {
					Ω(errBk).To(BeNil())
				})

				It("should set the default spec options", func() {
					foundZkBk := &v1beta1.ZookeeperBackup{}
					err = cl.Get(context.TODO(), req.NamespacedName, foundZkBk)
					Ω(err).To(BeNil())
					Ω(foundZkBk.Spec.BackupsToKeep).To(BeEquivalentTo("7"))
				})

				It("should requeue the request", func() {
					Ω(res.Requeue).To(BeTrue())
				})
			})

			Context("with default backup specs", func() {
				var (
					errBk   error
					err     error
					jobSpec *batchv1beta1.CronJob
				)

				BeforeEach(func() {
					jobSpec = newCronJobForCR(zkBk)

					zkCl.WithDefaults()
					zkCl.Status.ReadyReplicas = 3
					zkBk.WithDefaults()
					cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(zkCl, zkBk, leaderPod).Build()
					rCl = &ZookeeperClusterReconciler{Client: cl, Scheme: s, ZkClient: mockZkClient}
					reqCl = reconcile.Request{NamespacedName: types.NamespacedName{Name: ZkClusterName, Namespace: Namespace}}
					resCl, errCl = rCl.Reconcile(context.TODO(), reqCl)
					Ω(errCl).To(BeNil())
					Ω(resCl.RequeueAfter).To(Equal(ReconcileTime))
					rBk = &ZookeeperBackupReconciler{Client: cl, Scheme: s, LeaderGetter: MockGetLeader}
					res, errBk = rBk.Reconcile(context.TODO(), req)
				})

				It("shouldn't error", func() {
					Ω(errBk).To(BeNil())
				})

				It("should create PVC", func() {
					pvcSpec := newPVCForZookeeperBackup(zkBk)
					foundPVC := &corev1.PersistentVolumeClaim{}
					err = cl.Get(context.TODO(), types.NamespacedName{Name: pvcSpec.Name, Namespace: Namespace}, foundPVC)
					Ω(err).To(BeNil())
				})

				It("should create CronJob", func() {
					foundCronJob := &batchv1beta1.CronJob{}
					err = cl.Get(context.TODO(), types.NamespacedName{Name: jobSpec.Name, Namespace: Namespace}, foundCronJob)
					Ω(err).To(BeNil())
					Ω(*foundCronJob.Spec.Suspend).Should(BeFalse())
				})

				It("should requeue after ReconcileTime delay", func() {
					Ω(res.RequeueAfter).To(Equal(ReconcileTime))
				})

				Context("pvc parameters have changed", func() {

					BeforeEach(func() {
						zkBk.Spec.DataStorageClass = "backup-hdd"
						err = cl.Update(context.TODO(), zkBk)
						Ω(err).To(BeNil())
						res, errBk = rBk.Reconcile(context.TODO(), req)
						Ω(errBk).To(BeNil())
					})

					It("should update pvc", func() {
						pvcSpec := newPVCForZookeeperBackup(zkBk)
						foundPVC := &corev1.PersistentVolumeClaim{}
						err = cl.Get(context.TODO(), types.NamespacedName{Name: pvcSpec.Name, Namespace: Namespace}, foundPVC)
						Ω(err).To(BeNil())
						Ω(*foundPVC.Spec.StorageClassName).Should(BeEquivalentTo(zkBk.Spec.DataStorageClass))
					})

				})

				Context("cronJob parameters have changed", func() {

					BeforeEach(func() {
						zkBk.Spec.Schedule = "0 12 */1 * *"
						zkBk.Spec.BackupsToKeep = "15"
						err = cl.Update(context.TODO(), zkBk)
						Ω(err).To(BeNil())
						res, errBk = rBk.Reconcile(context.TODO(), req)
						Ω(errBk).To(BeNil())
					})

					It("should update cronJob", func() {
						foundCronJob := &batchv1beta1.CronJob{}
						err = cl.Get(context.TODO(), types.NamespacedName{Name: jobSpec.Name, Namespace: Namespace}, foundCronJob)
						Ω(err).To(BeNil())
						envVars := foundCronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env
						for _, envVar := range envVars {
							if envVar.Name == "BACKUPS_TO_KEEP" {
								Ω(envVar.Value).Should(BeEquivalentTo(zkBk.Spec.BackupsToKeep))
							}
						}
						Ω(foundCronJob.Spec.Schedule).Should(BeEquivalentTo(zkBk.Spec.Schedule))
						Ω(*foundCronJob.Spec.Suspend).Should(BeFalse())
					})
				})

				Context("ZK cluster becomes not Ready", func() {

					BeforeEach(func() {
						err = cl.Get(context.TODO(), types.NamespacedName{Name: zkCl.Name, Namespace: zkCl.Namespace}, zkCl)
						Ω(err).To(BeNil())
						zkCl.Status.Replicas = 3
						zkCl.Status.ReadyReplicas = 2
						err = cl.Update(context.TODO(), zkCl)
						Ω(err).To(BeNil())
						res, errBk = rBk.Reconcile(context.TODO(), req)
						Ω(errBk).To(BeNil())
					})

					It("should suspend cronJob", func() {
						foundCronJob := &batchv1beta1.CronJob{}
						err = cl.Get(context.TODO(), types.NamespacedName{Name: jobSpec.Name, Namespace: Namespace}, foundCronJob)
						Ω(err).To(BeNil())
						Ω(*foundCronJob.Spec.Suspend).Should(BeTrue())
					})
				})

				Context("can't find pod with leader", func() {
					BeforeEach(func() {
						err = cl.Delete(context.TODO(), leaderPod)
						Ω(err).To(BeNil())
						res, errBk = rBk.Reconcile(context.TODO(), req)
						Ω(errBk).To(BeNil())
					})

					It("should suspend cronJob", func() {
						foundCronJob := &batchv1beta1.CronJob{}
						err = cl.Get(context.TODO(), types.NamespacedName{Name: jobSpec.Name, Namespace: Namespace}, foundCronJob)
						Ω(err).To(BeNil())
						Ω(*foundCronJob.Spec.Suspend).Should(BeTrue())
					})
				})
			})
		})

		Context("Checking result when request namespace does not contains zookeeper backup", func() {
			var (
				errBk error
			)

			BeforeEach(func() {
				cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(zkBk).Build()
				rBk = &ZookeeperBackupReconciler{Client: cl, Scheme: s, LeaderGetter: MockGetLeader}
				req.NamespacedName.Namespace = "temp"
				res, errBk = rBk.Reconcile(context.TODO(), req)
			})
			It("should have false in reconcile result", func() {
				Ω(res.Requeue).To(BeFalse())
				Ω(errBk).To(BeNil())
			})
		})

		Context("ZK backup isn't registered", func() {
			var (
				errBk error
			)
			cl = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			rBk = &ZookeeperBackupReconciler{Client: cl, Scheme: s, LeaderGetter: MockGetLeader}
			res, errBk = rBk.Reconcile(context.TODO(), req)

			It("should raise an error and shouldn't requeue the request", func() {
				Ω(errBk).To(HaveOccurred())
				Ω(res.Requeue).To(BeFalse())
			})
		})
	})
})
