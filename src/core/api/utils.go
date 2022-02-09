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
