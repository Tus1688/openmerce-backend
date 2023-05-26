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
