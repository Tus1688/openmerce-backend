package customer

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
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
			QueryRow("SELECT shipping_area_refer FROM customer_addresses WHERE id = UUID_TO_BIN(?) AND customer_refer = UUID_TO_BIN(?)", request.AddressCode, customerId).
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
			QueryRow(`
				select sum(p.weight) as weight, sum((p.length * p.height * p.width)) as volume, sum(c.quantity * p.price) as gross_amount
				from products p, cart_items c
				left join inventories i on c.product_refer = i.product_refer
				where p.id = c.product_refer and c.checked = 1 and c.customer_refer = uuid_to_bin(?) and 
					c.quantity <= i.quantity and p.deleted_at is null
				group by c.customer_refer;
				`, customerId).Scan(&weight, &volume, &itemGrossAmount)
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
			Query(`
				select BIN_TO_UUID(p.id), p.name, p.price, c.quantity, p.description, p.weight
				from products p, cart_items c
				left join inventories i on c.product_refer = i.product_refer
				where p.id = c.product_refer and c.customer_refer = UUID_TO_BIN(?)
				and c.checked = 1 and c.quantity <= i.quantity and p.deleted_at is null`, customerId)
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
		return
	}
	// serialize the freight response into the list of choices
	var freightChoices []models.PreCheckoutFreight
	for _, value := range freightRes.Anteraja {
		freightChoices = append(freightChoices, models.PreCheckoutFreight{
			ProductCode: "anteraja-" + value.ProductCode,
			CourierName: "anteraja",
			ProductName: value.ProductName,
			Etd:         value.Etd,
			Rates:       value.Rates,
		})
	}
	for _, value := range freightRes.Sicepat {
		freightChoices = append(freightChoices, models.PreCheckoutFreight{
			ProductCode: "sicepat-" + value.ProductCode,
			CourierName: "sicepat",
			ProductName: value.ProductName,
			Etd:         value.Etd,
			Rates:       value.Rates,
		})
	}

	var freightCost int
	// check if the freight choice is valid and set the freightcost
	for _, choice := range freightChoices {
		if choice.ProductCode == request.CourierCode {
			freightCost = choice.Rates
			break
		}
	}
	log.Print(freightCost)

	// prepare the transaction
	paymentReq := midtrans.RequestSnap{}
	wg.Add(1)

	//	acquire the customer's details
	go func() {
		defer wg.Done()
		var firstName, lastName, email, phone string
		err = database.MysqlInstance.
			QueryRow("SELECT first_name, last_name, email, phone_number FROM customers WHERE id = UUID_TO_BIN(?)", customerId).
			Scan(&firstName, &lastName, &email, &phone)
		if err != nil {
			errChan <- err
			return
		}
		paymentReq.CustomerDetails = midtrans.CustomerDetails{}
	}()

	// begin to make change into the database and prepare the transaction of the payment gateway
	//
	tx, err := database.MysqlInstance.Begin()
	if err != nil {
		c.Status(500)
		return
	}
	defer tx.Rollback()

	c.Status(200)
}
