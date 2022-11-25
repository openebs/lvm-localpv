module github.com/openebs/lvm-localpv

go 1.19

replace google.golang.org/protobuf => google.golang.org/protobuf v1.25.0

replace k8s.io/api => k8s.io/api v0.20.2

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.2

replace k8s.io/apimachinery => k8s.io/apimachinery v0.20.2

replace k8s.io/apiserver => k8s.io/apiserver v0.20.2

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.2

replace k8s.io/client-go => k8s.io/client-go v0.20.2

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.2

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.2

replace k8s.io/code-generator => k8s.io/code-generator v0.20.2

replace k8s.io/component-base => k8s.io/component-base v0.20.2

replace k8s.io/cri-api => k8s.io/cri-api v0.20.2

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.2

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.2

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.2

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.2

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.2

replace k8s.io/kubectl => k8s.io/kubectl v0.20.2

replace k8s.io/kubelet => k8s.io/kubelet v0.20.2

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.2

replace k8s.io/metrics => k8s.io/metrics v0.20.2

replace k8s.io/node-api => k8s.io/node-api v0.20.2

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.2

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.2

replace k8s.io/sample-controller => k8s.io/sample-controller v0.20.2

require (
	github.com/container-storage-interface/spec v1.2.0
	github.com/docker/go-units v0.3.3
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3
	github.com/jpillora/go-ogle-analytics v0.0.0-20161213085824-14b04e0594ef
	github.com/kubernetes-csi/csi-lib-utils v0.9.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/openebs/lib-csi v0.6.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40
	google.golang.org/grpc v1.34.2
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.20.2
	k8s.io/klog v1.0.0
	//k8s.io/kubernetes v0.20.2
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.2.0
)

require (
	cloud.google.com/go v0.54.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/go-logr/logr v0.2.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/mailru/easyjson v0.7.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0 // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/text v0.3.4 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.0.0-20200616133436-c1934b75d054 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
	k8s.io/gengo v0.0.0-20201113003025-83324d819ded // indirect
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.0.2 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
