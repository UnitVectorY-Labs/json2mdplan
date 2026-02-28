package diagnostics

import "fmt"

type Error struct {
	Code      string
	Directive int
	Path      string
	Message   string
}

func (e *Error) Error() string {
	return e.Message
}

func New(code string, directive int, path string, format string, args ...any) *Error {
	return &Error{
		Code:      code,
		Directive: directive,
		Path:      path,
		Message:   fmt.Sprintf(format, args...),
	}
}

func Format(err error) string {
	if err == nil {
		return ""
	}

	if e, ok := err.(*Error); ok {
		return fmt.Sprintf("code=%s\ndirective=%d\npath=%s\nmessage=%s", e.Code, e.Directive, e.Path, e.Message)
	}

	return fmt.Sprintf("code=error\ndirective=-1\npath=\nmessage=%s", err.Error())
}
