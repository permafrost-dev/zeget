package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Error struct {
	Code   int
	Status string
	Body   []byte
	URL    string
}

type ErrorResponse struct {
	Message string `json:"message"`
	Doc     string `json:"documentation_url"`
}

func (ge *Error) Error() string {
	var msg ErrorResponse
	json.Unmarshal(ge.Body, &msg)

	if ge.Code == http.StatusForbidden {
		return fmt.Sprintf("%s: %s: %s", ge.Status, msg.Message, msg.Doc)
	}

	return fmt.Sprintf("%s (URL: %s)", ge.Status, ge.URL)
}
