package download

import "os"

func isLocalFile(s string) bool {
	_, err := os.Stat(s)
	return err == nil
}

func setIf[T interface{}](condition bool, original T, newValue T) T {
	if condition {
		return newValue
	}
	return original
}
