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

package iolimit

type Request struct {
	DeviceName       string
	PodUid           string
	ContainerRuntime string
	IOLimit          *IOMax
}

type ValidRequest struct {
	FilePath     string
	DeviceNumber *DeviceNumber
	IOMax        *IOMax
}

type IOMax struct {
	Riops uint64
	Wiops uint64
	Rbps  uint64
	Wbps  uint64
}

type DeviceNumber struct {
	Major uint64
	Minor uint64
}
