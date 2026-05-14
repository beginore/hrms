package service

import "errors"

var (
	ErrInvalidDateRange      = errors.New("end date must be after or equal to start date")
	ErrInvalidDateFormat     = errors.New("invalid date format")
	ErrInvalidCycleID        = errors.New("invalid cycle id")
	ErrInvalidEmployeeID     = errors.New("invalid employee id")
	ErrInvalidAdjustmentID   = errors.New("invalid adjustment id")
	ErrInvalidTaxRuleID      = errors.New("invalid tax rule id")
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidRate           = errors.New("invalid tax rate")
	ErrInvalidAdjustmentType = errors.New("invalid adjustment type")
	ErrInvalidTaxAppliesTo   = errors.New("invalid tax applies_to")
	ErrInvalidTaxPayer       = errors.New("invalid tax payer")
	ErrInvalidPolicy         = errors.New("invalid payroll policy")
	ErrCycleNotFound         = errors.New("payroll cycle not found")
	ErrPayrollItemNotFound   = errors.New("payroll item not found")
	ErrCycleLocked           = errors.New("payroll cycle is locked")
	ErrNoWorkingDays         = errors.New("payroll period has no working days")
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrInvalidSalaryRate     = errors.New("invalid salary rate")
	ErrForbidden             = errors.New("forbidden")
	ErrDatabaseQueryFailed   = errors.New("failed to query database")
	ErrDatabaseSaveFailed    = errors.New("failed to save data")
)
