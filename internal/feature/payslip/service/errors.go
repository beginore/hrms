package service

import "errors"

var (
	ErrInvalidPayslipID     = errors.New("invalid payslip id")
	ErrInvalidPayrollItemID = errors.New("invalid payroll item id")
	ErrInvalidCycleID       = errors.New("invalid cycle id")
	ErrInvalidEmployeeID    = errors.New("invalid employee id")
	ErrInvalidDateFormat    = errors.New("invalid date format")
	ErrInvalidStatus        = errors.New("invalid payslip status")
	ErrInvalidCycleStatus   = errors.New("payroll cycle is not ready for payslip generation")
	ErrReviewRequired       = errors.New("payroll item requires review")
	ErrPayslipNotFound      = errors.New("payslip not found")
	ErrPayrollItemNotFound  = errors.New("payroll item not found")
	ErrCycleNotFound        = errors.New("payroll cycle not found")
	ErrForbidden            = errors.New("forbidden")
	ErrPayslipLocked        = errors.New("payslip is locked")
	ErrDatabaseQueryFailed  = errors.New("failed to query database")
	ErrDatabaseSaveFailed   = errors.New("failed to save data")
	ErrRenderFailed         = errors.New("failed to render payslip")
	ErrSendFailed           = errors.New("failed to send payslip")
)
