package service

import (
	"bytes"
	"fmt"
	"strings"
)

func renderReportPDF(lines []string) []byte {
	const (
		fontSize   = 10
		lineHeight = 15
		left       = 48
		top        = 760
	)

	var content bytes.Buffer
	content.WriteString("BT\n")
	content.WriteString("/F1 10 Tf\n")
	content.WriteString(fmt.Sprintf("%d %d Td\n", left, top))
	for i, line := range wrapReportLines(lines) {
		if i == 0 {
			content.WriteString("/F1 18 Tf\n")
		} else if line == "" {
			content.WriteString("/F1 10 Tf\n")
		} else if isReportHeading(line) {
			content.WriteString("/F1 13 Tf\n")
		} else {
			content.WriteString(fmt.Sprintf("/F1 %d Tf\n", fontSize))
		}
		content.WriteString("(")
		content.WriteString(escapePDFText(line))
		content.WriteString(") Tj\n")
		content.WriteString(fmt.Sprintf("0 -%d Td\n", lineHeight))
	}
	content.WriteString("ET\n")

	stream := content.String()
	objects := []string{
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n",
		"2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n",
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj\n",
		"4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n",
		fmt.Sprintf("5 0 obj << /Length %d >> stream\n%s\nendstream endobj\n", len(stream), stream),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		buf.WriteString(obj)
	}
	xrefOffset := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString(fmt.Sprintf("trailer << /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefOffset))
	return buf.Bytes()
}

func wrapReportLines(lines []string) []string {
	const maxLen = 92
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) <= maxLen {
			out = append(out, line)
			continue
		}
		for len(line) > maxLen {
			cut := strings.LastIndex(line[:maxLen], " ")
			if cut < 40 {
				cut = maxLen
			}
			out = append(out, line[:cut])
			line = strings.TrimSpace(line[cut:])
		}
		out = append(out, line)
	}
	if len(out) > 45 {
		out = append(out[:44], "Report truncated for PDF preview. Use CSV export for full rows.")
	}
	return out
}

func isReportHeading(line string) bool {
	switch line {
	case "Summary", "Department Payroll", "Attendance Records":
		return true
	default:
		return false
	}
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	return value
}
