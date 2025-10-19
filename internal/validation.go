package internal

import "strings"


type ValidationErr struct {
	Details []string
}

func (v ValidationErr) Error() string {
	return "Failed to validate: " + strings.Join(v.Details, ", ")
}
