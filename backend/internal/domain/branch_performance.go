package domain

// BranchPerformance adalah total satu cabang dalam periode beserta peringkat dan tren hariannya.
type BranchPerformance struct {
	// Rank adalah posisi global (1 sampai total cabang), bukan posisi dalam halaman.
	Rank          int
	BranchID      string
	CourierCount  int
	TotalPackages int
	Delivered     int
	Failed        int
	OnTime        int
	Daily         []DailyCount
}

// SuccessRate adalah rasio paket terkirim terhadap total paket (0..1).
func (p BranchPerformance) SuccessRate() float64 {
	return CourierMetric{TotalPackages: p.TotalPackages, Delivered: p.Delivered}.SuccessRate()
}

// SLARate adalah rasio paket terkirim tepat waktu terhadap paket terkirim (0..1).
func (p BranchPerformance) SLARate() float64 {
	return CourierMetric{Delivered: p.Delivered, OnTime: p.OnTime}.SLARate()
}

// BranchPerformancePage adalah satu halaman ranking performa cabang.
type BranchPerformancePage struct {
	Items []BranchPerformance
	Total int
}
