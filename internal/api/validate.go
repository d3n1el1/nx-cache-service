package api

import (
	"fmt"
	"regexp"
)

var validPathSegment = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func validatePathParam(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", name)
	}

	if value == "." || value == ".." {
		return fmt.Errorf("%s must not be '.' or '..'", name)
	}

	if !validPathSegment.MatchString(value) {
		return fmt.Errorf("%s contains invalid characters", name)
	}

	return nil
}
