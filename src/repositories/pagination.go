/*
 *   Corona-Warn-App / cwa-map-backend
 *
 *   (C) 2020, T-Systems International GmbH
 *
 *   Deutsche Telekom AG and all other contributors /
 *   copyright owners license this file to you under the Apache
 *   License, Version 2.0 (the "License"); you may not use this
 *   file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing,
 *   software distributed under the License is distributed on an
 *   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *   KIND, either express or implied.  See the License for the
 *   specific language governing permissions and limitations
 *   under the License.
 */

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
