/*
 Copyright © 2021 The OpenEBS Authors

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

package driver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	apis "github.com/openebs/lvm-localpv/pkg/apis/openebs.io/lvm/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openebs/lib-csi/pkg/common/helpers"
	"k8s.io/apimachinery/pkg/api/resource"
)

const qosParamMax = "max"

// VolumeParams holds collection of supported settings that can
// be configured in storage class.
type VolumeParams struct {
	// VgPattern specifies vg regex to use for
	// provisioning logical volumes.
	VgPattern *regexp.Regexp

	Scheduler     string
	Shared        string
	ThinProvision string
	// extra optional metadata passed by external provisioner
	// if enabled. See --extra-create-metadata flag for more details.
	// https://github.com/kubernetes-csi/external-provisioner#recommended-optional-arguments
	PVCName       string
	PVCNamespace  string
	PVName        string
	FormatOptions string
}

// SnapshotParams holds collection of supported settings that can
// be configured in snapshot class.
type SnapshotParams struct {
	SnapSize    float64
	AbsSnapSize bool
}

// qosParamState describes "what to do" with one parameter.
// ParamsPresent=false => the key did not arrive from VAC => do not change anything
// ParamsPresent=true => the key is present => set ParamsValue ("max" or >0)
type qosParamState struct {
	ParamPresent bool
	ParamValue   string
}

// qosParams holds collection of supported settings that can be configured
// in VolumeAttributesClass.
type qosParams struct {
	UnifiedIOPS qosParamState
	ReadIOPS    qosParamState
	WriteIOPS   qosParamState

	UnifiedBPS qosParamState
	ReadBPS    qosParamState
	WriteBPS   qosParamState
}

// qosParamsPresent returns true if VAC request contains at least one supported QoS key.
// Helps determine the state of the VAC.
func (p *qosParams) qosParamsPresent() bool {
	if p == nil {
		return false
	}
	return p.UnifiedIOPS.ParamPresent ||
		p.ReadIOPS.ParamPresent || p.WriteIOPS.ParamPresent ||
		p.UnifiedBPS.ParamPresent ||
		p.ReadBPS.ParamPresent || p.WriteBPS.ParamPresent
}

// NewVolumeParams parses the input params and instantiates new VolumeParams.
func NewVolumeParams(m map[string]string) (*VolumeParams, error) {
	params := &VolumeParams{ // set up defaults, if any.
		Scheduler:     SpaceWeighted,
		Shared:        "no",
		ThinProvision: "no",
	}
	// parameter keys may be mistyped from the CRD specification when declaring
	// the storageclass, which kubectl validation will not catch. Because
	// parameter keys (not values!) are all lowercase, keys may safely be forced
	// to the lower case.
	m = helpers.GetCaseInsensitiveMap(&m)

	// for ensuring backward compatibility, we first check if
	// there is any volgroup param exists for storage class.

	vgPattern := m["vgpattern"]
	volGroup, ok := m["volgroup"]
	if ok {
		vgPattern = fmt.Sprintf("^%v$", volGroup)
	}

	params.FormatOptions = m["formatoptions"]

	var err error
	if params.VgPattern, err = regexp.Compile(vgPattern); err != nil {
		return nil, fmt.Errorf("invalid volgroup/vgpattern param %v: %v", vgPattern, err)
	}

	// parse string params
	stringParams := map[string]*string{
		"scheduler":     &params.Scheduler,
		"shared":        &params.Shared,
		"thinprovision": &params.ThinProvision,
	}
	for key, param := range stringParams {
		value, ok := m[key]
		if !ok {
			continue
		}
		*param = value
	}

	params.PVCName = m["csi.storage.k8s.io/pvc/name"]
	params.PVCNamespace = m["csi.storage.k8s.io/pvc/namespace"]
	params.PVName = m["csi.storage.k8s.io/pv/name"]

	return params, nil
}

// NewSnapshotParams parses the input params and instantiates new SnapshotParams.
func NewSnapshotParams(m map[string]string) (*SnapshotParams, error) {
	var err error
	params := &SnapshotParams{}

	// parameter keys may be mistyped from the CRD specification when declaring
	// the storageclass, which kubectl validation will not catch. Because
	// parameter keys (not values!) are all lowercase, keys may safely be forced
	// to the lower case.
	m = helpers.GetCaseInsensitiveMap(&m)

	size, ok := m["snapsize"]
	if ok {
		if strings.HasSuffix(size, "%") {
			snapSize := size[:len(size)-1]
			params.SnapSize, err = strconv.ParseFloat(snapSize, 64)
			if err != nil {
				return nil, err
			}
			if params.SnapSize < 1 || params.SnapSize > 100 {
				return nil, fmt.Errorf("snapSize percentage should be between 1 and 100, found %s", size)
			}
		} else {
			qty, err := resource.ParseQuantity(size)
			if err != nil {
				return nil, err
			}
			snapSize, _ := qty.AsInt64()
			if snapSize < 1 {
				return nil, fmt.Errorf("absolute snapSize should greater than 0, found %s", size)
			}
			params.AbsSnapSize = true
			params.SnapSize = float64(snapSize)
		}
	}

	return params, nil
}

// NewQoSParams parses and validates VAC mutable parameters into qosParams:
// keys are case-sensitive.
// recognizes only keys for QoS.
// validates and determines allowed parameters.
// rejects conflicting unified vs per-direction values.
func NewQoSParams(m map[string]string) (*qosParams, error) {
	var qosKeys = struct {
		IOPS, IOPSRead, IOPSWrite string
		BPS, BPSRead, BPSWrite    string
	}{
		IOPS:      "qosIopsLimit",
		IOPSRead:  "qosIopsReadLimit",
		IOPSWrite: "qosIopsWriteLimit",
		BPS:       "qosBandwithPerSec",
		BPSRead:   "qosBandwithReadPerSec",
		BPSWrite:  "qosBandwithWritePerSec",
	}

	// VAC didn't provide QoS keys at all.
	if len(m) == 0 {
		return nil, nil
	}

	p := &qosParams{}
	var errs []string

	iopsTargets := map[string]*qosParamState{
		qosKeys.IOPS:      &p.UnifiedIOPS,
		qosKeys.IOPSRead:  &p.ReadIOPS,
		qosKeys.IOPSWrite: &p.WriteIOPS,
	}
	bpsTargets := map[string]*qosParamState{
		qosKeys.BPS:      &p.UnifiedBPS,
		qosKeys.BPSRead:  &p.ReadBPS,
		qosKeys.BPSWrite: &p.WriteBPS,
	}

	for k, raw := range m {
		if raw != strings.TrimSpace(raw) {
			errs = append(errs, fmt.Sprintf("%s: value must not contain leading/trailing whitespace", k))
			continue
		}

		if dst, ok := iopsTargets[k]; ok {
			*dst = qosParamsParseIOPS(&errs, k, raw)
			continue
		}
		if dst, ok := bpsTargets[k]; ok {
			*dst = qosParamsParseBPS(&errs, k, raw)
			continue
		}
		errs = append(errs, fmt.Sprintf("%s: unsupported VAC key", k))
	}

	// check unified parameter and a separate parameter corresponding to it and find conflict
	errs = qosParamsConflict(errs, qosKeys.IOPS, p.UnifiedIOPS, qosKeys.IOPSRead, p.ReadIOPS)
	errs = qosParamsConflict(errs, qosKeys.IOPS, p.UnifiedIOPS, qosKeys.IOPSWrite, p.WriteIOPS)
	errs = qosParamsConflict(errs, qosKeys.BPS, p.UnifiedBPS, qosKeys.BPSRead, p.ReadBPS)
	errs = qosParamsConflict(errs, qosKeys.BPS, p.UnifiedBPS, qosKeys.BPSWrite, p.WriteBPS)

	// apply unified parameter when separate parameter not provided
	if p.UnifiedIOPS.ParamPresent {
		if _, ok := m[qosKeys.IOPSRead]; !ok {
			p.ReadIOPS = p.UnifiedIOPS
		}
		if _, ok := m[qosKeys.IOPSWrite]; !ok {
			p.WriteIOPS = p.UnifiedIOPS
		}
	}
	if p.UnifiedBPS.ParamPresent {
		if _, ok := m[qosKeys.BPSRead]; !ok {
			p.ReadBPS = p.UnifiedBPS
		}
		if _, ok := m[qosKeys.BPSWrite]; !ok {
			p.WriteBPS = p.UnifiedBPS
		}
	}

	if len(errs) > 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid VAC mutableParameters: "+strings.Join(errs, "; "))
	}
	return p, nil
}

// qosParamsConflict detects a conflict between a unified key (e.g. qosIopsLimit)
// and a per-direction key (e.g. qosIopsReadLimit).
// This prevents configurations such as qosIopsLimit="100" and qosIopsReadLimit="200".
func qosParamsConflict(errs []string, uKey string, uParam qosParamState, sKey string, sParam qosParamState) []string {
	if !uParam.ParamPresent || !sParam.ParamPresent {
		return errs
	}

	// In case of an error during parsing, additional checking
	if uParam.ParamValue == "" || sParam.ParamValue == "" {
		errs = append(errs, fmt.Sprintf("invalid value when checking parameters for conflict: %s vs %s", sKey, uKey))
	}
	if uParam.ParamValue != sParam.ParamValue {
		errs = append(errs, fmt.Sprintf("%s: conflicts with %s (%s vs %s)", sKey, uKey, sParam.ParamValue, uParam.ParamValue))
	}
	return errs
}

// qosParamsParseIOPS parses an IOPS parameter value.
// "max" => remove limit
// uint64 string (>0) => ParamValue set
// empty/invalid/zero => validation error is appended.
func qosParamsParseIOPS(errs *[]string, key, raw string) qosParamState {
	ps := qosParamState{ParamPresent: true}

	if raw == "" {
		*errs = append(*errs, fmt.Sprintf("%s: empty value", key))
		return ps
	}
	if raw == qosParamMax {
		ps.ParamValue = qosParamMax
		return ps
	}

	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		*errs = append(*errs, fmt.Sprintf("%s: invalid uint64 %q", key, raw))
		return ps
	}
	if v == 0 {
		*errs = append(*errs, fmt.Sprintf("%s: must be > 0 or 'max'", key))
		return ps
	}
	ps.ParamValue = strconv.FormatUint(v, 10)
	return ps
}

// qosParamsParseBPS parses a throughput parameter value using Kubernetes Quantity.
// "max" => remove limit
// Quantity that converts to a positive int64 => ParamValue set
// empty/invalid/non-integer/<=0 => validation error is appended.
func qosParamsParseBPS(errs *[]string, key, raw string) qosParamState {
	ps := qosParamState{ParamPresent: true}

	if raw == "" {
		*errs = append(*errs, fmt.Sprintf("%s: empty value", key))
		return ps
	}
	if raw == qosParamMax {
		ps.ParamValue = qosParamMax
		return ps
	}

	v, err := resource.ParseQuantity(raw)
	if err != nil {
		*errs = append(*errs, fmt.Sprintf("%s: invalid quantity %q", key, raw))
		return ps
	}
	n, ok := v.AsInt64()
	if !ok || n <= 0 {
		*errs = append(*errs, fmt.Sprintf("%s: must be > 0 or 'max'", key))
		return ps
	}

	ps.ParamValue = strconv.FormatInt(n, 10)
	return ps
}

// QoSParamsCreateVolume builds the initial VolumeQoS spec for volume creation.
// Only parameters that were explicitly provided AND contain a numeric value
// are written. Returns nil if VAC didn't provide any QoS keys.
func QoSParamsCreateVolume(in *qosParams) *apis.VolumeQoS {
	if in == nil || !in.qosParamsPresent() {
		return nil
	}

	out := qosParamsDefaultMax()
	apply := func(ps qosParamState, dst *string) {
		if !ps.ParamPresent {
			return
		}
		*dst = ps.ParamValue // "max" or number
	}

	apply(in.ReadIOPS, &out.ReadIOPS)
	apply(in.WriteIOPS, &out.WriteIOPS)
	apply(in.ReadBPS, &out.ReadBPS)
	apply(in.WriteBPS, &out.WriteBPS)

	return &out
}

// QoSParamsModifyVolume applies mutable parameters to an existing VolumeQoS spec and
// reports whether it changed. No-op if VAC request contains no QoS keys.
func QoSParamsModifyVolume(in *qosParams, old *apis.VolumeQoS) (*apis.VolumeQoS, bool) {
	if in == nil || !in.qosParamsPresent() {
		return old, false
	}

	var out apis.VolumeQoS
	if old != nil {
		out = *old
	} else {
		out = qosParamsDefaultMax()
	}

	before := out

	apply := func(ps qosParamState, dst *string) {
		// CSI spec on ControllerModifyVolumeRequest says: SPs MUST NOT
		// modify volumes based on the absence of keys, only keys
		// that are specified should result in modifications to the volume.
		if !ps.ParamPresent {
			return
		}
		*dst = ps.ParamValue
	}

	apply(in.ReadIOPS, &out.ReadIOPS)
	apply(in.WriteIOPS, &out.WriteIOPS)
	apply(in.ReadBPS, &out.ReadBPS)
	apply(in.WriteBPS, &out.WriteBPS)

	changed := (out != before)
	return &out, changed
}

// qosParamsDefaultMax prepares basic VAC structure to which restrictions will be applied.
func qosParamsDefaultMax() apis.VolumeQoS {
	return apis.VolumeQoS{
		ReadBPS:   qosParamMax,
		WriteBPS:  qosParamMax,
		ReadIOPS:  qosParamMax,
		WriteIOPS: qosParamMax,
	}
}
