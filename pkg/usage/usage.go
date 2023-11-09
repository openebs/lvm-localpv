/*
Copyright 2020 The OpenEBS Authors.

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

package usage

import (
	"strconv"

	ga4Client "github.com/openebs/go-ogle-analytics/client"
	ga4Event "github.com/openebs/go-ogle-analytics/event"
	k8sapi "github.com/openebs/lib-csi/pkg/client/k8s"
)

// Usage struct represents all information about a usage metric sent to
// Google Analytics with respect to the application
type Usage struct {
	// OpenebsEventBuilder to build the OpenEBSEvent
	OpenebsEventBuilder *ga4Event.OpenebsEventBuilder

	// GA4 Analytics Client
	AnalyticsClient *ga4Client.MeasurementClient
}

// New returns an instance of Usage
func New() *Usage {
	client, err := ga4Client.NewMeasurementClient(
		ga4Client.WithApiSecret("XXXXX"),
		ga4Client.WithMeasurementId("G-XXXXXX"),
	)
	if err != nil {
		return nil
	}
	openebsEventBuilder := ga4Event.NewOpenebsEventBuilder()
	return &Usage{AnalyticsClient: client, OpenebsEventBuilder: openebsEventBuilder}
}

// SetVolumeName i.e pv name
func (u *Usage) SetVolumeName(name string) *Usage {
	u.OpenebsEventBuilder.VolumeName(name)
	return u
}

// SetVolumeName i.e pvc name
func (u *Usage) SetVolumeClaimName(name string) *Usage {
	u.OpenebsEventBuilder.VolumeClaimName(name)
	return u
}

// SetCategory sets the category of an event
func (u *Usage) SetCategory(c string) *Usage {
	u.OpenebsEventBuilder.Category(c)
	return u
}

// SetAction sets the action of an event
func (u *Usage) SetAction(a string) *Usage {
	u.OpenebsEventBuilder.Action(a)
	return u
}

// SetLabel sets the label for an event
func (u *Usage) SetLabel(l string) *Usage {
	u.OpenebsEventBuilder.Label(l)
	return u
}

// SetValue sets the value for an event's label
func (u *Usage) SetValue(v string) *Usage {
	u.OpenebsEventBuilder.Value(v)
	return u
}

// SetVolumeCapacity sets the storage capacity of the volume for a volume event
func (u *Usage) SetVolumeCapacity(volCapG string) *Usage {
	s, _ := toGigaUnits(volCapG)
	u.SetValue(strconv.FormatInt(s, 10))
	return u
}

// SetReplicaCount Wrapper for setting replica count for volume events
// NOTE: This doesn't get the replica count in a volume de-provision event.
// TODO: Pick the current value of replica-count from the CAS-engine
func (u *Usage) SetReplicaCount(count, method string) *Usage {
	if method == VolumeProvision && count == "" {
		// Case: When volume-provision the replica count isn't specified
		// it is set to three by default by the m-apiserver
		u.OpenebsEventBuilder.Action(DefaultReplicaCount)
	} else {
		// Catch all case for volume-deprovision event and
		// volume-provision event with an overridden replica-count
		u.OpenebsEventBuilder.Action(Replica + count)
	}
	return u
}

// CommonBuild is a common builder method for Usage struct
func (u *Usage) CommonBuild() *Usage {
	v := NewVersion()
	_ = v.getVersion(false)

	u.OpenebsEventBuilder.
		Project(AppName).
		EngineInstaller(v.installerType).
		K8sVersion(v.k8sVersion).
		EngineVersion(v.openebsVersion).
		EngineInstaller(v.installerType).
		EngineName(DefaultCASType).
		NodeArch(v.nodeArch).
		NodeOs(v.nodeOs).
		NodeKernelVersion(v.nodeKernelVersion)

	return u
}

// ApplicationBuilder Application builder is used for adding k8s&openebs environment detail
// for non install events
func (u *Usage) ApplicationBuilder() *Usage {
	v := NewVersion()
	_ = v.getVersion(false)

	u.AnalyticsClient.SetClientId(v.id)
	u.OpenebsEventBuilder.K8sDefaultNsUid(v.id)

	return u
}

// InstallBuilder is a concrete builder for install events
func (u *Usage) InstallBuilder(override bool) *Usage {
	v := NewVersion()
	clusterSize, _ := k8sapi.NumberOfNodes()
	_ = v.getVersion(override)

	u.AnalyticsClient.SetClientId(v.id)
	u.OpenebsEventBuilder.
		K8sDefaultNsUid(v.id).
		Category(InstallEvent).
		Action(RunningStatus).
		Label(EventLabelNode).
		Value(strconv.Itoa(clusterSize))

	return u
}
