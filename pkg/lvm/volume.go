// Copyright © 2019 The OpenEBS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lvm

import (
	"os"

	apis "github.com/pawanpraka1/dynamic-lvm/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/pawanpraka1/dynamic-lvm/pkg/builder/volbuilder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	// LvmNamespaceKey is the environment variable to get openebs namespace
	//
	// This environment variable is set via kubernetes downward API
	LvmNamespaceKey string = "LVM_NAMESPACE"
	// GoogleAnalyticsKey This environment variable is set via env
	GoogleAnalyticsKey string = "OPENEBS_IO_ENABLE_ANALYTICS"
	// LVMFinalizer for the LVMVolume CR
	LVMFinalizer string = "lvm.openebs.io/finalizer"
	// VolGroupKey is key for LVM group name
	VolGroupKey string = "openebs.io/volgroup"
	// LVMNodeKey will be used to insert Label in LVMVolume CR
	LVMNodeKey string = "kubernetes.io/nodename"
	// LVMTopologyKey is supported topology key for the lvm driver
	LVMTopologyKey string = "openebs.io/nodename"
	// LVMStatusPending shows object has not handled yet
	LVMStatusPending string = "Pending"
	// LVMStatusFailed shows object operation has failed
	LVMStatusFailed string = "Failed"
	// LVMStatusReady shows object has been processed
	LVMStatusReady string = "Ready"
)

var (
	// LvmNamespace is openebs system namespace
	LvmNamespace string

	// NodeID is the NodeID of the node on which the pod is present
	NodeID string

	// GoogleAnalyticsEnabled should send google analytics or not
	GoogleAnalyticsEnabled string
)

func init() {

	LvmNamespace = os.Getenv(LvmNamespaceKey)
	if LvmNamespace == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		klog.Fatalf("LVM_NAMESPACE environment variable not set")
	}
	NodeID = os.Getenv("OPENEBS_NODE_ID")
	if NodeID == "" && os.Getenv("OPENEBS_NODE_DRIVER") != "" {
		klog.Fatalf("NodeID environment variable not set")
	}

	GoogleAnalyticsEnabled = os.Getenv(GoogleAnalyticsKey)
}

// ProvisionVolume creates a LVMVolume CR,
// watcher for volume is present in CSI agent
func ProvisionVolume(
	vol *apis.LVMVolume,
) error {

	_, err := volbuilder.NewKubeclient().WithNamespace(LvmNamespace).Create(vol)
	if err == nil {
		klog.Infof("provisioned volume %s", vol.Name)
	}

	return err
}

// DeleteVolume deletes the corresponding LVM Volume CR
func DeleteVolume(volumeID string) (err error) {
	err = volbuilder.NewKubeclient().WithNamespace(LvmNamespace).Delete(volumeID)
	if err == nil {
		klog.Infof("deprovisioned volume %s", volumeID)
	}

	return
}

// GetLVMVolume fetches the given LVMVolume
func GetLVMVolume(volumeID string) (*apis.LVMVolume, error) {
	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(LvmNamespace).Get(volumeID, getOptions)
	return vol, err
}

// GetLVMVolumeState returns LVMVolume OwnerNode and State for
// the given volume. CreateVolume request may call it again and
// again until volume is "Ready".
func GetLVMVolumeState(volID string) (string, string, error) {
	getOptions := metav1.GetOptions{}
	vol, err := volbuilder.NewKubeclient().
		WithNamespace(LvmNamespace).Get(volID, getOptions)

	if err != nil {
		return "", "", err
	}

	return vol.Spec.OwnerNodeID, vol.Status.State, nil
}

// UpdateVolInfo updates LVMVolume CR with node id and finalizer
func UpdateVolInfo(vol *apis.LVMVolume) error {
	finalizers := []string{LVMFinalizer}
	labels := map[string]string{LVMNodeKey: NodeID}

	if vol.Finalizers != nil {
		return nil
	}

	newVol, err := volbuilder.BuildFrom(vol).
		WithFinalizer(finalizers).
		WithVolumeStatus(LVMStatusReady).
		WithLabels(labels).Build()

	if err != nil {
		return err
	}

	_, err = volbuilder.NewKubeclient().WithNamespace(LvmNamespace).Update(newVol)
	return err
}

// RemoveVolFinalizer adds finalizer to LVMVolume CR
func RemoveVolFinalizer(vol *apis.LVMVolume) error {
	vol.Finalizers = nil

	_, err := volbuilder.NewKubeclient().WithNamespace(LvmNamespace).Update(vol)
	return err
}
