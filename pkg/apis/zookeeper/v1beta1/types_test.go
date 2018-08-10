package v1beta1

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
)

func TestDeploy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deploy ZookeeperCluster Spec")
}

var _ = Describe("Deploy", func() {

	var z ZookeeperCluster
	appName := "example"
	svcName := "example-headless"

	BeforeEach(func() {
		z = ZookeeperCluster{ObjectMeta: metav1.ObjectMeta{Name: appName}}
	})


	Context("with all defaults", func() {

		BeforeEach(func() {
			z.Spec.withDefaults(&z)
		})

		It("should have a defaults size of 3", func() {
			Expect(z.Spec.Size).To(BeEquivalentTo(3))
		})

		It("should have an app label", func() {
			Expect(z.Spec.Labels["app"]).To(Equal(appName))
		})

		Context("container image", func() {

			var i ContainerImage

			BeforeEach(func() {
				i = z.Spec.Image
			})

			It("should have the default image repo", func() {
				Expect(i.Repository).To(Equal(DefaultZkContainerRepository))
			})

			It("Should have the default container tag", func() {
				Expect(i.Tag).To(Equal(DefaultZkContainerVersion))
			})

			It("Should have the default container pull policy", func() {
				Expect(i.PullPolicy).To(BeEquivalentTo(DefaultZkContainerPolicy))
			})

		})

		Context("Zookeeper config", func() {

			var c ZookeeperConfig

			BeforeEach(func() { c = *z.Spec.Conf })

			It("should have an init limit of 10", func() {
				Expect(c.InitLimit).To(BeEquivalentTo(10))
			})

			It("should have a tick time of 2000", func() {
				Expect(c.TickTime).To(BeEquivalentTo(2000))
			})

			It("should have a sync limit of 2", func() {
				Expect(c.SyncLimit).To(BeEquivalentTo(2))
			})
		})

		Context("Ports", func() {

			It("should have a client port at 2181", func() {
				expectPortAndName("client", 2181, z.Spec.Ports)
			})

			It("should have a quorum port at 2888", func() {
				expectPortAndName("quorum", 2888, z.Spec.Ports)
			})

			It("should have a leader port at 3888", func() {
				expectPortAndName("leader-election", 3888, z.Spec.Ports)
			})

		})

		Context("Pod policy", func() {
			var p PodPolicy

			BeforeEach(func() {
				p = *z.Spec.Pod
			})

			It("should have an app label", func() {
				Expect(p.Labels["app"]).To(Equal(appName))
			})

			It("should have a termination grace period", func() {
				Expect(p.TerminationGracePeriodSeconds).
					To(BeEquivalentTo(DefaultTerminationGracePeriod))
			})

			Context("affinity", func() {

				var aff v1.PodAffinityTerm

				BeforeEach(func() {
					aff = p.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0]
				})

				It("should have a hostname topo key", func() {
					Expect(aff.TopologyKey).To(Equal("kubernetes.io/hostname"))
				})

				Context("first label selector", func() {
					var match metav1.LabelSelectorRequirement

					BeforeEach(func() {
						match = aff.LabelSelector.MatchExpressions[0]
					})

					It("should have key of 'app'", func() {
						Expect(match.Key).To(Equal("app"))
					})

					It("should have operator of 'in'", func() {
						Expect(match.Operator).To(Equal(metav1.LabelSelectorOpIn))
					})

					It("should have a value of the servicename", func() {
						Expect(match.Values).To(ContainElement(svcName))
					})
				})

			})
		})

		Context("Persistent volumes", func() {

			var p v1.PersistentVolumeClaimSpec

			BeforeEach(func() {
				p = *z.Spec.PersistentVolumeClaimSpec
			})

			It("should be an RWO claim", func() {
				Expect(p.AccessModes[0]).To(Equal(v1.ReadWriteOnce))
			})

			It("should request 20Gi as a default", func() {
				storageQty := p.Resources.Requests[v1.ResourceStorage]
				Expect(storageQty.ScaledValue(9)).To(BeEquivalentTo(20))
			})

		})

	})

	// TODO test non-default cases
})

func expectPortAndName(name string, portNum int32, ports []v1.ContainerPort)  {
	p := findPortByName(name, ports)
	Expect(p.Name).To(Equal(name))
	Expect(p.ContainerPort).To(BeEquivalentTo(portNum))
}

func findPortByName(name string, ports []v1.ContainerPort) v1.ContainerPort {
	for _, port := range ports {
		if port.Name == name {
			return port
		}
	}
	return v1.ContainerPort{}
}
