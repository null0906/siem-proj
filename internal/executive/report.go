package executive

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

func RenderPDF(summary Summary) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(16, 15, 16)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()

	setFill(pdf, 7, 12, 18)
	pdf.Rect(0, 0, 210, 297, "F")
	setText(pdf, 238, 244, 251)

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 8, "SecComply Executive Security Report", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	setText(pdf, 133, 149, 171)
	pdf.CellFormat(0, 5, fmt.Sprintf("Board summary for the %s  |  Generated %s UTC", summary.PeriodLabel, summary.GeneratedAt.Format("02 Jan 2006 15:04")), "", 1, "L", false, 0, "")
	pdf.Ln(3)
	line(pdf)

	pdf.SetFont("Helvetica", "B", 10)
	setText(pdf, 133, 149, 171)
	pdf.CellFormat(0, 6, "SECURITY POSTURE", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 31)
	setText(pdf, 0, 196, 255)
	pdf.CellFormat(42, 13, fmt.Sprintf("%d / 100", summary.Posture.Score), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 16)
	setText(pdf, 238, 244, 251)
	pdf.CellFormat(25, 13, "Grade "+summary.Posture.Grade, "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 11)
	if summary.Posture.Delta >= 0 {
		setText(pdf, 0, 255, 136)
		pdf.CellFormat(0, 13, fmt.Sprintf("+%d points this period", summary.Posture.Delta), "", 1, "R", false, 0, "")
	} else {
		setText(pdf, 255, 90, 90)
		pdf.CellFormat(0, 13, fmt.Sprintf("%d points this period", summary.Posture.Delta), "", 1, "R", false, 0, "")
	}
	pdf.SetFont("Helvetica", "", 10)
	setText(pdf, 205, 216, 230)
	pdf.MultiCell(0, 5, ascii(summary.ExecutiveMessage), "", "L", false)
	pdf.Ln(4)

	sectionTitle(pdf, "What changed")
	for _, highlight := range summary.Changes.Highlights {
		bullet(pdf, highlight)
	}
	pdf.Ln(3)

	sectionTitle(pdf, "Top business risks")
	for _, risk := range summary.TopRisks {
		pdf.SetFont("Helvetica", "B", 10)
		setText(pdf, 238, 244, 251)
		pdf.CellFormat(9, 5, fmt.Sprintf("%d.", risk.Rank), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 5, ascii(risk.Title), "", 1, "L", false, 0, "")
		pdf.SetX(25)
		pdf.SetFont("Helvetica", "", 8.5)
		setText(pdf, 133, 149, 171)
		pdf.MultiCell(0, 4.2, ascii(risk.Summary), "", "L", false)
		pdf.Ln(1.2)
	}

	sectionTitle(pdf, "Compliance readiness estimate")
	for _, framework := range summary.Frameworks {
		pdf.SetFont("Helvetica", "B", 9)
		setText(pdf, 205, 216, 230)
		pdf.CellFormat(34, 6, framework.Name, "", 0, "L", false, 0, "")
		setFill(pdf, 25, 43, 64)
		pdf.Rect(pdf.GetX(), pdf.GetY()+2, 95, 2.5, "F")
		setFill(pdf, 0, 196, 255)
		pdf.Rect(pdf.GetX(), pdf.GetY()+2, 95*float64(framework.Score)/100, 2.5, "F")
		pdf.SetX(pdf.GetX() + 100)
		pdf.SetFont("Helvetica", "B", 9)
		setText(pdf, 238, 244, 251)
		pdf.CellFormat(14, 6, fmt.Sprintf("%d%%", framework.Score), "", 0, "R", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		setText(pdf, 133, 149, 171)
		pdf.CellFormat(0, 6, framework.State, "", 1, "R", false, 0, "")
	}

	pdf.SetY(-20)
	line(pdf)
	pdf.SetFont("Helvetica", "", 7.5)
	setText(pdf, 84, 101, 125)
	pdf.CellFormat(0, 5, "Confidential  |  Readiness figures are live operational estimates, not audit opinions.", "", 0, "L", false, 0, "")
	pdf.CellFormat(0, 5, time.Now().UTC().Format("2006-01-02"), "", 0, "R", false, 0, "")

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func sectionTitle(pdf *fpdf.Fpdf, title string) {
	line(pdf)
	pdf.Ln(2)
	pdf.SetFont("Helvetica", "B", 11)
	setText(pdf, 238, 244, 251)
	pdf.CellFormat(0, 6, title, "", 1, "L", false, 0, "")
	pdf.Ln(1)
}

func bullet(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont("Helvetica", "B", 9)
	setText(pdf, 0, 196, 255)
	pdf.CellFormat(6, 5, "-", "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	setText(pdf, 205, 216, 230)
	pdf.MultiCell(0, 5, ascii(text), "", "L", false)
}

func line(pdf *fpdf.Fpdf) {
	pdf.SetDrawColor(25, 56, 78)
	pdf.Line(16, pdf.GetY(), 194, pdf.GetY())
}

func setText(pdf *fpdf.Fpdf, r, g, b int) {
	pdf.SetTextColor(r, g, b)
}

func setFill(pdf *fpdf.Fpdf, r, g, b int) {
	pdf.SetFillColor(r, g, b)
}

func ascii(value string) string {
	replacer := strings.NewReplacer("—", "-", "–", "-", "▲", "+", "▼", "-", "·", "-")
	return replacer.Replace(value)
}
