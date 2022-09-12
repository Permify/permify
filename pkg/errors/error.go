package errors

import (
	"bytes"
	"fmt"
	"text/template"
)

var (
	ValidationError     = NewError(Validation)
	ServiceError        = NewError(Service)
	CircuitBreakerError = NewError(Service)
	DatabaseError       = NewError(Database)
)

var (
	Validation Kind = "validation"
	Service    Kind = "service"
	Database   Kind = "database"
)

type (
	Kind string

	Error interface {
		Error() string
		Kind() Kind
		SetKind(Kind) Error
		SubKind() Kind
		SetSubKind(Kind) Error
		Message() string
		SetMessage(string) Error
		Params() map[string]interface{}
		AddParam(name string, value interface{}) Error
		SetParams(map[string]interface{}) Error
	}

	ErrorObject struct {
		kind    Kind
		subKind Kind
		message string
		params  map[string]interface{}
	}

	Errors map[string]error
)

// String -
func (k Kind) String() string {
	return string(k)
}

// NewError -
func NewError(kind Kind) Error {
	return ErrorObject{
		kind: kind,
	}
}

// SetKind -
func (e ErrorObject) SetKind(kind Kind) Error {
	e.kind = kind
	return e
}

// Kind -
func (e ErrorObject) Kind() Kind {
	return e.kind
}

// SetSubKind -
func (e ErrorObject) SetSubKind(kind Kind) Error {
	e.subKind = kind
	return nil
}

// SubKind -
func (e ErrorObject) SubKind() Kind {
	return e.subKind
}

// SetParams -
func (e ErrorObject) SetParams(params map[string]interface{}) Error {
	e.params = params
	return e
}

// AddParam -
func (e ErrorObject) AddParam(name string, value interface{}) Error {
	if e.params == nil {
		e.params = make(map[string]interface{})
	}

	e.params[name] = value
	return e
}

// Params -
func (e ErrorObject) Params() map[string]interface{} {
	if e.params == nil {
		//e.params = map[string]interface{}{
		//	"info": e.Error(),
		//}
	}
	return e.params
}

// SetMessage -
func (e ErrorObject) SetMessage(message string) Error {
	e.message = message
	return e
}

// Message -
func (e ErrorObject) Message() string {
	return e.message
}

// Error -
func (e ErrorObject) Error() string {
	msg := e.message
	if msg == "" {
		msg = fmt.Sprintf("message: %s%s", e.kind.String(), ", "+e.subKind.String())
	}
	if len(e.params) == 0 {
		return msg
	}
	res := bytes.Buffer{}
	_ = template.Must(template.New("err").Parse(msg)).Execute(&res, e.params)
	return res.String()
}

var _ Error = ErrorObject{}
