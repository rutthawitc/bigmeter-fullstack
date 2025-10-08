package alert

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var thaiMonths = []string{
	"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
}

// FormatAlertMessage formats alert statistics into a Thai language message
func FormatAlertMessage(stats *AlertStats, link string) string {
	// Format current date in Thai
	now := stats.GeneratedAt
	thaiYear := now.Year() + 543
	thaiMonth := thaiMonths[now.Month()-1]
	dateStr := fmt.Sprintf("%02d %s %d", now.Day(), thaiMonth, thaiYear)

	var builder strings.Builder

	// Header
	builder.WriteString("🔔 แจ้งเตือน\n")
	builder.WriteString(fmt.Sprintf("📅 ประจำวันที่ %s\n", dateStr))
	builder.WriteString(fmt.Sprintf("📊 สรุปข้อมูลการใช้น้ำของผู้ใช้น้ำรายใหญ่ที่มีผลต่างการใช้น้ำลดลง %.0f%% ขึ้นไป ดังนี้\n", stats.Threshold))
	builder.WriteString("\n---\n\n")

	// Branch list
	if len(stats.BranchAlerts) == 0 {
		builder.WriteString("ไม่พบรายการที่เข้าเงื่อนไข\n")
	} else {
		for _, branchAlert := range stats.BranchAlerts {
			branchName := branchAlert.BranchName
			if branchName == "" {
				branchName = branchAlert.BranchCode
			}
			builder.WriteString(fmt.Sprintf("- %s %d ราย\n", branchName, branchAlert.Count))
		}
	}

	builder.WriteString("\n---\n\n")

	// Footer
	if link != "" {
		builder.WriteString(fmt.Sprintf("💡 เข้าตรวจสอบข้อมูลเพิ่มเติมได้ที่ %s\n", link))
	}
	builder.WriteString("⏳ ขอให้เร่งรัดดำเนินการตรวจสอบด้วยครับ\n")

	return builder.String()
}

// FormatThaiMonth formats YYYYMM to Thai month name
func FormatThaiMonth(ym string) string {
	if len(ym) != 6 {
		return ym
	}

	year, err := strconv.Atoi(ym[:4])
	if err != nil {
		return ym
	}

	month, err := strconv.Atoi(ym[4:])
	if err != nil || month < 1 || month > 12 {
		return ym
	}

	thaiYear := year + 543
	return fmt.Sprintf("%s %d", thaiMonths[month-1], thaiYear)
}

// FormatThaiDate formats a time.Time to Thai date format
func FormatThaiDate(t time.Time) string {
	thaiYear := t.Year() + 543
	thaiMonth := thaiMonths[t.Month()-1]
	return fmt.Sprintf("%02d %s %d", t.Day(), thaiMonth, thaiYear)
}
