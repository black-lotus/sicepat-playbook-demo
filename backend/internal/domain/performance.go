package domain

import "time"

// DailyCount adalah hitungan paket satu kurir pada satu tanggal WIB.
// Date adalah tanggal sipil dan dibandingkan lewat Format(time.DateOnly).
type DailyCount struct {
	Date          time.Time
	TotalPackages int
	Delivered     int
	Failed        int
	OnTime        int
}

// CourierPerformance adalah total satu kurir dalam periode beserta peringkat dan tren hariannya.
type CourierPerformance struct {
	// Rank adalah posisi global (1 sampai total kurir), bukan posisi dalam halaman.
	Rank          int
	CourierID     string
	CourierName   string
	BranchID      string
	TotalPackages int
	Delivered     int
	Failed        int
	OnTime        int
	// LeaderSuccessRate adalah success rate peringkat 1 di seluruh hasil, bukan hanya halaman ini.
	LeaderSuccessRate float64
	Daily             []DailyCount
}

// SuccessRate adalah rasio paket terkirim terhadap total paket (0..1).
func (p CourierPerformance) SuccessRate() float64 {
	return CourierMetric{TotalPackages: p.TotalPackages, Delivered: p.Delivered}.SuccessRate()
}

// SLARate adalah rasio paket terkirim tepat waktu terhadap paket terkirim (0..1).
func (p CourierPerformance) SLARate() float64 {
	return CourierMetric{Delivered: p.Delivered, OnTime: p.OnTime}.SLARate()
}

// SuccessRateGap adalah selisih success rate pemimpin dengan kurir ini, minimal 0.
func (p CourierPerformance) SuccessRateGap() float64 {
	return max(p.LeaderSuccessRate-p.SuccessRate(), 0)
}

// SuccessRateStatus membandingkan success rate mentah kurir dengan target success rate cabang.
// Mengembalikan nil bila t nil.
func (p CourierPerformance) SuccessRateStatus(t *KPITarget) *TargetStatus {
	if t == nil {
		return nil
	}
	s := CompareTarget(p.SuccessRate(), t.SuccessRate)
	return &s
}

// CourierPerformancePage adalah satu halaman ranking performa kurir.
type CourierPerformancePage struct {
	Items []CourierPerformance
	Total int
}

// FillDaily mengembalikan satu entri per tanggal dari start sampai end inklusif, urut menaik.
// Tanggal tanpa data bernilai 0, dan data di luar periode diabaikan. Hasilnya tidak pernah nil.
func FillDaily(counts []DailyCount, start, end time.Time) []DailyCount {
	byDate := make(map[string]DailyCount, len(counts))
	for _, c := range counts {
		key := c.Date.Format(time.DateOnly)
		cur := byDate[key]
		cur.TotalPackages += c.TotalPackages
		cur.Delivered += c.Delivered
		cur.Failed += c.Failed
		cur.OnTime += c.OnTime
		byDate[key] = cur
	}

	days := []DailyCount{}
	first := civilDate(start)
	last := civilDate(end)
	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		c := byDate[d.Format(time.DateOnly)]
		c.Date = d
		days = append(days, c)
	}
	return days
}

// civilDate memotong t menjadi tanggal sipil (00:00 UTC) tanpa menggeser tanggalnya.
func civilDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
