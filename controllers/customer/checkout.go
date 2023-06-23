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

package customer

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/logging"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/Tus1688/openmerce-backend/service/freight"
	"github.com/Tus1688/openmerce-backend/service/midtrans"
	"github.com/gin-gonic/gin"
)

func Checkout(c *gin.Context) {
	var request models.CheckoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	freightReq := freight.PrecalculateFreightRequest{}
	var itemGrossAmount int
	var items []models.CheckoutItemInternal
	errChan := make(chan error)
	wg := sync.WaitGroup{}
	// mutex for the freightReq and paymentReq
	mu := sync.Mutex{}
	wg.Add(3)

	// validate the freight pricing
	go func() {
		defer wg.Done()
		var id uint32
		err = database.MysqlInstance.
			QueryRow(
				"SELECT shipping_area_refer FROM customer_addresses WHERE id = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)",
				request.AddressCode, customerId,
			).
			Scan(&id)
		if err != nil {
			if err == sql.ErrNoRows {
				errChan <- fmt.Errorf("address not found")
				return
			}
			errChan <- err
			return
		}
		mu.Lock()
		freightReq.ID = id
		mu.Unlock()
	}()

	//	count all volume, weight, gross amount of the items in the cart
	go func() {
		defer wg.Done()
		var weight, volume float64
		err = database.MysqlInstance.
			QueryRow(
				`
				select sum(p.weight * c.quantity) as weight, sum((p.length * p.height * p.width) * c.quantity) as volume, sum(c.quantity * p.price) as gross_amount
				from products p, cart_items c
				left join inventories i on c.product_refer = i.product_refer
				where p.id = c.product_refer and c.checked = 1 and c.customer_refer = uuid_to_bin(?) and 
					c.quantity <= i.quantity and p.deleted_at is null
				group by c.customer_refer;
				`, customerId,
			).Scan(&weight, &volume, &itemGrossAmount)
		if err != nil {
			if err == sql.ErrNoRows {
				errChan <- fmt.Errorf("you haven't selected any item in the cart")
				return
			}
			errChan <- err
			return
		}
		mu.Lock()
		freightReq.Weight = weight
		freightReq.Volume = volume
		mu.Unlock()
	}()

	// fill the items that will be sent to the midtrans API
	go func() {
		defer wg.Done()
		rows, err := database.MysqlInstance.
			Query(
				`
				select BIN_TO_UUID(p.id), p.name, p.price, c.quantity, p.description, p.weight
				from products p, cart_items c
				left join inventories i on c.product_refer = i.product_refer
				where p.id = c.product_refer and c.customer_refer = UUID_TO_BIN(?)
				and c.checked = 1 and c.quantity <= i.quantity and p.deleted_at is null`, customerId,
			)
		if err != nil {
			errChan <- err
			return
		}
		defer rows.Close()
		for rows.Next() {
			var item models.CheckoutItemInternal
			err = rows.Scan(&item.Id, &item.Name, &item.Price, &item.Quantity, &item.Description, &item.Weight)
			if err != nil {
				errChan <- err
				return
			}
			items = append(items, item)
		}
	}()

	// wait for the initial data to be filled
	wg.Wait()

	// check if there is any error
	if len(errChan) > 0 {
		if err := <-errChan; err != nil {
			if strings.Contains(err.Error(), "address not found") {
				c.JSON(404, gin.H{"error": err.Error()})
				return
			} else if strings.Contains(err.Error(), "you haven't selected any item in the cart") {
				c.JSON(409, gin.H{"error": err.Error()})
				return
			} else {
				c.Status(500)
				go logging.InsertLog(logging.ERROR, "checkout1-"+err.Error())
				return
			}
		}
	}

	// get the freight pricing
	freightRes, err := freightReq.CalculateFreight()
	if err != nil {
		if strings.Contains(err.Error(), "there are no rates available for this route") {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout2-"+err.Error())
		return
	}
	// serialize the freight response into the list of choices
	var freightChoices []models.PreCheckoutFreight
	for _, value := range freightRes.Anteraja {
		freightChoices = append(
			freightChoices, models.PreCheckoutFreight{
				ProductCode: "anteraja-" + value.ProductCode,
				CourierName: "anteraja",
				ProductName: value.ProductName,
				Etd:         value.Etd,
				Rates:       value.Rates,
			},
		)
	}
	for _, value := range freightRes.Sicepat {
		freightChoices = append(
			freightChoices, models.PreCheckoutFreight{
				ProductCode: "sicepat-" + value.ProductCode,
				CourierName: "sicepat",
				ProductName: value.ProductName,
				Etd:         value.Etd,
				Rates:       value.Rates,
			},
		)
	}

	var freightCost int
	// check if the freight choice is valid and set the freightcost
	for _, choice := range freightChoices {
		if choice.ProductCode == request.CourierCode {
			freightCost = choice.Rates
			break
		}
	}
	if freightCost == 0 {
		c.JSON(409, gin.H{"error": "invalid courier code"})
		return
	}

	// prepare the transaction
	paymentReq := midtrans.RequestSnap{}
	wg.Add(1)

	//	acquire the customer's details
	go func() {
		defer wg.Done()
		var firstName, lastName, email, phone string
		err = database.MysqlInstance.
			QueryRow(
				"SELECT first_name, last_name, email, coalesce(phone_number, '') FROM customers WHERE id = UUID_TO_BIN(?)",
				customerId,
			).
			Scan(&firstName, &lastName, &email, &phone)
		if err != nil {
			errChan <- err
			return
		}
		paymentReq.CustomerDetails = midtrans.CustomerDetails{
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Phone:     phone,
		}
	}()

	// begin to make change into the database and payment gateway
	tx, err := database.MysqlInstance.Begin()
	if err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout3-"+err.Error())
		return
	}
	// rollback the transaction if there is any error
	defer tx.Rollback()
	// insert the order first into the database
	res, err := tx.
		Exec(
			`
			INSERT INTO orders (customer_refer, customer_address_refer, courier_code, freight_cost, item_cost, gross_amount)
			VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?, ?, ?)`, customerId, request.AddressCode, request.CourierCode,
			freightCost, itemGrossAmount, itemGrossAmount+freightCost,
		)
	if err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout4-"+err.Error())
		return
	}
	orderId, err := res.LastInsertId()
	if err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout5-"+err.Error())
		return
	}
	// insert the order items into the database
	stmt, err := tx.
		Prepare(
			`
			INSERT INTO order_items(order_refer, product_refer, on_buy_name, on_buy_description, on_buy_price, on_buy_weight, quantity)
			VALUES (?, UUID_TO_BIN(?), ?, ?, ?, ?, ?)
		`,
		)
	if err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout6-"+err.Error())
		return
	}
	defer stmt.Close()
	for _, item := range items {
		_, err := stmt.Exec(orderId, item.Id, item.Name, item.Description, item.Price, item.Weight, item.Quantity)
		if err != nil {
			c.Status(500)
			go logging.InsertLog(logging.ERROR, "checkout7-"+err.Error())
			return
		}
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()

	//check if there is error in errorChan
	for err := range errChan {
		if err != nil {
			c.Status(500)
			go logging.InsertLog(logging.ERROR, "checkout8-"+err.Error())
			return
		}
	}

	// fill the paymentReq.TransactionDetails based on the orderID
	paymentReq.TransactionDetails = midtrans.TransactionDetails{
		OrderId:     midtrans.BaseOrderId + "-" + strconv.FormatInt(orderId, 10),
		GrossAmount: itemGrossAmount + freightCost,
	}
	// fill the paymentReq.ItemDetails
	for _, item := range items {
		var name string
		if len(item.Name) > 50 {
			name = item.Name[:50]
		} else {
			name = item.Name
		}
		paymentReq.ItemDetails = append(
			paymentReq.ItemDetails, models.CheckoutItem{
				Id:       item.Id,
				Name:     name,
				Price:    item.Price,
				Quantity: item.Quantity,
			},
		)
	}
	paymentReq.ItemDetails = append(
		paymentReq.ItemDetails, models.CheckoutItem{
			Id:       "freight-" + midtrans.BaseOrderId + "-" + strconv.FormatInt(orderId, 10),
			Name:     request.CourierCode,
			Price:    freightCost,
			Quantity: 1,
		},
	)
	// already filled the customer's details above
	// set the expiry time into 1 day
	paymentReq.Expiry = midtrans.Expiry{
		//	StartTime: will be time.Now() utc to string with format "2020-06-30 15:07:00 -0700"
		StartTime: time.Now().Format("2006-01-02 15:04:05 -0700"),
		Unit:      "day",
		Duration:  1,
	}

	// create the payment request to midtrans
	paymentRes, err := paymentReq.CreatePayment()
	if err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout9-"+err.Error())
		return
	}

	// commit the transaction
	if err := tx.Commit(); err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "checkout10-"+err.Error())
		return
	}

	// delete the item in the cart
	go func() {
		_, err := database.MysqlInstance.Exec(
			"DELETE FROM cart_items WHERE customer_refer = UUID_TO_BIN(?) AND checked = 1", customerId,
		)
		if err != nil {
			go logging.InsertLog(logging.ERROR, "checkout11-"+err.Error())
		}
		// put updateCartCache after the delete query to make sure the cache is updated after the delete query
		updateCartCache(customerId)
	}()

	// update the orders table and fill the payment_token and payment_redirect_url
	// we run this on another goroutine because the user will be redirected to the payment page and we don't want to wait for this to finish
	go func() {
		_, err := database.MysqlInstance.Exec(
			"UPDATE orders SET payment_token = ?, payment_redirect_url = ? WHERE id = ?",
			paymentRes.Token, paymentRes.RedirectUrl, orderId,
		)
		if err != nil {
			go logging.InsertLog(logging.ERROR, "checkout12-"+err.Error())
		}
	}()

	c.JSON(200, paymentRes)
}

