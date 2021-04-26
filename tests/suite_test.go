/*
Copyright 2020 The OpenEBS Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openebs/lvm-localpv/pkg/builder/volbuilder"
	"github.com/openebs/lvm-localpv/tests/deploy"
	"github.com/openebs/lvm-localpv/tests/pod"
	"github.com/openebs/lvm-localpv/tests/pv"
	"github.com/openebs/lvm-localpv/tests/pvc"
	"github.com/openebs/lvm-localpv/tests/sc"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/klog"

	// auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	// volume group name where volume provisioning will happen
	VOLGROUP = "lvmvg"
)

var (
	LVMClient        *volbuilder.Kubeclient
	SCClient         *sc.Kubeclient
	PVCClient        *pvc.Kubeclient
	PVClient         *pv.Kubeclient
	DeployClient     *deploy.Kubeclient
	PodClient        *pod.KubeClient
	nsName           = "lvm"
	scName           = "lvmpv-sc"
	LocalProvisioner = "local.csi.openebs.io"
	pvcName          = "lvmpv-pvc"
	snapName         = "lvmpv-snap"
	appName          = "busybox-lvmpv"

	nsObj            *corev1.Namespace
	scObj            *storagev1.StorageClass
	deployObj        *appsv1.Deployment
	pvcObj           *corev1.PersistentVolumeClaim
	appPod           *corev1.PodList
	accessModes      = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	capacity         = "5368709120" // 5Gi
	KubeConfigPath   string
	OpenEBSNamespace string
)

func init() {
	KubeConfigPath = os.Getenv("KUBECONFIG")

	OpenEBSNamespace = os.Getenv("LVM_NAMESPACE")
	if OpenEBSNamespace == "" {
		klog.Fatalf("LVM_NAMESPACE environment variable not set")
	}
	SCClient = sc.NewKubeClient(sc.WithKubeConfigPath(KubeConfigPath))
	PVCClient = pvc.NewKubeClient(pvc.WithKubeConfigPath(KubeConfigPath))
	DeployClient = deploy.NewKubeClient(deploy.WithKubeConfigPath(KubeConfigPath))
	PodClient = pod.NewKubeClient(pod.WithKubeConfigPath(KubeConfigPath))
	LVMClient = volbuilder.NewKubeclient(volbuilder.WithKubeConfigPath(KubeConfigPath))
}

func TestSource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test LVMPV volume provisioning")
}
