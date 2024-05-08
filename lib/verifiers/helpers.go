package verifiers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
)

type HashAlgorithm string

const (
	MD5     HashAlgorithm = "md5"
	SHA1    HashAlgorithm = "sha1"
	SHA256  HashAlgorithm = "sha256"
	SHA512  HashAlgorithm = "sha512"
	Unknown HashAlgorithm = "unknown"
)

func DetermineHashTypeByLength(hash string) HashAlgorithm {
	if len(hash) == md5.Size {
		return "md5"
	} else if len(hash) == sha1.Size {
		return "sha1"
	} else if len(hash) == sha256.Size {
		return "sha256"
	} else if len(hash) == sha512.Size {
		return "sha512"
	}

	return "unknown"
}
