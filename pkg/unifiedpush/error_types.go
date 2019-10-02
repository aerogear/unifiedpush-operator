package unifiedpush

import (
	"fmt"
	"strings"
)

//CreateError wraps some extra messaging that we may receive if UPS returns an error.
type CreateError struct {
	Errors     map[string]string
	Message    string
	StatusCode int
}

func (e CreateError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Received HTTP status %d with message %s ", e.StatusCode, e.Message))
	for key, value := range e.Errors {
		sb.WriteString(fmt.Sprintf("\t %s:%s ", key, value))
	}
	return sb.String()
}
