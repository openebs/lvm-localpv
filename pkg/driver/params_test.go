package driver

import (
	"reflect"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewQoSParams_EmptyMapReturnsNil(t *testing.T) {
	p, err := NewQoSParams(map[string]string{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil params, got %#v", p)
	}
}

func TestNewQoSParams_NilMapReturnsNil(t *testing.T) {
	var m map[string]string
	p, err := NewQoSParams(m)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if p != nil {
		t.Fatalf("expected nil params, got %#v", p)
	}
}

func TestNewQoSParams_Cases(t *testing.T) {
	type want struct {
		params *qosParams
		code   codes.Code
		errHas []string
	}

	tests := []struct {
		name string
		in   map[string]string
		want want
	}{
		{
			name: "low: unified iops + unified bps",
			in: map[string]string{
				"qosIopsLimit":      "20",
				"qosBandwithPerSec": "20000",
			},
			want: want{
				code: codes.OK,
				params: &qosParams{
					UnifiedIOPS: qosParamState{ParamPresent: true, ParamValue: "20"},
					ReadIOPS:    qosParamState{ParamPresent: true, ParamValue: "20"}, // applied from unified
					WriteIOPS:   qosParamState{ParamPresent: true, ParamValue: "20"}, // applied from unified

					UnifiedBPS: qosParamState{ParamPresent: true, ParamValue: "20000"},
					ReadBPS:    qosParamState{ParamPresent: true, ParamValue: "20000"}, // applied from unified
					WriteBPS:   qosParamState{ParamPresent: true, ParamValue: "20000"}, // applied from unified
				},
			},
		},
		{
			name: "iops-max: iops=max, bps numeric",
			in: map[string]string{
				"qosIopsLimit":      "max",
				"qosBandwithPerSec": "7000000",
			},
			want: want{
				code: codes.OK,
				params: &qosParams{
					UnifiedIOPS: qosParamState{ParamPresent: true, ParamValue: "max"},
					ReadIOPS:    qosParamState{ParamPresent: true, ParamValue: "max"},
					WriteIOPS:   qosParamState{ParamPresent: true, ParamValue: "max"},

					UnifiedBPS: qosParamState{ParamPresent: true, ParamValue: "7000000"},
					ReadBPS:    qosParamState{ParamPresent: true, ParamValue: "7000000"},
					WriteBPS:   qosParamState{ParamPresent: true, ParamValue: "7000000"},
				},
			},
		},
		{
			name: "mi quantity bps valid",
			in: map[string]string{
				"qosIopsLimit":      "10",
				"qosBandwithPerSec": "7000Mi",
			},
			want: want{
				code: codes.OK,
				params: &qosParams{
					UnifiedIOPS: qosParamState{ParamPresent: true, ParamValue: "10"},
					ReadIOPS:    qosParamState{ParamPresent: true, ParamValue: "10"},
					WriteIOPS:   qosParamState{ParamPresent: true, ParamValue: "10"},

					UnifiedBPS: qosParamState{ParamPresent: true, ParamValue: "7340032000"},
					ReadBPS:    qosParamState{ParamPresent: true, ParamValue: "7340032000"},
					WriteBPS:   qosParamState{ParamPresent: true, ParamValue: "7340032000"},
				},
			},
		},
		{
			name: "directional limits valid",
			in: map[string]string{
				"qosIopsReadLimit":       "100",
				"qosIopsWriteLimit":      "200",
				"qosBandwithReadPerSec":  "7000Mi",
				"qosBandwithWritePerSec": "8000Mi",
			},
			want: want{
				code: codes.OK,
				params: &qosParams{
					UnifiedIOPS: qosParamState{ParamPresent: false, ParamValue: ""},
					ReadIOPS:    qosParamState{ParamPresent: true, ParamValue: "100"},
					WriteIOPS:   qosParamState{ParamPresent: true, ParamValue: "200"},

					UnifiedBPS: qosParamState{ParamPresent: false, ParamValue: ""},
					ReadBPS:    qosParamState{ParamPresent: true, ParamValue: "7340032000"},
					WriteBPS:   qosParamState{ParamPresent: true, ParamValue: "8388608000"},
				},
			},
		},
		{
			name: "unlim: max/max",
			in: map[string]string{
				"qosIopsLimit":      "max",
				"qosBandwithPerSec": "max",
			},
			want: want{
				code: codes.OK,
				params: &qosParams{
					UnifiedIOPS: qosParamState{ParamPresent: true, ParamValue: "max"},
					ReadIOPS:    qosParamState{ParamPresent: true, ParamValue: "max"},
					WriteIOPS:   qosParamState{ParamPresent: true, ParamValue: "max"},

					UnifiedBPS: qosParamState{ParamPresent: true, ParamValue: "max"},
					ReadBPS:    qosParamState{ParamPresent: true, ParamValue: "max"},
					WriteBPS:   qosParamState{ParamPresent: true, ParamValue: "max"},
				},
			},
		},
		{
			name: "zero iops invalid",
			in: map[string]string{
				"qosIopsLimit":      "0",
				"qosBandwithPerSec": "7000000",
			},
			want: want{
				code:   codes.InvalidArgument,
				errHas: []string{"qosIopsLimit: must be > 0 or 'max'"},
			},
		},
		{
			name: "duplicate key different case is unsupported (case-sensitive)",
			in: map[string]string{
				"qosIopsLimit":      "10",
				"qosIopslimit":      "20",
				"qosBandwithPerSec": "7000Mi",
			},
			want: want{
				code:   codes.InvalidArgument,
				errHas: []string{"qosIopslimit: unsupported VAC key"},
			},
		},
		{
			name: "conflict: unified iops=max with read=1000",
			in: map[string]string{
				"qosIopsLimit":     "max",
				"qosIopsReadLimit": "1000",
			},
			want: want{
				code:   codes.InvalidArgument,
				errHas: []string{"qosIopsReadLimit: conflicts with qosIopsLimit (1000 vs max)"},
			},
		},
		{
			name: "wrong key unsupported",
			in: map[string]string{
				"wrong": "1",
			},
			want: want{
				code:   codes.InvalidArgument,
				errHas: []string{"wrong: unsupported VAC key"},
			},
		},
		{
			name: "whitespace in value rejected",
			in: map[string]string{
				"qosIopsLimit":      " 20",
				"qosBandwithPerSec": "20000",
			},
			want: want{
				code:   codes.InvalidArgument,
				errHas: []string{"qosIopsLimit: value must not contain leading/trailing whitespace"},
			},
		},
		{
			name: "empty string in value rejected",
			in: map[string]string{
				"qosBandwithPerSec": "",
			},
			want: want{
				code:   codes.InvalidArgument,
				errHas: []string{"qosBandwithPerSec: empty value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewQoSParams(tt.in)

			if tt.want.code == codes.OK {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if got == nil {
					t.Fatalf("expected non-nil params")
				}
				if !reflect.DeepEqual(got, tt.want.params) {
					t.Fatalf("params mismatch\n got:  %#v\n want: %#v", got, tt.want.params)
				}
				if !got.qosParamsPresent() {
					t.Fatalf("expected qosParamsPresent()=true, got false")
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error with code %v, got nil", tt.want.code)
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected grpc status error, got %T: %v", err, err)
			}
			if st.Code() != tt.want.code {
				t.Fatalf("expected code %v, got %v (msg=%q)", tt.want.code, st.Code(), st.Message())
			}
			msg := st.Message()
			for _, sub := range tt.want.errHas {
				if !strings.Contains(msg, sub) {
					t.Fatalf("expected error message to contain %q, got %q", sub, msg)
				}
			}
		})
	}
}

func TestQosParamsConflict(t *testing.T) {
	tests := []struct {
		name      string
		uKey      string
		u         qosParamState
		sKey      string
		s         qosParamState
		wantAdded []string
	}{
		{
			name:      "no conflict if unified not present",
			uKey:      "qosIopsLimit",
			u:         qosParamState{ParamPresent: false, ParamValue: "10"},
			sKey:      "qosIopsReadLimit",
			s:         qosParamState{ParamPresent: true, ParamValue: "20"},
			wantAdded: nil,
		},
		{
			name:      "no conflict if separate not present",
			uKey:      "qosIopsLimit",
			u:         qosParamState{ParamPresent: true, ParamValue: "10"},
			sKey:      "qosIopsReadLimit",
			s:         qosParamState{ParamPresent: false, ParamValue: "20"},
			wantAdded: nil,
		},
		{
			name:      "no conflict if equal values",
			uKey:      "qosIopsLimit",
			u:         qosParamState{ParamPresent: true, ParamValue: "100"},
			sKey:      "qosIopsReadLimit",
			s:         qosParamState{ParamPresent: true, ParamValue: "100"},
			wantAdded: nil,
		},
		{
			name: "conflict if different values",
			uKey: "qosIopsLimit",
			u:    qosParamState{ParamPresent: true, ParamValue: "max"},
			sKey: "qosIopsReadLimit",
			s:    qosParamState{ParamPresent: true, ParamValue: "1000"},
			wantAdded: []string{
				"qosIopsReadLimit: conflicts with qosIopsLimit (1000 vs max)",
			},
		},
		{
			name: "invalid value check triggers when empty param values",
			uKey: "qosIopsLimit",
			u:    qosParamState{ParamPresent: true, ParamValue: ""},
			sKey: "qosIopsReadLimit",
			s:    qosParamState{ParamPresent: true, ParamValue: "1000"},
			wantAdded: []string{
				"invalid value when checking parameters for conflict",
				"qosIopsReadLimit: conflicts with qosIopsLimit (1000 vs )",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errs []string
			errs = qosParamsConflict(errs, tt.uKey, tt.u, tt.sKey, tt.s)

			if len(tt.wantAdded) == 0 {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %#v", errs)
				}
				return
			}
			joined := strings.Join(errs, "; ")
			for _, sub := range tt.wantAdded {
				if !strings.Contains(joined, sub) {
					t.Fatalf("expected errs to contain %q, got %q", sub, joined)
				}
			}
		})
	}
}
