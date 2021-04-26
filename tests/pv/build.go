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

package v1alpha1

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Builder is the builder object for PV
type Builder struct {
	pv   *PV
	errs []error
}

// NewBuilder returns new instance of Builder
func NewBuilder() *Builder {
	return &Builder{pv: &PV{object: &corev1.PersistentVolume{}}}
}

// WithName sets the Name field of PV with provided value.
func (b *Builder) WithName(name string) *Builder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing PV name"))
		return b
	}
	b.pv.object.Name = name
	return b
}

// WithAnnotations sets the Annotations field of PV with provided arguments
func (b *Builder) WithAnnotations(annotations map[string]string) *Builder {
	if len(annotations) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing annotations"))
		return b
	}
	b.pv.object.Annotations = annotations
	return b
}

// WithLabels sets the Labels field of PV with provided arguments
func (b *Builder) WithLabels(labels map[string]string) *Builder {
	if len(labels) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing labels"))
		return b
	}
	b.pv.object.Labels = labels
	return b
}

// WithReclaimPolicy sets the PV ReclaimPolicy field with provided argument
func (b *Builder) WithReclaimPolicy(reclaimPolicy corev1.PersistentVolumeReclaimPolicy) *Builder {
	b.pv.object.Spec.PersistentVolumeReclaimPolicy = reclaimPolicy
	return b
}

// WithVolumeMode sets the VolumeMode field in PV with provided arguments
func (b *Builder) WithVolumeMode(volumeMode corev1.PersistentVolumeMode) *Builder {
	b.pv.object.Spec.VolumeMode = &volumeMode
	return b
}

// WithAccessModes sets the AccessMode field in PV with provided arguments
func (b *Builder) WithAccessModes(accessMode []corev1.PersistentVolumeAccessMode) *Builder {
	if len(accessMode) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing accessmodes"))
		return b
	}
	b.pv.object.Spec.AccessModes = accessMode
	return b
}

// WithCapacity sets the Capacity field in PV by converting string
// capacity into Quantity
func (b *Builder) WithCapacity(capacity string) *Builder {
	resCapacity, err := resource.ParseQuantity(capacity)
	if err != nil {
		b.errs = append(b.errs, errors.Wrapf(err, "failed to build PV object: failed to parse capacity {%s}", capacity))
		return b
	}
	return b.WithCapacityQty(resCapacity)
}

// WithCapacityQty sets the Capacity field in PV with provided arguments
func (b *Builder) WithCapacityQty(resCapacity resource.Quantity) *Builder {
	resourceList := corev1.ResourceList{
		corev1.ResourceName(corev1.ResourceStorage): resCapacity,
	}
	b.pv.object.Spec.Capacity = resourceList
	return b
}

// WithLocalHostDirectory sets the LocalVolumeSource field of PV with provided hostpath
func (b *Builder) WithLocalHostDirectory(path string) *Builder {
	return b.WithLocalHostPathFormat(path, "")
}

// WithLocalHostPathFormat sets the LocalVolumeSource field of PV with provided hostpath
// and request to format it with fstype - if not already formatted. A "" value for fstype
// indicates that the Local PV can determine the type of FS.
func (b *Builder) WithLocalHostPathFormat(path, fstype string) *Builder {
	if len(path) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing PV path"))
		return b
	}
	volumeSource := corev1.PersistentVolumeSource{
		Local: &corev1.LocalVolumeSource{
			Path:   path,
			FSType: &fstype,
		},
	}

	b.pv.object.Spec.PersistentVolumeSource = volumeSource
	return b
}

// WithPersistentVolumeSource sets the volume source field of PV with provided source
func (b *Builder) WithPersistentVolumeSource(source *corev1.PersistentVolumeSource) *Builder {
	if source == nil {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing PV source"))
		return b
	}
	b.pv.object.Spec.PersistentVolumeSource = *source
	return b
}

// WithNodeAffinity sets the NodeAffinity field of PV with provided node name
func (b *Builder) WithNodeAffinity(nodeName string) *Builder {
	if len(nodeName) == 0 {
		b.errs = append(b.errs, errors.New("failed to build PV object: missing PV node name"))
		return b
	}
	nodeAffinity := &corev1.VolumeNodeAffinity{
		Required: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      KeyNode,
							Operator: corev1.NodeSelectorOpIn,
							Values: []string{
								nodeName,
							},
						},
					},
				},
			},
		},
	}
	b.pv.object.Spec.NodeAffinity = nodeAffinity
	return b
}

// Build returns the PV API instance
func (b *Builder) Build() (*corev1.PersistentVolume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pv.object, nil
}
