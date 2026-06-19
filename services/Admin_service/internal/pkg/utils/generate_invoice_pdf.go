package utils

import (
	"admin_service/internal/domain/entities"
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func GeneratePDF(data entities.InvoiceData) ([]byte,error) {
	const (
		pageW  = 210.0 // A4 mm
		pageH  = 297.0
		margin = 18.0

		// Brand colours (RGB)
		inkR, inkG, inkB         = 15, 15, 15   // near-black body
		accentR, accentG, accentB = 30, 30, 30  // slightly lighter
		mutedR, mutedG, mutedB   = 110, 110, 110 // muted grey labels
		lineR, lineG, lineB      = 220, 220, 215 // hairline rules
		greenR, greenG, greenB   = 72, 187, 120  // success green badge
		bgR, bgG, bgB            = 250, 250, 248 // very light tint band
	)

	f := gofpdf.New("P", "mm", "A4", "")
	f.SetMargins(margin, margin, margin)
	f.SetAutoPageBreak(false, 0)
	f.AddPage()

	// ── Helpers ──────────────────────────────────────────────────────────────

	setFont := func(style string, size float64) {
		f.SetFont("Helvetica", style, size)
	}
	setColor := func(r, g, b int) {
		f.SetTextColor(r, g, b)
		f.SetDrawColor(r, g, b)
	}
	setFill := func(r, g, b int) {
		f.SetFillColor(r, g, b)
	}
	money := func(v float64) string {
		// Indian locale formatting: ₹ 1,23,456.00
		// gofpdf uses Latin-1; we use "Rs." to stay safe with built-in fonts.
		return fmt.Sprintf("Rs. %s", formatINR(v))
	}
	fmtDate := func(t time.Time) string {
		return t.Format("02 Jan 2006")
	}
	drawHRule := func(y float64) {
		setColor(lineR, lineG, lineB)
		f.SetLineWidth(0.25)
		f.Line(margin, y, pageW-margin, y)
	}
	label := func(x, y float64, txt string) {
		setFont("", 7)
		setColor(mutedR, mutedG, mutedB)
		f.Text(x, y, strings.ToUpper(txt))
	}
	value := func(x, y float64, txt string) {
		setFont("", 10)
		setColor(inkR, inkG, inkB)
		f.Text(x, y, txt)
	}

	// ── Top accent bar (3 mm) ─────────────────────────────────────────────────
	setFill(inkR, inkG, inkB)
	f.Rect(0, 0, pageW, 3, "F")

	// ── Brand / header ───────────────────────────────────────────────────────
	y := 18.0

	// Logo text — "WATERFALL"
	setFont("B", 22)
	setColor(inkR, inkG, inkB)
	f.Text(margin, y, "WATERFALL")

	// Invoice label + number (right-aligned)
	setFont("", 9)
	setColor(mutedR, mutedG, mutedB)
	invLabel := "INVOICE"
	f.Text(pageW-margin-f.GetStringWidth(invLabel), y-6, invLabel)

	setFont("B", 13)
	setColor(inkR, inkG, inkB)
	invNum := data.InvoiceNumber
	f.Text(pageW-margin-f.GetStringWidth(invNum), y, invNum)

	// Tagline
	y += 5
	setFont("", 8)
	setColor(mutedR, mutedG, mutedB)
	f.Text(margin, y, "Job Scheduling Platform")

	// ── Divider ──────────────────────────────────────────────────────────────
	y += 7
	drawHRule(y)
	y += 8

	// ── From / To / Date meta block ───────────────────────────────────────────
	col2 := margin + 68.0
	col3 := pageW/2 + 8

	label(margin, y, "From")
	label(col2, y, "Bill To")
	label(col3, y, "Invoice Date")
	label(col3+50, y, "Due / Next Payment")

	y += 5
	value(margin, y, "Waterfall Technologies")
	value(col2, y, data.UserName)
	value(col3, y, fmtDate(data.CreatedDate))
	value(col3+50, y, fmtDate(data.NextPayment))

	y += 5
	setFont("", 8.5)
	setColor(mutedR, mutedG, mutedB)
	f.Text(margin, y, "waterfall.helper@gmail.com")
	f.Text(col2, y, data.UserEmail)

	// ── Divider ──────────────────────────────────────────────────────────────
	y += 12
	drawHRule(y)

	// ── Line-item table header ────────────────────────────────────────────────
	y += 1
	tableTop := y

	// Tinted header band
	setFill(bgR, bgG, bgB)
	f.Rect(margin, tableTop, pageW-2*margin, 9, "F")

	y += 6.5
	setFont("B", 7.5)
	setColor(mutedR, mutedG, mutedB)
	colDesc := margin + 3.0
	colQty := pageW - margin - 62.0
	colUnit := pageW - margin - 36.0
	colAmt := pageW - margin - 3.0

	f.Text(colDesc, y, "DESCRIPTION")
	f.Text(colQty, y, "QTY")

	// Right-align column headers
	unitLabel := "UNIT PRICE"
	f.Text(colUnit-f.GetStringWidth(unitLabel), y, unitLabel)
	amtLabel := "AMOUNT"
	f.Text(colAmt-f.GetStringWidth(amtLabel), y, amtLabel)

	y += 3
	drawHRule(y)

	// ── Line items ────────────────────────────────────────────────────────────
	y += 9
	setFont("", 10)
	setColor(inkR, inkG, inkB)
	f.Text(colDesc, y, fmt.Sprintf("%s Plan - Monthly Subscription", data.PlanName))
	f.Text(colQty, y, "1")
	unitStr := money(data.PlanAmount)
	f.Text(colUnit-f.GetStringWidth(unitStr), y, unitStr)
	amtStr := money(data.PlanAmount)
	f.Text(colAmt-f.GetStringWidth(amtStr), y, amtStr)

	y += 5
	setFont("", 8.5)
	setColor(mutedR, mutedG, mutedB)
	periodStr := fmt.Sprintf("Billing period: %s - %s", fmtDate(data.CreatedDate), fmtDate(data.NextPayment))
	f.Text(colDesc, y, periodStr)


	// ── Subtotal / Total block ────────────────────────────────────────────────
	y += 10
	drawHRule(y - 2)

	// Total band
	y += 1
	setFill(inkR, inkG, inkB)
	f.Rect(pageW/2, y, pageW/2-margin, 12, "F")

	y += 8
	setFont("B", 10)
	setColor(250, 250, 248)
	totalLabel := "TOTAL PAID"
	f.Text(pageW/2+4, y, totalLabel)
	totalStr := money(data.TotalPaid)
	f.Text(colAmt-f.GetStringWidth(totalStr), y, totalStr)

	// ── Payment status badge ──────────────────────────────────────────────────
	y += 16
	badgeW := 26.0
	badgeH := 8.0
	badgeX := margin

	setFill(greenR, greenG, greenB)
	f.RoundedRect(badgeX, y-badgeH+1, badgeW, badgeH, 1.5, "1234", "F")
	setFont("B", 7.5)
	setColor(255, 255, 255)
	paidLabel := "PAID"
	f.Text(badgeX+(badgeW-f.GetStringWidth(paidLabel))/2, y-1.5, paidLabel)

	// ── Notes / Footer ────────────────────────────────────────────────────────
	y += 14
	drawHRule(y)
	y += 7

	setFont("", 8)
	setColor(mutedR, mutedG, mutedB)
	notes := []string{
		"Thank you for using Waterfall. This invoice is auto-generated and valid without a signature.",
		"For billing queries, contact us at billing@waterfall.dev or visit https://waterfall.varunjp.in/support.",
		"Waterfall Technologies  |  GSTIN: XXAAAAA0000A1Z5  |  CIN: U00000XX0000PTC000000",
	}
	for _, line := range notes {
		f.Text(margin, y, line)
		y += 5
	}

	// Page number
	setFont("", 7.5)
	setColor(mutedR, mutedG, mutedB)
	pg := "Page 1 of 1"
	f.Text(pageW-margin-f.GetStringWidth(pg), pageH-10, pg)

	// Bottom accent bar
	setFill(inkR, inkG, inkB)
	f.Rect(0, pageH-3, pageW, 3, "F")

	// ── Export ───────────────────────────────────────────────────────────────
	var buf bytes.Buffer
	if err := f.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatINR(amount float64) string {
	paise := int64(amount*100 + 0.5)
	rupees := paise / 100
	dec := paise % 100

	s := fmt.Sprintf("%d", rupees)
	n := len(s)
	if n <= 3 {
		return fmt.Sprintf("%s.%02d", s, dec)
	}

	var parts []string
	// Last 3 digits
	parts = append(parts, s[n-3:])
	s = s[:n-3]
	// Remaining digits in groups of 2
	for len(s) > 2 {
		parts = append([]string{s[len(s)-2:]}, parts...)
		s = s[:len(s)-2]
	}
	if len(s) > 0 {
		parts = append([]string{s}, parts...)
	}

	return fmt.Sprintf("%s.%02d", strings.Join(parts, ","), dec)
}