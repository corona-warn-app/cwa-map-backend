package repositories

import (
	"net/http"
	"strconv"
)

type PagedResult struct {
	Count int64
}

type PageRequest struct {
	Page int
	Size int
}

func ParsePageRequest(r *http.Request) PageRequest {
	result := PageRequest{
		Page: 0,
		Size: 50,
	}

	if param, hasParameter := r.URL.Query()["page"]; hasParameter {
		if value, err := strconv.ParseInt(param[0], 10, 32); err == nil {
			result.Page = int(value)
		}
	}

	if param, hasParameter := r.URL.Query()["size"]; hasParameter {
		if value, err := strconv.ParseInt(param[0], 10, 32); err == nil {
			result.Size = int(value)
		}
	}

	return result
}
