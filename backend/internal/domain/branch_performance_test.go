package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func TestBranchPerformance_Rates(t *testing.T) {
	tests := []struct {
		name        string
		p           domain.BranchPerformance
		wantSuccess float64
		wantSLARate float64
	}{
		{"normal", domain.BranchPerformance{TotalPackages: 1000, Delivered: 833, OnTime: 700}, 0.833, 700.0 / 833.0},
		{"tanpa paket", domain.BranchPerformance{}, 0, 0},
		{"tanpa terkirim", domain.BranchPerformance{TotalPackages: 5, Failed: 5}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.wantSuccess, tt.p.SuccessRate(), 1e-12)
			assert.InDelta(t, tt.wantSLARate, tt.p.SLARate(), 1e-12)
		})
	}
}