func CancelCheckout(c *gin.Context) {
	var request models.APICommonQueryID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var state string
	err = database.MysqlInstance.
		QueryRow(
			"SELECT COALESCE(transaction_status, '') FROM orders WHERE id = ? AND customer_refer = UUID_TO_BIN(?) AND is_paid = 0 AND is_cancelled = 0",
			request.ID, customerId,
		).
		Scan(&state)
	if err != nil {
		// if there is no row it means the order id is not exist for the current customer
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}
	// if there is nothing in state it means the customer haven't chosen the payment method
	// and the midtrans haven't created the transaction
	if state == "" {
		// only set the is_cancelled to 1
		_, err := database.MysqlInstance.Exec(
			"UPDATE orders SET is_cancelled = 1, transaction_status = 'cancel', status_description = 'customer request for cancel' WHERE id = ? AND customer_refer = UUID_TO_BIN(?)",
			request.ID, customerId,
		)
		if err != nil {
			c.Status(500)
			return
		}
		c.Status(200)
		return
	}
	// suppose the transaction_status already filled and the transaction already created in midtrans, we need to cancel the transaction first in midtrans
	if err := midtrans.DeleteOrder(strconv.Itoa(request.ID)); err != nil {
		go logging.InsertLog(
			logging.ERROR,
			"cancel checkout unable to cancel the transaction in midtrans, order id:"+strconv.Itoa(request.ID)+"err :"+err.Error(),
		)
		c.Status(500)
		return
	}
	// set the transaction status to pending cancel as we need to wait for the midtrans to completely cancel the transaction
	_, err = database.MysqlInstance.
		Exec(
			"UPDATE orders SET transaction_status = 'pending cancel', status_description = 'customer request for cancel' WHERE id = ? AND customer_refer = UUID_TO_BIN(?)",
			request.ID, customerId,
		)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}
