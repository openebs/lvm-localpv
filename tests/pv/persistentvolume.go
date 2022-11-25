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

package pv

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// PV is a wrapper over persistentvolume api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type PV struct {
	object *corev1.PersistentVolume
}

// List is a wrapper over persistentvolume api
// object. It provides build, validations and other common
// logic to be used by various feature specific callers.
type List struct {
	items []*PV
}

// Len returns the number of items present
// in the PVList
func (p *List) Len() int {
	return len(p.items)
}

// ToAPIList converts PVList to API PVList
func (p *List) ToAPIList() *corev1.PersistentVolumeList {
	plist := &corev1.PersistentVolumeList{}
	for _, pvc := range p.items {
		plist.Items = append(plist.Items, *pvc.object)
	}
	return plist
}

type pvBuildOption func(*PV)

// NewForAPIObject returns a new instance of PV
func NewForAPIObject(obj *corev1.PersistentVolume, opts ...pvBuildOption) *PV {
	p := &PV{object: obj}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Predicate defines an abstraction
// to determine conditional checks
// against the provided pvc instance
type Predicate func(*PV) bool

// IsNil returns true if the PV instance
// is nil
func (p *PV) IsNil() bool {
	return p.object == nil
}

// GetPath returns path configured on VolumeSource
// The VolumeSource can be either Local or HostPath
func (p *PV) GetPath() string {
	local := p.object.Spec.PersistentVolumeSource.Local
	if local != nil {
		return local.Path
	}
	//Handle the case of Local PV created in 0.9 using
	//HostPath VolumeSource
	hostPath := p.object.Spec.PersistentVolumeSource.HostPath
	if hostPath != nil {
		return hostPath.Path
	}
	return ""
}

// GetAffinitedNodeHostname returns hostname configured using the NodeAffinity
// This method expects only a single hostname to be set.
//
// The PV object will have the node's hostname specified as follows:
//
//	nodeAffinity:
//	  required:
//	    nodeSelectorTerms:
//	    - matchExpressions:
//	      - key: kubernetes.io/hostname
//	        operator: In
//	        values:
//	        - hostname
func (p *PV) GetAffinitedNodeHostname() string {
	nodeAffinity := p.object.Spec.NodeAffinity
	if nodeAffinity == nil {
		return ""
	}
	required := nodeAffinity.Required
	if required == nil {
		return ""
	}

	hostname := ""
	for _, selectorTerm := range required.NodeSelectorTerms {
		for _, expression := range selectorTerm.MatchExpressions {
			if expression.Key == KeyNode &&
				expression.Operator == corev1.NodeSelectorOpIn {
				if len(expression.Values) != 1 {
					return ""
				}
				hostname = expression.Values[0]
				break
			}
		}
		if hostname != "" {
			break
		}
	}
	return hostname
}

// IsNil is predicate to filter out nil PV
// instances
func IsNil() Predicate {
	return func(p *PV) bool {
		return p.IsNil()
	}
}

// ContainsName is filter function to filter pv's
// based on the name
func ContainsName(name string) Predicate {
	return func(p *PV) bool {
		return strings.Contains(p.object.GetName(), name)
	}
}

// PredicateList holds a list of predicate
type PredicateList []Predicate

// all returns true if all the predicates
// succeed against the provided pv
// instance
func (l PredicateList) all(p *PV) bool {
	for _, pred := range l {
		if !pred(p) {
			return false
		}
	}
	return true
}

// WithFilter adds filters on which the pv's has to be filtered
func (b *ListBuilder) WithFilter(pred ...Predicate) *ListBuilder {
	b.filters = append(b.filters, pred...)
	return b
}
