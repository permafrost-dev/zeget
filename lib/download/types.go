package download

type AcceptContentType string

const (
	AcceptBinary     AcceptContentType = "application/octet-stream"
	AcceptGitHubJSON AcceptContentType = "application/vnd.github+json"
	AcceptText       AcceptContentType = "text/plain"
)
