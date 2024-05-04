package download

import "os"

func isLocalFile(s string) bool {
	_, err := os.Stat(s)
	return err == nil
}
