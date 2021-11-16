module github.com/pravega/zookeeper-operator

go 1.16

require (
	github.com/Azure/go-autorest/autorest v0.11.21 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.16 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/Microsoft/hcsshim v0.8.10-0.20200715222032-5eafd1556990 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/containerd v1.4.11 // indirect
	github.com/containerd/ttrpc v1.0.2 // indirect
	github.com/docker/docker v20.10.2+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.4.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/uuid v1.1.5 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/operator-framework/api v0.10.7 // indirect
	github.com/operator-framework/operator-lib v0.7.0
	github.com/operator-framework/operator-registry v1.13.9 // indirect
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20211007125505-59d4e928ea9d // indirect
	golang.org/x/oauth2 v0.0.0-20211005180243-6b3c2da341f1 // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/tools v0.1.7 // indirect
	google.golang.org/genproto v0.0.0-20211007155348-82e027067bd4 // indirect
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2 // indirect
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20210929172449-94abcedd1aa4 // indirect
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b // indirect
	sigs.k8s.io/controller-runtime v0.10.2
)

replace (
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.4.0

	k8s.io/api => k8s.io/api v0.19.13

	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.13

	k8s.io/apimachinery => k8s.io/apimachinery v0.19.14-rc.0

	k8s.io/apiserver => k8s.io/apiserver v0.19.13

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.13

	k8s.io/client-go => k8s.io/client-go v0.19.13

	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.13

	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.13

	k8s.io/code-generator => k8s.io/code-generator v0.19.14-rc.0

	k8s.io/component-base => k8s.io/component-base v0.19.13

	k8s.io/controller-manager => k8s.io/controller-manager v0.19.14-rc.0

	k8s.io/cri-api => k8s.io/cri-api v0.19.14-rc.0

	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.13

	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.13

	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.13

	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.13

	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.13

	k8s.io/kubectl => k8s.io/kubectl v0.19.13

	k8s.io/kubelet => k8s.io/kubelet v0.19.13

	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.13

	k8s.io/metrics => k8s.io/metrics v0.19.13

	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.13

	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.19.13

	k8s.io/sample-controller => k8s.io/sample-controller v0.19.13
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.5
)
