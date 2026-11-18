package domain

import "time"

const (
	// slaCutOffHour adalah jam (WIB) batas penugasan: sebelum jam ini batas SLA hari itu juga.
	slaCutOffHour = 16
	// wibOffsetSeconds adalah selisih WIB dari UTC (UTC+7).
	wibOffsetSeconds = 7 * 60 * 60
)

// WIB adalah zona waktu bisnis (UTC+7). Penyimpanan tetap UTC; batas hari laporan
// dan batas SLA dihitung dalam zona ini.
var WIB = time.FixedZone("WIB", wibOffsetSeconds)

// DayStartWIB mengembalikan 00:00 WIB pada tanggal WIB dari t.
func DayStartWIB(t time.Time) time.Time {
	w := t.In(WIB)
	return time.Date(w.Year(), w.Month(), w.Day(), 0, 0, 0, 0, WIB)
}

// SLADeadline menghitung batas SLA dari waktu penugasan: ditugaskan sebelum 16:00 WIB
// berarti batasnya 23:59:59 WIB hari itu, pukul 16:00 WIB atau lebih berarti 16:00:00 WIB besok.
// Hasilnya dalam WIB.
func SLADeadline(assignedAt time.Time) time.Time {
	w := assignedAt.In(WIB)
	if w.Hour() < slaCutOffHour {
		return time.Date(w.Year(), w.Month(), w.Day(), 23, 59, 59, 0, WIB)
	}
	return time.Date(w.Year(), w.Month(), w.Day()+1, slaCutOffHour, 0, 0, 0, WIB)
}
