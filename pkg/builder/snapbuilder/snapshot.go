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

import apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"

type LVMSnapshot struct {
	Object *apis.LVMSnapshot
}

type LVMSnapshotList struct {
	List apis.LVMSnapshotList
}

func From(snap *apis.LVMSnapshot) *LVMSnapshot {
	return &LVMSnapshot{
		Object: snap,
	}
}

func (snapList *LVMSnapshotList) Len() int {
	return len(snapList.List.Items)
}
