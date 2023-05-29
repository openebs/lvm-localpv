/*
Copyright © 2019 The OpenEBS Authors

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

package config

// Config struct fills the parameters of request or user input
type Config struct {
	// DriverName to be registered at CSI
	DriverName string

	// PluginType flags if the driver is
	// it is a node plugin or controller
	// plugin
	PluginType string

	// Version of the CSI controller/node driver
	Version string

	// Endpoint on which requests are made by kubelet
	// or external provisioner
	//
	// NOTE:
	//  - Controller/node plugin will listen on this
	//  - This will be a unix based socket
	Endpoint string

	// NodeID helps in differentiating the nodes on
	// which node drivers are running. This is useful
	// in case of topologies and publishing or
	// unpublishing volumes on nodes
	NodeID string

	// SetIOLimits if set to true, directs the driver
	// to set iops, bps limits on a pod using a volume
	// provisioned on its node. For this to work,
	// CSIDriver.Spec.podInfoOnMount must be set to 'true'
	SetIOLimits bool

	// ContainerRuntime informs the driver of the container runtime
	// used on the node, so that the driver can make assumptions
	// about the cgroup path for a pod requesting a volume mount
	ContainerRuntime string

	// RIopsLimitPerGB provides read iops rate limits per volume group type
	// as a string slice, in the form ["vg1-prefix=100", "vg2-prefix=200"]
	RIopsLimitPerGB *[]string

	// WIopsLimitPerGB provides write iops rate limits per volume group type
	// as a string slice, in the form ["vg1-prefix=100", "vg2-prefix=200"]
	WIopsLimitPerGB *[]string

	// RBpsLimitPerGB provides read bps rate limits per volume group type
	// as a string slice, in the form ["vg1-prefix=100", "vg2-prefix=200"]
	RBpsLimitPerGB *[]string

	// WBpsLimitPerGB provides read bps rate limits per volume group type
	// as a string slice, in the form ["vg1-prefix=100", "vg2-prefix=200"]
	WBpsLimitPerGB *[]string

	// The TCP network address where the prometheus metrics endpoint will listen (example: `:9080`).
	// The default is empty string, which means metrics endpoint is disabled.
	ListenAddress string

	// The HTTP path where prometheus metrics will be exposed. Default is `/metrics`.
	MetricsPath string

	// Exclude metrics about the exporter itself (process_*, go_*).
	DisableExporterMetrics bool

	// KubeAPIQPS is the QPS to use while talking with Kubernetes API server.
	KubeAPIQPS int

	// KubeAPIBurst is the burst to allow while talking with Kubernetes API server.
	KubeAPIBurst int
}

// Default returns a new instance of config
// required to initialize a driver instance
func Default() *Config {
	return &Config{}
}
