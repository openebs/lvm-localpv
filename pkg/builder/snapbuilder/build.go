/*
Copyright 2021 The OpenEBS Authors

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

package snapbuilder

import (
	"github.com/openebs/lib-csi/pkg/common/errors"
	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
)

// Builder is the builder object for LVMSnapshot
type Builder struct {
	snap *LVMSnapshot
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		snap: &LVMSnapshot{
			Object: &apis.LVMSnapshot{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(snap *apis.LVMSnapshot) *Builder {
	if snap == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build snap object: nil snap"),
		)
		return b
	}
	return &Builder{
		snap: &LVMSnapshot{
			Object: snap,
		},
	}
}

// WithNamespace sets the namespace of  LVMSnapshot
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi snap object: missing namespace",
			),
		)
		return b
	}
	b.snap.Object.Namespace = namespace
	return b
}

// WithName sets the name of LVMSnapshot
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build csi snap object: missing name",
			),
		)
		return b
	}
	b.snap.Object.Name = name
	return b
}

// WithLabels merges existing labels if any
// with the ones that are provided here
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		return b
	}

	if b.snap.Object.Labels == nil {
		b.snap.Object.Labels = map[string]string{}
	}

	for key, value := range labels {
		b.snap.Object.Labels[key] = value
	}
	return b
}

// WithFinalizer merge existing finalizers if any
// with the ones that are provided here
func (b *Builder) WithFinalizer(finalizer []string) *Builder {
	b.snap.Object.Finalizers = append(b.snap.Object.Finalizers, finalizer...)
	return b
}

// WithCapacity sets the Capacity of lvm snapshot
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
	b.snap.Object.Spec.Capacity = capacity
	return b
}

// Build returns LVMSnapshot API object
func (b *Builder) Build() (*apis.LVMSnapshot, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.snap.Object, nil
}
