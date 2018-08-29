package zk

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDeploy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deploy ZookeeperCluster Spec")
}

var _ = Describe("Deploy", func() {
	var z *v1beta1.ZookeeperCluster
	name := "example"

	BeforeEach(func() {
		z = &v1beta1.ZookeeperCluster{ObjectMeta: metav1.ObjectMeta{Name: name}}
	})

	Context("with all defaults", func() {

		configMapName := "example-configmap"
		zkPorts := zkPorts{
			Client: 2181,
			Quorum: 2888,
			Leader: 3888,
		}

		BeforeEach(func() {
			z.WithDefaults()
		})

		Context("#makeZkSts", func() {
			var sts *appsv1.StatefulSet
			serviceName := "example-headless"

			BeforeEach(func() {
				sts = makeZkSts(configMapName, zkPorts, z)
			})

			It("should generate a stateful-set", func() {
				Expect(sts.TypeMeta.Kind).To(Equal("StatefulSet"))
			})

			It("should have an app label", func() {
				Expect(sts.ObjectMeta.Labels["app"]).To(Equal(name))
			})

			It("should have a service name", func() {
				Expect(sts.Spec.ServiceName).To(Equal(serviceName))
			})

			It("should have a selector", func() {
				Expect(sts.Spec.Selector.MatchLabels["app"]).To(Equal(name))
			})

			It("should have an update strategy", func() {
				Expect(sts.Spec.UpdateStrategy.Type).To(BeEquivalentTo(appsv1.RollingUpdateStatefulSetStrategyType))
			})

			It("should have a replica count", func() {
				Expect(*sts.Spec.Replicas).To(BeEquivalentTo(3))
			})

			It("should have a pod management policy", func() {
				Expect(sts.Spec.PodManagementPolicy).To(Equal(appsv1.OrderedReadyPodManagement))
			})

			Context("pod template spec", func() {
				var p v1.PodTemplateSpec

				BeforeEach(func() {
					p = sts.Spec.Template
				})

				It("should have a generate name field", func() {
					Expect(p.ObjectMeta.GenerateName).To(Equal(name))
				})

				It("should have an app label", func() {
					Expect(p.ObjectMeta.Labels["app"]).To(Equal(name))
				})

				It("should have an affinity rule", func() {
					Expect(p.Spec.Affinity).To(Equal(z.Spec.Pod.Affinity))
				})

				It("should have a termination grace period", func() {
					Expect(*p.Spec.TerminationGracePeriodSeconds).
						To(BeEquivalentTo(z.Spec.Pod.TerminationGracePeriodSeconds))
				})

				Context("zookeeper container", func() {
					var c v1.Container

					BeforeEach(func() {
						c = p.Spec.Containers[0]
					})

					It("should be named 'zookeeper'", func() {
						Expect(c.Name).To(Equal("zookeeper"))
					})

					It("should have the default image", func() {
						Expect(c.Image).
							To(Equal(v1beta1.DefaultZkContainerRepository + ":" + v1beta1.DefaultZkContainerVersion))
					})

					It("should have the default pull policy", func() {
						Expect(c.ImagePullPolicy).To(BeEquivalentTo(v1.PullAlways))
					})

					It("should have the zookeeper ports", func() {
						for name, number := range map[string]int{
							"client":          2181,
							"quorum":          2888,
							"leader-election": 3888,
						} {
							port, err := portByName(c.Ports, name)
							Expect(err).To(BeNil())
							Expect(port.ContainerPort).To(BeEquivalentTo(number))
						}
					})

					Context("readiness probe", func() {
						var probe v1.Probe

						BeforeEach(func() {
							probe = *c.ReadinessProbe
						})

						It("should use an exec command", func() {
							Expect(probe.Exec.Command).To(BeEquivalentTo([]string{"zookeeperReady.sh"}))
						})

						It("should have a delay of 10 seconds", func() {
							Expect(probe.InitialDelaySeconds).To(BeEquivalentTo(10))
						})

						It("should have a timeout of 10 seconds", func() {
							Expect(probe.TimeoutSeconds).To(BeEquivalentTo(10))
						})

					})

					Context("liveness probe", func() {
						var probe v1.Probe

						BeforeEach(func() {
							probe = *c.LivenessProbe
						})

						It("should use an exec command", func() {
							Expect(probe.Exec.Command).To(BeEquivalentTo([]string{"zookeeperLive.sh"}))
						})

						It("should have a delay of 10 seconds", func() {
							Expect(probe.InitialDelaySeconds).To(BeEquivalentTo(10))
						})

						It("should have a timeout of 10 seconds", func() {
							Expect(probe.TimeoutSeconds).To(BeEquivalentTo(10))
						})

					})

					Context("volume mounts", func() {

						It("should have two mounts", func() {
							Expect(len(c.VolumeMounts)).To(Equal(2))
						})

						Context("data mount", func() {
							var m v1.VolumeMount
							var err error

							BeforeEach(func() {
								m, err = mountByName(c.VolumeMounts, "data")
							})

							It("should be named log", func() {
								Expect(err).To(BeNil())
								Expect(m.Name).To(Equal("data"))
							})

							It("should be mounted under /data", func() {
								Expect(err).To(BeNil())
								Expect(m.MountPath).To(Equal("/data"))
							})
						})

						Context("conf mount", func() {
							var m v1.VolumeMount
							var err error

							BeforeEach(func() {
								m, err = mountByName(c.VolumeMounts, "conf")
							})

							It("should be named log", func() {
								Expect(err).To(BeNil())
								Expect(m.Name).To(Equal("conf"))
							})

							It("should be mounted under /conf", func() {
								Expect(err).To(BeNil())
								Expect(m.MountPath).To(Equal("/conf"))
							})
						})

					})

				})
			})

			Context("volume claim template", func() {
				var v v1.PersistentVolumeClaim

				BeforeEach(func() {
					v = sts.Spec.VolumeClaimTemplates[0]
				})

				It("should be named data", func() {
					Expect(v.ObjectMeta.Name).To(Equal("data"))
				})

				It("should have an app label", func() {
					Expect(v.ObjectMeta.Labels["app"]).To(Equal(z.Name))
				})

			})
		})

	})
})

func portByName(ports []v1.ContainerPort, name string) (port v1.ContainerPort, err error) {
	for _, port := range ports {
		if port.Name == name {
			return port, err
		}
	}
	return port, fmt.Errorf("port not found")
}

func mountByName(mounts []v1.VolumeMount, name string) (mount v1.VolumeMount, err error) {
	for _, mount := range mounts {
		if mount.Name == name {
			return mount, err
		}
	}
	return mount, fmt.Errorf("mount not found")
}
