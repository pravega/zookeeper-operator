module github.com/q8s-io/zookeeper-operator-pravega

go 1.14

require (
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/pkg/errors v0.9.1
	github.com/q8s-io/statefulset-pingcap v0.4.1-0.20201120090223-e7e4a4323345
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/tools v0.0.0-20200331202046-9d5940d49312 // indirect
	k8s.io/api v0.17.5
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	k8s.io/api => k8s.io/api v0.17.5
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.5
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.17.5
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.5
	k8s.io/client-go => k8s.io/client-go v0.17.5
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.5
	k8s.io/code-generator => k8s.io/code-generator v0.17.6-beta.0
	k8s.io/component-base => k8s.io/component-base v0.17.5
	k8s.io/cri-api => k8s.io/cri-api v0.17.7-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.5
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.5
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.5
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.5
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.5
	k8s.io/kubectl => k8s.io/kubectl v0.17.5
	k8s.io/kubelet => k8s.io/kubelet v0.17.5
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.5
	k8s.io/metrics => k8s.io/metrics v0.17.5
	k8s.io/node-api => k8s.io/node-api v0.17.5
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.5
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.17.5
	k8s.io/sample-controller => k8s.io/sample-controller v0.17.5
)
