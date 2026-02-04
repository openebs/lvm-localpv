package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openebs/lvm-localpv/pkg/version"
)

// identity is the server implementation
// for CSI IdentityServer
type identity struct {
	driver *CSIDriver
	csi.UnimplementedIdentityServer
}

// NewIdentity returns a new instance of CSI
// IdentityServer
func NewIdentity(d *CSIDriver) csi.IdentityServer {
	return &identity{
		driver: d,
	}
}

// GetPluginInfo returns the version and name of
// this service
//
// This implements csi.IdentityServer
func (id *identity) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest,
) (*csi.GetPluginInfoResponse, error) {

	if id.driver.config.DriverName == "" {
		return nil, status.Error(codes.Unavailable, "missing driver name")
	}

	if id.driver.config.Version == "" {
		return nil, status.Error(codes.Unavailable, "missing driver version")
	}

	return &csi.GetPluginInfoResponse{
		Name: id.driver.config.DriverName,
		// TODO
		// verify which version needs to be used:
		// config.version or version.Current()
		VendorVersion: version.Current(),
	}, nil
}

// TODO
// Need to implement this
//
// # Probe checks if the plugin is running or not
//
// This implements csi.IdentityServer
func (id *identity) Probe(
	ctx context.Context,
	req *csi.ProbeRequest,
) (*csi.ProbeResponse, error) {

	return &csi.ProbeResponse{}, nil
}

// GetPluginCapabilities returns supported capabilities
// of this plugin
//
// Currently it reports whether this plugin can serve
// the Controller interface. Controller interface methods
// are called dependant on this
//
// This implements csi.IdentityServer
func (id *identity) GetPluginCapabilities(
	ctx context.Context,
	req *csi.GetPluginCapabilitiesRequest,
) (*csi.GetPluginCapabilitiesResponse, error) {

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
					},
				},
			},
		},
	}, nil
}
