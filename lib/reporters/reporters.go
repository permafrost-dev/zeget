package reporters

type Reporter interface {
	Report(input ...interface{}) error
}
