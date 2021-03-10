/*
Copyright 2019 The OpenEBS Authors

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

package nodebuilder

import (
	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Builder is the builder object for LVMNode
type Builder struct {
	node *LVMNode
	errs []error
}

// LVMNode is a wrapper over
// LVMNode API instance
type LVMNode struct {
	// LVMVolume object
	Object *apis.LVMNode
}

// From returns a new instance of
// lvm volume
func From(node *apis.LVMNode) *LVMNode {
	return &LVMNode{
		Object: node,
	}
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{
		node: &LVMNode{
			Object: &apis.LVMNode{},
		},
	}
}

// BuildFrom returns new instance of Builder
// from the provided api instance
func BuildFrom(node *apis.LVMNode) *Builder {
	if node == nil {
		b := NewBuilder()
		b.errs = append(
			b.errs,
			errors.New("failed to build lvm node object: nil node"),
		)
		return b
	}
	return &Builder{
		node: &LVMNode{
			Object: node,
		},
	}
}

// WithNamespace sets the namespace of LVMNode
func (b *Builder) WithNamespace(namespace string) *Builder {
	if namespace == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm node object: missing namespace",
			),
		)
		return b
	}
	b.node.Object.Namespace = namespace
	return b
}

// WithName sets the name of LVMNode
func (b *Builder) WithName(name string) *Builder {
	if name == "" {
		b.errs = append(
			b.errs,
			errors.New(
				"failed to build lvm node object: missing name",
			),
		)
		return b
	}
	b.node.Object.Name = name
	return b
}

// WithVolumeGroups sets the volume groups of LVMNode
func (b *Builder) WithVolumeGroups(vgs []apis.VolumeGroup) *Builder {
	b.node.Object.VolumeGroups = vgs
	return b
}

// WithOwnerReferences sets the owner references of LVMNode
func (b *Builder) WithOwnerReferences(ownerRefs ...metav1.OwnerReference) *Builder {
	b.node.Object.OwnerReferences = ownerRefs
	return b
}

// Build returns LVMNode API object
func (b *Builder) Build() (*apis.LVMNode, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}

	return b.node.Object, nil
}
