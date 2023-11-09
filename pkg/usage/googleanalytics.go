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
	"k8s.io/klog/v2"
)

// Send sends a single usage metric to Google Analytics with some
// compulsory fields defined in Google Analytics API
// bindings(jpillora/go-ogle-analytics)
func (u *Usage) Send() {
	// Instantiate a Gclient with the tracking ID
	go func() {
		client := u.AnalyticsClient
		event := u.OpenebsEventBuilder.Build()

		if err := client.Send(event); err != nil {
			klog.Errorf(err.Error())
			return
		}
	}()
}
