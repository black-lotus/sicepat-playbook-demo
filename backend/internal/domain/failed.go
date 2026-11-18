package domain

import "time"

// NoReasonLabel adalah label alasan untuk event gagal tanpa failed_reason (NULL atau kosong).
const NoReasonLabel = "Tanpa alasan"

// FailedFilter adalah filter ringkasan gagal kirim. Berbeda dari MetricFilter, tidak ada pagination.
type FailedFilter struct {
	StartDate time.Time
	EndDate   time.Time
	BranchID  string
}

// Validate memastikan periode wajar (wajib, urut, maksimal MaxPeriodDays).
func (f FailedFilter) Validate() error {
	return validatePeriod(f.StartDate, f.EndDate)
}

// PreviousPeriod mengembalikan rentang dengan jumlah hari sama, tepat sebelum start.
// Hanya tanggal sipil yang dipakai, sehingga hasilnya benar lintas bulan dan tahun.
func PreviousPeriod(start, end time.Time) (time.Time, time.Time) {
	loc := start.Location()
	s := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	e := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)
	days := int(e.Sub(s).Round(24*time.Hour)/(24*time.Hour)) + 1
	return s.AddDate(0, 0, -days), s.AddDate(0, 0, -1)
}

// FailedReason adalah jumlah event gagal untuk satu alasan.
type FailedReason struct {
	Reason string
	Count  int
}

// FailedCourier adalah satu kurir dalam daftar gagal terbanyak.
type FailedCourier struct {
	Rank          int
	CourierID     string
	CourierName   string
	BranchID      string
	TotalPackages int
	Failed        int
}

// FailedRate adalah rasio gagal terhadap total paket kurir (0..1), 0 bila total 0.
func (c FailedCourier) FailedRate() float64 {
	if c.TotalPackages == 0 {
		return 0
	}
	return float64(c.Failed) / float64(c.TotalPackages)
}

// FailedSummary adalah ringkasan gagal kirim untuk satu periode. PreviousTotalFailed diisi service.
type FailedSummary struct {
	TotalPackages       int
	TotalFailed         int
	PreviousTotalFailed int
	Reasons             []FailedReason
	TopCouriers         []FailedCourier
}

// FailedRate adalah rasio gagal terhadap total paket (0..1), 0 bila total 0.
func (s FailedSummary) FailedRate() float64 {
	if s.TotalPackages == 0 {
		return 0
	}
	return float64(s.TotalFailed) / float64(s.TotalPackages)
}

// FailedChangeRate adalah perubahan relatif total gagal terhadap periode pembanding; nil bila periode itu 0.
func (s FailedSummary) FailedChangeRate() *float64 {
	if s.PreviousTotalFailed == 0 {
		return nil
	}
	r := float64(s.TotalFailed-s.PreviousTotalFailed) / float64(s.PreviousTotalFailed)
	return &r
}

// Share adalah porsi satu alasan terhadap total gagal (0..1), 0 bila total gagal 0.
func (s FailedSummary) Share(r FailedReason) float64 {
	if s.TotalFailed == 0 {
		return 0
	}
	return float64(r.Count) / float64(s.TotalFailed)
}
