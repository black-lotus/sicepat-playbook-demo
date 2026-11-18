package domain

import (
	"fmt"
	"time"
)

const (
	// DefaultPageSize dipakai saat client tidak mengirim page_size.
	DefaultPageSize = 20
	// MaxPageSize membatasi jumlah baris per halaman.
	MaxPageSize = 100
	// MaxPeriodDays membatasi rentang periode satu query.
	MaxPeriodDays = 31
)

// CourierMetric adalah agregat KPI harian satu kurir.
type CourierMetric struct {
	CourierID     string
	CourierName   string
	BranchID      string
	Date          time.Time
	TotalPackages int
	Delivered     int
	Failed        int
	OnTime        int
}

// SuccessRate adalah rasio paket terkirim terhadap total paket (0..1).
func (m CourierMetric) SuccessRate() float64 {
	if m.TotalPackages == 0 {
		return 0
	}
	return float64(m.Delivered) / float64(m.TotalPackages)
}

// SLARate adalah rasio paket terkirim tepat waktu terhadap paket terkirim (0..1).
func (m CourierMetric) SLARate() float64 {
	if m.Delivered == 0 {
		return 0
	}
	return float64(m.OnTime) / float64(m.Delivered)
}

// CourierMetricPage adalah satu halaman hasil query metrik kurir.
type CourierMetricPage struct {
	Items []CourierMetric
	Total int
}

// MetricFilter adalah filter query metrik kurir.
type MetricFilter struct {
	StartDate time.Time
	EndDate   time.Time
	BranchID  string
	Page      int
	PageSize  int
}

// WithDefaults mengisi nilai default untuk page dan page size.
func (f MetricFilter) WithDefaults() MetricFilter {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = DefaultPageSize
	}
	return f
}

// Validate memastikan filter punya periode yang wajar dan pagination yang terbatas.
func (f MetricFilter) Validate() error {
	switch {
	case f.StartDate.IsZero() || f.EndDate.IsZero():
		return fmt.Errorf("%w: start_date dan end_date wajib diisi", ErrInvalidFilter)
	case f.EndDate.Before(f.StartDate):
		return fmt.Errorf("%w: end_date tidak boleh sebelum start_date", ErrInvalidFilter)
	case f.EndDate.Sub(f.StartDate) > (MaxPeriodDays-1)*24*time.Hour:
		return fmt.Errorf("%w: periode maksimal %d hari", ErrInvalidFilter, MaxPeriodDays)
	case f.Page < 1:
		return fmt.Errorf("%w: page minimal 1", ErrInvalidFilter)
	case f.PageSize < 1 || f.PageSize > MaxPageSize:
		return fmt.Errorf("%w: page_size harus 1..%d", ErrInvalidFilter, MaxPageSize)
	}
	return nil
}

// Offset adalah jumlah baris yang dilewati untuk halaman ini.
func (f MetricFilter) Offset() int {
	return (f.Page - 1) * f.PageSize
}

// CacheKey adalah key cache yang unik untuk kombinasi filter ini.
func (f MetricFilter) CacheKey() string {
	return fmt.Sprintf("courier-metrics:%s:%s:%s:%d:%d",
		f.StartDate.Format(time.DateOnly), f.EndDate.Format(time.DateOnly), f.BranchID, f.Page, f.PageSize)
}
