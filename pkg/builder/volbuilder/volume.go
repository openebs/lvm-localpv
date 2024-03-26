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

package volbuilder

import (
	"strconv"

	"github.com/openebs/lib-csi/pkg/common/errors"
	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Builder is the builder object for LVMVolume
type Builder struct {
	volume *LVMVolume
	errs   []error
}

// LVMVolume is a wrapper over
// LVMVolume API instance
type LVMVolume struct {
	// LVMVolume object
	Object *apis.LVMVolume
}

// From returns a new instance of
// lvm volume
func From(vol *apis.LVMVolume) *LVMVolume {
	return &LVMVolume{
		Object: vol,
	}
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		volume: &LVMVolume{
			Object: &apis.LVMVolume{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(volume *apis.LVMVolume) *Builder {
	if volume == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil volume"),
		)
		return b
	}
	return &Builder{
		volume: &LVMVolume{
			Object: volume,
		},
	}
}

// WithNamespace sets the namespace of  LVMVolume
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm volume object: missing namespace",
			),
		)
		return b
	}
	b.volume.Object.Namespace = namespace
	return b
}

// WithName sets the name of LVMVolume
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm volume object: missing name",
			),
		)
		return b
	}
	b.volume.Object.Name = name
	return b
}

// WithCapacity sets the Capacity of lvm volume by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	if capacity == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm volume object: missing capacity",
			),
		)
		return b
	}
	b.volume.Object.Spec.Capacity = capacity
	return b
}

// WithOwnerNode sets owner node for the LVMVolume where the volume should be provisioned
func (b *Builder) WithOwnerNode(host string) *Builder {
	b.volume.Object.Spec.OwnerNodeID = host
	return b
}

// WithVolumeStatus sets LVMVolume status
func (b *Builder) WithVolumeStatus(status string) *Builder {
	b.volume.Object.Status.State = status
	return b
}

// WithShared sets where filesystem is shared or not
func (b *Builder) WithShared(shared string) *Builder {
	b.volume.Object.Spec.Shared = shared
	return b
}

// WithThinProvision sets where thinProvision is enable or not
func (b *Builder) WithThinProvision(thinProvision string) *Builder {
	b.volume.Object.Spec.ThinProvision = thinProvision
	return b
}

// WithRaidType sets the RAID type for all logical volumes in the volume group
func (b *Builder) WithRaidType(raidType string) *Builder {
	b.volume.Object.Spec.RaidType = raidType
	return b
}

// WithIntegrity sets where integrity is enable or not
func (b *Builder) WithIntegrity(integrity string) *Builder {
	b.volume.Object.Spec.Integrity = integrity
	return b
}

// WithMirrors sets the RAID mirror count
func (b *Builder) WithMirrors(mirrors string) *Builder {
	mirrorCount, err := strconv.ParseUint(mirrors, 10, 32)
	if err != nil {
		b.errs = append(
			b.errs,
			errors.New(
				"invalid mirror count: must be a positive 32-bit integer",
			),
			err,
		)
	}

	b.volume.Object.Spec.Mirrors = uint(mirrorCount)
	return b
}

// WithNoSync sets the `nosync` RAID option
func (b *Builder) WithNoSync(nosync string) *Builder {
	b.volume.Object.Spec.NoSync = nosync
	return b
}

func (b *Builder) WithStripeCount(stripes string) *Builder {
	stripeCount, err := strconv.ParseUint(stripes, 10, 32)
	if err != nil {
		b.errs = append(
			b.errs,
			errors.New(
				"invalid stripe count: must be a positive 32-bit integer",
			),
			err,
		)
	}

	b.volume.Object.Spec.StripeCount = uint(stripeCount)
	return b
}

// WithStripeSize sets the size of each stripe for a RAID volume
func (b *Builder) WithStripeSize(size string) *Builder {
	stripeSize, err := resource.ParseQuantity(size)
	if err != nil {
		b.errs = append(
			b.errs,
			errors.New(
				"invalid stripe size",
			),
			err,
		)
	}

	value := stripeSize.Value()
	if value < 0 {
		b.errs = append(b.errs, errors.New("invalid stripe size: value must be positive"))
	}

	b.volume.Object.Spec.StripeSize = uint(value)
	return b
}

// WithLvCreateOptions sets any additional LVM options used when creating a volume
func (b *Builder) WithLvCreateOptions(options string) *Builder {
	b.volume.Object.Spec.LvCreateOptions = options
	return b
}

// WithVolGroup sets volume group name for creating volume
func (b *Builder) WithVolGroup(vg string) *Builder {
	if vg == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm volume object: missing vg name",
			),
		)
		return b
	}
	b.volume.Object.Spec.VolGroup = vg
	return b
}

// WithVgPattern sets volume group regex pattern.
func (b *Builder) WithVgPattern(pattern string) *Builder {
	if pattern == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm volume object: missing vg name",
			),
		)
		return b
	}
	b.volume.Object.Spec.VgPattern = pattern
	return b
}

// WithNodeName sets NodeID for creating the volume
func (b *Builder) WithNodeName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm volume object: missing node name",
			),
		)
		return b
	}
	b.volume.Object.Spec.OwnerNodeID = name
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.volume.Object.Labels == nil {
		b.volume.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.volume.Object.Labels[key] = value
	}
	return b
}

// WithFinalizer sets Finalizer name creating the volume
func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.volume.Object.Finalizers = append(b.volume.Object.Finalizers, finalizer...)
	return b
}

// Build returns LVMVolume API object
func (b *Builder) Build() (*apis.LVMVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.volume.Object, nil
}
