package apperror

type Error struct {
	HTTPStatusCode int
	Message        string
	Err            error
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}
