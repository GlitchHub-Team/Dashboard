package common

import "fmt"

func PageLimitToOffset(page int, limit int) (int, error) {
	if page < 1 {
		return 0, fmt.Errorf("cannot retrieve page %v", page)
	}
	if limit < 1 {
		return 0, fmt.Errorf("cannot retrieve limit %v", limit)
	}

	return limit * (page - 1), nil
}
