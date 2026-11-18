package domain

import (
	"math/big"
	"time"
)

// KPITarget adalah target KPI bulanan satu cabang. Month adalah tanggal sipil hari pertama bulan
// (lokasi UTC, seperti DailyCount.Date).
type KPITarget struct {
	BranchID          string
	Month             time.Time
	SuccessRate       float64
	SLARate           float64
	DeliveredPackages int
}

// PeriodTarget adalah target untuk sebuah periode: rate dari bulan end, volume diprorata per hari.
type PeriodTarget struct {
	SuccessRate       float64
	SLARate           float64
	DeliveredPackages int
}

// TargetStatus adalah hasil perbandingan nilai aktual dengan target.
type TargetStatus string

const (
	// TargetMet berarti nilai aktual sama atau di atas target.
	TargetMet TargetStatus = "met"
	// TargetBelow berarti nilai aktual di bawah target.
	TargetBelow TargetStatus = "below"
)

// MonthOf mengembalikan hari pertama bulan dari tanggal sipil WIB t, sebagai tanggal sipil UTC.
func MonthOf(t time.Time) time.Time {
	w := t.In(WIB)
	return time.Date(w.Year(), w.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// TargetForPeriod menghitung target periode start sampai end (inklusif, tanggal WIB) dari target satu
// cabang. Rate memakai target bulan end; nil bila bulan end tanpa target. Volume adalah jumlah dari
// setiap hari: target bulan hari itu dibagi jumlah hari bulan itu (bulan tanpa target menyumbang 0),
// dibulatkan sekali di akhir dengan math.Round.
func TargetForPeriod(targets []KPITarget, start, end time.Time) *PeriodTarget {
	byMonth := make(map[string]KPITarget, len(targets))
	for _, t := range targets {
		byMonth[t.Month.Format(time.DateOnly)] = t
	}

	endTarget, ok := byMonth[MonthOf(end).Format(time.DateOnly)]
	if !ok {
		return nil
	}

	// Jumlah pecahan dihitung eksak (big.Rat) supaya nilai tepat .5 tidak meleset karena float.
	volume := new(big.Rat)
	first := wibCivilDate(start)
	lastDay := wibCivilDate(end)
	for d := first; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		month := MonthOf(d)
		t, found := byMonth[month.Format(time.DateOnly)]
		if !found {
			continue
		}
		daysInMonth := int64(month.AddDate(0, 1, 0).Sub(month) / (24 * time.Hour))
		volume.Add(volume, big.NewRat(int64(t.DeliveredPackages), daysInMonth))
	}

	return &PeriodTarget{
		SuccessRate:       endTarget.SuccessRate,
		SLARate:           endTarget.SLARate,
		DeliveredPackages: roundHalfAwayFromZero(volume),
	}
}

// roundHalfAwayFromZero membulatkan r ke bilangan bulat terdekat, setengah menjauhi nol (sama dengan
// math.Round). Volume target tidak pernah negatif, tetapi tandanya tetap ditangani.
func roundHalfAwayFromZero(r *big.Rat) int {
	half := big.NewRat(1, 2)
	if r.Sign() < 0 {
		half.Neg(half)
	}
	shifted := new(big.Rat).Add(r, half)
	// Quo pada big.Int memotong ke arah nol, sehingga hasilnya menjauhi nol untuk setengah.
	q := new(big.Int).Quo(shifted.Num(), shifted.Denom())
	return int(q.Int64())
}

// CompareTarget membandingkan nilai aktual mentah dengan target: met bila actual >= target.
func CompareTarget(actual, target float64) TargetStatus {
	if actual >= target {
		return TargetMet
	}
	return TargetBelow
}

// BranchTargetResult adalah status capaian satu cabang terhadap target periodenya.
type BranchTargetResult struct {
	Target            PeriodTarget
	SuccessRateStatus TargetStatus
	SLARateStatus     TargetStatus
	DeliveredStatus   TargetStatus
}

// EvaluateTarget membandingkan capaian cabang dengan target periode, memakai nilai mentah.
// Mengembalikan nil bila t nil. Volume membandingkan Delivered dengan t.DeliveredPackages.
func EvaluateTarget(p BranchPerformance, t *PeriodTarget) *BranchTargetResult {
	if t == nil {
		return nil
	}
	return &BranchTargetResult{
		Target:            *t,
		SuccessRateStatus: CompareTarget(p.SuccessRate(), t.SuccessRate),
		SLARateStatus:     CompareTarget(p.SLARate(), t.SLARate),
		DeliveredStatus:   CompareTarget(float64(p.Delivered), float64(t.DeliveredPackages)),
	}
}

// BranchPerformanceWithTarget adalah performa cabang beserta evaluasi targetnya (nil bila tanpa target).
type BranchPerformanceWithTarget struct {
	BranchPerformance
	Target *BranchTargetResult
}

// BranchPerformanceResultPage adalah satu halaman ranking cabang dengan target.
type BranchPerformanceResultPage struct {
	Items []BranchPerformanceWithTarget
	Total int
}

// wibCivilDate mengembalikan tanggal sipil WIB dari t sebagai 00:00 UTC.
func wibCivilDate(t time.Time) time.Time {
	w := t.In(WIB)
	return time.Date(w.Year(), w.Month(), w.Day(), 0, 0, 0, 0, time.UTC)
}
