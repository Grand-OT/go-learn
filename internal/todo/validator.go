package todo

import (
	"errors"
	"strings"
)

const maxTitleLen = 140

func validate(t *TodoCreateRequest) error {
	t.Title = strings.TrimSpace(t.Title)
	if len(t.Title) > maxTitleLen || len(t.Title) == 0 {
		return errors.New("title must not be empty or too long")
	}
	return nil
}
