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
)

// ListBuilder enables building an instance of
// PVlist
type ListBuilder struct {
	list    *PVList
	filters PredicateList
	errs    []error
}

// NewListBuilder returns an instance of ListBuilder
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: &PVList{}}
}

// ListBuilderForAPIObjects builds the ListBuilder object based on PV api list
func ListBuilderForAPIObjects(pvs *corev1.PersistentVolumeList) *ListBuilder {
	b := &ListBuilder{list: &PVList{}}
	if pvs == nil {
		b.errs = append(b.errs, errors.New("failed to build pv list: missing api list"))
		return b
	}
	for _, pv := range pvs.Items {
		pv := pv
		b.list.items = append(b.list.items, &PV{object: &pv})
	}
	return b
}

// ListBuilderForObjects builds the ListBuilder object based on PVList
func ListBuilderForObjects(pvs *PVList) *ListBuilder {
	b := &ListBuilder{}
	if pvs == nil {
		b.errs = append(b.errs, errors.New("failed to build pv list: missing object list"))
		return b
	}
	b.list = pvs
	return b
}

// List returns the list of pv
// instances that was built by this
// builder
func (b *ListBuilder) List() (*PVList, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("failed to list pv: %+v", b.errs)
	}
	if b.filters == nil || len(b.filters) == 0 {
		return b.list, nil
	}
	filteredList := &PVList{}
	for _, pv := range b.list.items {
		if b.filters.all(pv) {
			filteredList.items = append(filteredList.items, pv)
		}
	}
	return filteredList, nil
}

// Len returns the number of items present
// in the PVCList of a builder
func (b *ListBuilder) Len() (int, error) {
	l, err := b.List()
	if err != nil {
		return 0, err
	}
	return l.Len(), nil
}

// APIList builds core API PV list using listbuilder
func (b *ListBuilder) APIList() (*corev1.PersistentVolumeList, error) {
	l, err := b.List()
	if err != nil {
		return nil, err
	}
	return l.ToAPIList(), nil
}
