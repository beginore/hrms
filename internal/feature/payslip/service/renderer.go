package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func renderPayslipPDF(payload []byte) ([]byte, error) {
	lines, err := payslipLines(payload)
	if err != nil {
		return nil, err
	}
	return simplePDF(lines), nil
}

func renderPayslipEmailBody(payload []byte) (string, error) {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return "", err
	}
	employeeName := nestedString(data, "employee", "full_name")
	periodStart := nestedString(data, "period", "start")
	periodEnd := nestedString(data, "period", "end")
	netSalary := nestedString(data, "amounts", "net_salary")
	currency := nestedString(data, "period", "currency")
	return fmt.Sprintf(`Hello %s,

Your payslip for %s - %s has been generated.

Net salary: %s %s

You can view the payslip in Smart EMP.

Best regards,
Smart EMP Payroll Team`, employeeName, periodStart, periodEnd, netSalary, currency), nil
}

func payslipLines(payload []byte) ([]string, error) {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	currency := nestedString(data, "period", "currency")
	lines := []string{
		"Smart EMP PAYSLIP",
		"====================================================================",
		"Payslip ID: " + stringValue(data["payslip_id"]),
		"Generated at: " + stringValue(data["generated_at"]),
		"",
		"ORGANIZATION",
		pdfRow("Name", nestedString(data, "organization", "name")),
		pdfRow("VAT", nestedString(data, "organization", "vat_id")),
		pdfRow("Address", nestedString(data, "organization", "address")),
		"",
		"EMPLOYEE",
		pdfRow("Full name", nestedString(data, "employee", "full_name")),
		pdfRow("Email", nestedString(data, "employee", "email")),
		pdfRow("Department", nestedString(data, "employee", "department")),
		pdfRow("Position", nestedString(data, "employee", "position")),
		"",
		"PAY PERIOD",
		pdfRow("Dates", nestedString(data, "period", "start")+" - "+nestedString(data, "period", "end")),
		pdfRow("Cycle status", nestedString(data, "period", "status")),
		pdfRow("Currency", currency),
		"",
		"EARNINGS",
		pdfMoneyRow("Base salary", nestedString(data, "amounts", "base_salary"), currency),
		pdfMoneyRow("Overtime", nestedString(data, "amounts", "overtime_amount"), currency),
		pdfMoneyRow("Bonuses", nestedString(data, "amounts", "bonuses_total"), currency),
		pdfMoneyRow("Gross salary", nestedString(data, "amounts", "gross_salary"), currency),
		"",
		"DEDUCTIONS AND TAXES",
		pdfMoneyRow("Deductions", nestedString(data, "amounts", "deductions_total"), currency),
		pdfMoneyRow("Employee taxes", nestedString(data, "amounts", "taxes_total"), currency),
		"",
		"ATTENDANCE",
		pdfRow("Working days", nestedString(data, "attendance", "working_days")),
		pdfRow("Paid days", nestedString(data, "attendance", "paid_days")),
		pdfRow("Unpaid days", nestedString(data, "attendance", "unpaid_days")),
		pdfRow("Late days", nestedString(data, "attendance", "late_days")),
		pdfRow("Absent days", nestedString(data, "attendance", "absent_days")),
		pdfRow("Overtime minutes", nestedString(data, "attendance", "overtime_minutes")),
		"",
		"FINAL",
		pdfMoneyRow("NET SALARY", nestedString(data, "amounts", "net_salary"), currency),
		pdfMoneyRow("Employer taxes", nestedString(data, "amounts", "employer_taxes_total"), currency),
		pdfMoneyRow("Total employer cost", nestedString(data, "amounts", "total_employer_cost"), currency),
		"====================================================================",
	}

	if adjustments, ok := data["adjustments"].([]any); ok && len(adjustments) > 0 {
		lines = append(lines, "", "ADJUSTMENTS")
		for _, raw := range adjustments {
			if item, ok := raw.(map[string]any); ok {
				label := stringValue(item["type"]) + " / " + stringValue(item["category"])
				lines = append(lines, pdfMoneyRow(label, stringValue(item["amount"]), currency))
			}
		}
	}

	return lines, nil
}

func pdfRow(label, value string) string {
	if value == "" {
		value = "-"
	}
	return fmt.Sprintf("%-24s %s", label+":", value)
}

func pdfMoneyRow(label, amount, currency string) string {
	if amount == "" {
		amount = "0.00"
	}
	if currency == "" {
		return fmt.Sprintf("%-40s %15s", label, amount)
	}
	return fmt.Sprintf("%-40s %15s %s", label, amount, currency)
}

func nestedString(data map[string]any, keys ...string) string {
	var current any = data
	for _, key := range keys {
		asMap, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = asMap[key]
	}
	return stringValue(current)
}

func stringValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func simplePDF(lines []string) []byte {
	var content strings.Builder
	content.WriteString("BT\n/F1 12 Tf\n50 790 Td\n")
	for i, line := range lines {
		if i > 0 {
			content.WriteString("0 -16 Td\n")
		}
		content.WriteString("(")
		content.WriteString(escapePDFText(line))
		content.WriteString(") Tj\n")
	}
	content.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content.String()), content.String()),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		buf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", i+1, obj))
	}
	xrefOffset := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(objects)+1))
	buf.WriteString("0000000000 65535 f \n")
	keys := make([]int, 0, len(offsets)-1)
	for i := 1; i < len(offsets); i++ {
		keys = append(keys, i)
	}
	sort.Ints(keys)
	for _, i := range keys {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefOffset))
	return buf.Bytes()
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	return value
}
