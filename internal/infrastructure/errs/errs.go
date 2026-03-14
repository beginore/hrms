package errs

import "encoding/json"

type Code uint

const (
	CodeAlreadyExists Code = iota
)

type Error struct {
	Message string `json:"message"`
	Code    Code   `json:"-"`
	Err     error  `json:"-"`
}

func New(code Code, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

func (e *Error) Error() string {
	data, err := json.Marshal(e)
	if err != nil {
		return "invalid error"
	}
	return string(data)
}

func (e *Error) WithCause(err error) *Error {
	e.Err = err
	return e
}

func (e *Error) SetCode(code Code) {
	e.Code = code
}

func (e *Error) SetCause(err error) {
	e.Err = err
}

func (e *Error) SetMessage(message string) {
	e.Message = message
}
