package pagination

import (
	"errors"
)

var (
	ErrInvalidPage  = errors.New("invalid page parameter")
	ErrInvalidLimit = errors.New("invalid limit parameter")
)

func PageLimitToOffset(page int, limit int) (int, error) {
	if page < 1 {
		return 0, ErrInvalidPage
	}
	if limit < 1 {
		return 0, ErrInvalidLimit
	}

	return limit * (page - 1), nil
}
