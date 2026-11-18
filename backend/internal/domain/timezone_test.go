package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func TestSLADeadline(t *testing.T) {
	wib := func(month time.Month, day, hour, minute, sec int) time.Time {
		return time.Date(2026, month, day, hour, minute, sec, 0, domain.WIB)
	}
	tests := []struct {
		name       string
		assignedAt time.Time
		want       time.Time
	}{
		{name: "pagi hari yang sama", assignedAt: wib(11, 7, 9, 0, 0), want: wib(11, 7, 23, 59, 59)},
		{name: "satu detik sebelum cut-off", assignedAt: wib(11, 7, 15, 59, 59), want: wib(11, 7, 23, 59, 59)},
		{name: "tepat cut-off", assignedAt: wib(11, 7, 16, 0, 0), want: wib(11, 8, 16, 0, 0)},
		{name: "malam hari", assignedAt: wib(11, 7, 21, 0, 0), want: wib(11, 8, 16, 0, 0)},
		{name: "pergantian bulan", assignedAt: wib(11, 30, 17, 0, 0), want: wib(12, 1, 16, 0, 0)},
		{name: "input UTC dihitung dalam WIB", assignedAt: time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC), want: wib(11, 8, 16, 0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.SLADeadline(tt.assignedAt)
			assert.True(t, tt.want.Equal(got), "want %s, got %s", tt.want, got)
			assert.Equal(t, domain.WIB, got.Location())
		})
	}
}

func TestDayStartWIB(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{name: "16:59 UTC masih hari WIB yang sama", in: time.Date(2026, 11, 17, 16, 59, 0, 0, time.UTC), want: time.Date(2026, 11, 17, 0, 0, 0, 0, domain.WIB)},
		{name: "17:00 UTC sudah hari WIB berikutnya", in: time.Date(2026, 11, 17, 17, 0, 0, 0, time.UTC), want: time.Date(2026, 11, 18, 0, 0, 0, 0, domain.WIB)},
		{name: "tengah malam UTC", in: time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), want: time.Date(2026, 11, 17, 0, 0, 0, 0, domain.WIB)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.DayStartWIB(tt.in)
			assert.True(t, tt.want.Equal(got), "want %s, got %s", tt.want, got)
			assert.Equal(t, domain.WIB, got.Location())
		})
	}
}
