package api

import (
	"net/http"
	"strconv"
)

func GetFloatParameter(r *http.Request, name string) (float64, bool, error) {
	values, ok := r.URL.Query()[name]
	if !ok {
		return 0, false, nil
	}

	if value, err := strconv.ParseFloat(values[0], 64); err == nil {
		return value, true, nil
	} else {
		return 0, false, err
	}
}
