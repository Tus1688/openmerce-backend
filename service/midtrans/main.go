package midtrans

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/gin-gonic/gin"
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

func HandleNotifications(c *gin.Context) {
	var request WebhookNotification
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// verify the signature
	// SHA512(order_id+status_code+gross_amount+ServerKey)
	hash := sha512.New()
	hash.Write([]byte(request.OrderId + request.StatusCode + request.GrossAmount + ServerKey))
	signature := fmt.Sprintf("%x", hash.Sum(nil))
	// if not verified, return 401
	if signature != request.SignatureKey {
		c.Status(401)
		return
	}
	// strip the BaseOrderId+"-" from the order id
	// for example if the order id is "something-1", then the order id in database is 1
	OrderId := request.OrderId[len(BaseOrderId)+1:]
	// if there is FraudStatus, always check if it is "accept"
	// if transaction_status value is settlement or capture change the is_paid to true
	if request.TransactionStatus == "settlement" || request.TransactionStatus == "capture" {
		if request.FraudStatus != "deny" && request.FraudStatus != "challenge" {
			// update the is_paid to true
			_, err := database.MysqlInstance.
				Exec("UPDATE orders SET is_paid = true, transaction_status = ? WHERE id = ?", request.TransactionStatus, OrderId)
			if err != nil {
				log.Print(err)
				// retry once
				c.Status(500)
				return
			}
		}
		// if the request.TransactionStatus is "cancel" or "deny" or "expire"
		// set the is_cancelled to true
	} else if request.TransactionStatus == "cancel" || request.TransactionStatus == "deny" || request.TransactionStatus == "expire" {
		_, err := database.MysqlInstance.
			Exec("UPDATE orders SET is_cancelled = true, transaction_status = ? WHERE id = ?", request.TransactionStatus, OrderId)
		if err != nil {
			log.Print(err)
			// retry once
			c.Status(500)
			return
		}
	} else {
		// otherwise, update the transaction_status only
		_, err := database.MysqlInstance.
			Exec("UPDATE orders SET transaction_status = ? WHERE id = ?", request.TransactionStatus, OrderId)
		if err != nil {
			log.Print(err)
			// retry once
			c.Status(500)
			return
		}
	}

	c.Status(200)
}
