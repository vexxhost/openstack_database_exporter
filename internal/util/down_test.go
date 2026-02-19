package util

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestDownCollector(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		subsystem string
		wantName  string
	}{
		{
			name:      "cinder",
			namespace: "openstack",
			subsystem: "cinder",
			wantName:  "openstack_cinder_up",
		},
		{
			name:      "container_infra",
			namespace: "openstack",
			subsystem: "container_infra",
			wantName:  "openstack_container_infra_up",
		},
		{
			name:      "loadbalancer",
			namespace: "openstack",
			subsystem: "loadbalancer",
			wantName:  "openstack_loadbalancer_up",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			reg.MustRegister(NewDownCollector(tt.namespace, tt.subsystem))

			mfs, err := reg.Gather()
			if err != nil {
				t.Fatalf("failed to gather metrics: %v", err)
			}

			if len(mfs) != 1 {
				t.Fatalf("expected 1 metric family, got %d", len(mfs))
			}

			mf := mfs[0]
			if mf.GetName() != tt.wantName {
				t.Errorf("expected metric name %q, got %q", tt.wantName, mf.GetName())
			}

			metrics := mf.GetMetric()
			if len(metrics) != 1 {
				t.Fatalf("expected 1 metric, got %d", len(metrics))
			}

			if metrics[0].GetGauge().GetValue() != 0 {
				t.Errorf("expected value 0, got %f", metrics[0].GetGauge().GetValue())
			}

			// Verify it renders properly in Prometheus exposition format
			var buf strings.Builder
			enc := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
			if err := enc.Encode(mf); err != nil {
				t.Fatalf("failed to encode metric: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantName+" 0") {
				t.Errorf("exposition output missing %q, got: %s", tt.wantName+" 0", output)
			}
		})
	}
}
