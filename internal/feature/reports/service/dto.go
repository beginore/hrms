package service

type DashboardResponse struct {
	TotalEmployees       int    `json:"total_employees"`
	ActiveEmployees      int    `json:"active_employees"`
	PresentToday         int    `json:"present_today"`
	LateToday            int    `json:"late_today"`
	OnLeaveToday         int    `json:"on_leave_today"`
	AbsentToday          int    `json:"absent_today"`
	MonthlyPayroll       string `json:"monthly_payroll"`
	PayrollCurrency      string `json:"payroll_currency"`
	EmployeeGrowth       int    `json:"employee_growth"`
	PayrollGrowthPercent string `json:"payroll_growth_percent"`
	AttendanceRate       string `json:"attendance_rate"`
}

type PayrollSummaryResponse struct {
	PeriodStart        string `json:"period_start"`
	PeriodEnd          string `json:"period_end"`
	CyclesCount        int    `json:"cycles_count"`
	EmployeesPaid      int    `json:"employees_paid"`
	GrossSalaryTotal   string `json:"gross_salary_total"`
	NetSalaryTotal     string `json:"net_salary_total"`
	BonusesTotal       string `json:"bonuses_total"`
	DeductionsTotal    string `json:"deductions_total"`
	TaxesTotal         string `json:"taxes_total"`
	EmployerTaxesTotal string `json:"employer_taxes_total"`
	TotalEmployerCost  string `json:"total_employer_cost"`
	AverageNetSalary   string `json:"average_net_salary"`
	Currency           string `json:"currency"`
}

type ChartSeries struct {
	Name string    `json:"name"`
	Data []float64 `json:"data"`
}

type LineChartResponse struct {
	Labels []string      `json:"labels"`
	Series []ChartSeries `json:"series"`
}

type PieItem struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type PieChartResponse struct {
	Items []PieItem `json:"items"`
}

type DepartmentPayrollResponse struct {
	Rows  []DepartmentPayrollItem `json:"rows"`
	Chart BarChartResponse        `json:"chart"`
}

type BarChartResponse struct {
	Labels []string      `json:"labels"`
	Series []ChartSeries `json:"series"`
}

type DepartmentPayrollItem struct {
	DepartmentID       *string `json:"department_id,omitempty"`
	DepartmentName     string  `json:"department_name"`
	EmployeesCount     int     `json:"employees_count"`
	AverageSalary      string  `json:"average_salary"`
	GrossSalaryTotal   string  `json:"gross_salary_total"`
	NetSalaryTotal     string  `json:"net_salary_total"`
	BonusesTotal       string  `json:"bonuses_total"`
	DeductionsTotal    string  `json:"deductions_total"`
	TaxesTotal         string  `json:"taxes_total"`
	EmployerTaxesTotal string  `json:"employer_taxes_total"`
	TotalEmployerCost  string  `json:"total_employer_cost"`
	Currency           string  `json:"currency"`
}

type AttendanceTodayResponse struct {
	Date           string           `json:"date"`
	TotalActive    int              `json:"total_active"`
	Present        int              `json:"present"`
	Late           int              `json:"late"`
	OnLeave        int              `json:"on_leave"`
	Absent         int              `json:"absent"`
	Remote         int              `json:"remote"`
	AttendanceRate string           `json:"attendance_rate"`
	Chart          PieChartResponse `json:"chart"`
}

type EmployeeStatisticsResponse struct {
	TotalEmployees    int              `json:"total_employees"`
	ActiveEmployees   int              `json:"active_employees"`
	InactiveEmployees int              `json:"inactive_employees"`
	AverageSalary     string           `json:"average_salary"`
	ByDepartment      BarChartResponse `json:"by_department"`
	ByRole            PieChartResponse `json:"by_role"`
	ByStatus          PieChartResponse `json:"by_status"`
}

type DepartmentStatisticsResponse struct {
	Rows                []DepartmentStatisticsItem `json:"rows"`
	EmployeeChart       BarChartResponse           `json:"employee_chart"`
	AttendanceRateChart BarChartResponse           `json:"attendance_rate_chart"`
	PayrollCostChart    BarChartResponse           `json:"payroll_cost_chart"`
}

type DepartmentStatisticsItem struct {
	DepartmentID      *string `json:"department_id,omitempty"`
	DepartmentName    string  `json:"department_name"`
	EmployeesCount    int     `json:"employees_count"`
	AverageSalary     string  `json:"average_salary"`
	AttendanceRate    string  `json:"attendance_rate"`
	AbsentDays        int     `json:"absent_days"`
	LateDays          int     `json:"late_days"`
	OvertimeMinutes   int     `json:"overtime_minutes"`
	NetSalaryTotal    string  `json:"net_salary_total"`
	TotalEmployerCost string  `json:"total_employer_cost"`
	Currency          string  `json:"currency"`
}

type ExportReportResponse struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"-"`
}
