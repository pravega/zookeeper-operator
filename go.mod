module github.com/pravega/zookeeper-operator

go 1.16

require (
	github.com/Azure/go-autorest/autorest v0.11.19 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.14 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/Microsoft/hcsshim v0.8.10-0.20200715222032-5eafd1556990 // indirect
	github.com/containerd/containerd v1.4.8 // indirect
	github.com/containerd/ttrpc v1.0.2 // indirect
	github.com/docker/docker v20.10.2+incompatible // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.4.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.1.5 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/operator-framework/api v0.3.25 // indirect
	github.com/operator-framework/operator-registry v1.13.9 // indirect
	github.com/operator-framework/operator-sdk v0.19.4
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/rogpeppe/go-internal v1.5.2 // indirect
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414
	github.com/sirupsen/logrus v1.7.1
	github.com/spf13/cobra v1.1.3 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210716203947-853a461950ff // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20210716133855-ce7ef5c701ea // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.19.13
	k8s.io/apiextensions-apiserver v0.19.13 // indirect
	k8s.io/apimachinery v0.19.13
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20210527164424-3c818078ee3d // indirect
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477 // indirect
	sigs.k8s.io/controller-runtime v0.6.5
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
)
