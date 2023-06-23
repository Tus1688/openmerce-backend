// Copyright (c) 2023. Tus1688
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package freight

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var BaseUrl string
var Authorization string

func (r *CalculateFreightRequest) CalculateFreight() (WholeResult, error) {
	url := BaseUrl + "/api/v1/internal/rate-complex"
	body, err := json.Marshal(r)
	if err != nil {
		return WholeResult{}, err
	}
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		return WholeResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", Authorization)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return WholeResult{}, err
	}
	if res.StatusCode == 404 {
		return WholeResult{}, fmt.Errorf("there are no rates available for this route")
	}
	defer res.Body.Close()
	var result WholeResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return WholeResult{}, err
	}
	return result, nil
}

func (r *PrecalculateFreightRequest) CalculateFreight() (WholeResult, error) {
	url := BaseUrl + "/api/v1/internal/rate-complex-precalculate"
	body, err := json.Marshal(r)
	if err != nil {
		return WholeResult{}, err
	}
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(body))
	if err != nil {
		return WholeResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", Authorization)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return WholeResult{}, err
	}
	if res.StatusCode == 404 {
		return WholeResult{}, fmt.Errorf("there are no rates available for this route")
	}
	defer res.Body.Close()
	var result WholeResult
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return WholeResult{}, err
	}
	return result, nil
}
