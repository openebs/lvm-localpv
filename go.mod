module github.com/openebs/lvm-localpv

go 1.14

replace k8s.io/api => k8s.io/api v0.15.12

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.15.12

replace k8s.io/apimachinery => k8s.io/apimachinery v0.15.13-beta.0

replace k8s.io/apiserver => k8s.io/apiserver v0.15.12

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.15.12

replace k8s.io/client-go => k8s.io/client-go v0.15.12

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.15.12

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.15.12

replace k8s.io/code-generator => k8s.io/code-generator v0.15.13-beta.0

replace k8s.io/component-base => k8s.io/component-base v0.15.12

replace k8s.io/cri-api => k8s.io/cri-api v0.15.13-beta.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.15.12

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.15.12

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.15.12

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.15.12

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.15.12

replace k8s.io/kubectl => k8s.io/kubectl v0.15.13-beta.0

replace k8s.io/kubelet => k8s.io/kubelet v0.15.12

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.15.12

replace k8s.io/metrics => k8s.io/metrics v0.15.12

replace k8s.io/node-api => k8s.io/node-api v0.15.12

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.15.12

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.15.12

replace k8s.io/sample-controller => k8s.io/sample-controller v0.15.12

require (
	github.com/container-storage-interface/spec v1.2.0
	github.com/docker/go-units v0.3.3
	github.com/ghodss/yaml v0.0.0-20180820084758-c7ce16629ff4
	github.com/golang/protobuf v1.4.2
	github.com/jpillora/go-ogle-analytics v0.0.0-20161213085824-14b04e0594ef
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/openebs/lib-csi v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.0-20180319062004-c439c4fa0937
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd
	google.golang.org/grpc v1.34.2
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.15.12
	k8s.io/klog v1.0.0
	k8s.io/kubernetes v1.15.12
	sigs.k8s.io/controller-runtime v0.2.0
)
