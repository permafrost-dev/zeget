package app

type ReturnStatusCode int

const (
	Success    ReturnStatusCode = iota
	Failure    ReturnStatusCode = -1
	FatalError ReturnStatusCode = -2
)

type ReturnStatus struct {
	Code ReturnStatusCode
	Err  error
	Msg  string
}

func NewReturnStatus(code ReturnStatusCode, err error, msg string) *ReturnStatus {
	return &ReturnStatus{
		Code: code,
		Err:  err,
		Msg:  msg,
	}
}
