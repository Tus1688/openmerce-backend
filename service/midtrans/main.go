package midtrans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var ServerKey string
var ServerKeyEncoded string
var BaseUrl string

// BaseOrderId is used to prefix the order id in database
// for example if the order id is 1, then the order id in midtrans is "something-1"
var BaseOrderId string

func (r *RequestSnap) CreatePayment() (ResponseSnap, error) {
	url := BaseUrl + "/snap/v1/transactions"
	body, err := json.Marshal(r)
	if err != nil {
		return ResponseSnap{}, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return ResponseSnap{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+ServerKeyEncoded)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return ResponseSnap{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 201 {
		var result ResponseErrorSnap
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return ResponseSnap{}, err
		}
		return ResponseSnap{}, fmt.Errorf("%v", result.ErrorMessages)
	}
	var result ResponseSnap
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return ResponseSnap{}, err
	}
	return result, nil
}
