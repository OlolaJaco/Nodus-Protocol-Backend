package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// PageParams holds validated pagination values parsed from query strings.
type PageParams struct {
	Page   int
	Limit  int
	Offset int
}

// ParsePageParams extracts and validates ?page= and ?limit= from the request.
func ParsePageParams(c *gin.Context) PageParams {
	page := parseInt(c.DefaultQuery("page", "1"), DefaultPage)
	limit := parseInt(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)), DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	return PageParams{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

// TotalPages returns the number of pages given a total record count and limit.
func TotalPages(total int64, limit int) int64 {
	if limit <= 0 {
		return 0
	}
	return (total + int64(limit) - 1) / int64(limit)
}

func parseInt(s string, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}
