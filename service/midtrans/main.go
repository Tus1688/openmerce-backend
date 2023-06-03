package midtrans

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/logging"
	"github.com/gin-gonic/gin"
)

var ServerKey string
var ServerKeyEncoded string
var BaseUrlSnap string
var BaseUrlCoreApi string

// BaseOrderId is used to prefix the order id in database
// for example if the order id is 1, then the order id in midtrans is "something-1"
var BaseOrderId string

func (r *RequestSnap) CreatePayment() (ResponseSnap, error) {
	url := BaseUrlSnap + "/snap/v1/transactions"
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

func DeleteOrder(orderId string) error {
	url := BaseUrlCoreApi + "/v2/" + BaseOrderId + "-" + orderId + "/cancel"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+ServerKeyEncoded)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		var result ResponseErrorDeleteOrder
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return err
		}
		return fmt.Errorf("%v", result.StatusMessage)
	}
	return nil
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
				go logging.InsertLog(logging.ERROR, "midtrans webhook error: unable to update order :"+OrderId)
				// retry once
				c.Status(500)
				return
			}
			go stockHandler(OrderId)
		}
		// if the request.TransactionStatus is "cancel" or "deny" or "expire"
		// set the is_cancelled to true
	} else if request.TransactionStatus == "cancel" || request.TransactionStatus == "deny" || request.TransactionStatus == "expire" {
		_, err := database.MysqlInstance.
			Exec("UPDATE orders SET is_cancelled = true, transaction_status = ? WHERE id = ?", request.TransactionStatus, OrderId)
		if err != nil {
			go logging.InsertLog(logging.ERROR, "midtrans webhook error: unable to update order :"+OrderId)
			// retry once
			c.Status(500)
			return
		}
	} else {
		// otherwise, update the transaction_status only
		_, err := database.MysqlInstance.
			Exec("UPDATE orders SET transaction_status = ? WHERE id = ?", request.TransactionStatus, OrderId)
		if err != nil {
			go logging.InsertLog(logging.ERROR, "midtrans webhook error: unable to update order :"+OrderId)
			// retry once
			c.Status(500)
			return
		}
	}

	c.Status(200)
}

// stockHandler is used to handle the stock and supposed to run in another goroutine
func stockHandler(orderID string) {
	// get the items in order first
	var items []productHelper
	rows, err := database.MysqlInstance.
		Query("SELECT BIN_TO_UUID(product_refer), quantity FROM order_items WHERE order_refer = ?", orderID)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to get items in order :"+orderID)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item productHelper
		err := rows.Scan(&item.productID, &item.quantity)
		if err != nil {
			go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to scan items in order :"+orderID)
			return
		}
		items = append(items, item)
	}

	// acquire the lock to prevent race condition
	tx, err := database.MysqlInstance.Begin()
	if err != nil {
		go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to begin transaction, order :"+orderID)
		return
	}
	defer tx.Rollback()
	// lock the rows of inventories table
	_, err = tx.Exec("SELECT * FROM inventories WHERE id IN (SELECT product_refer FROM order_items WHERE order_refer = ?) FOR UPDATE", orderID)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to lock rows of inventories table, order :"+orderID)
		return
	}
	// update and make sure the stock after decreased is not negative
	for _, item := range items {
		// set the current quantity in inventories into quantity - item.quantity where id = item.productID and quantity >= item.quantity
		res, err := tx.
			Exec("UPDATE inventories SET quantity = quantity - ? WHERE product_refer = UUID_TO_BIN(?) AND quantity >= ?", item.quantity, item.productID, item.quantity)
		if err != nil {
			go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to update inventories table, order :"+orderID)
			return
		}
		// if the affected rows is 0, then the quantity is not enough
		affected, err := res.RowsAffected()
		if err != nil {
			go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to get affected rows, order :"+orderID)
			return
		}
		// abort the transaction and update the order status to deny and set the need_refund to true
		if affected == 0 {
			// run the query on different transaction
			_, err := database.MysqlInstance.
				Exec("UPDATE orders SET transaction_status = 'deny', need_refund = true, status_description = 'sorry, stock is not enough to fulfill the orders' WHERE id = ?", orderID)
			if err != nil {
				go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to update order status to deny and need_refund to true, order :"+orderID)
				return
			}
			return
		}
	}
	// commit the transaction
	if err := tx.Commit(); err != nil {
		go logging.InsertLog(logging.ERROR, "midtrans stock handler error: unable to commit transaction, order :"+orderID)
		return
	}
}

type productHelper struct {
	productID string
	quantity  int
}
