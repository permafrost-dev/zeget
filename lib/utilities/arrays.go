package utilities

func FilterArr[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func IsInArr[T any](arr []T, val T, test func(T, T) bool) bool {
	for _, a := range arr {
		if test(a, val) {
			return true
		}
	}
	return false
}
