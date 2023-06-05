package customer

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/Tus1688/openmerce-backend/auth"
	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func GetOrder(c *gin.Context) {
	// the token should be valid and exist as it is protected by TokenExpiredCustomer middleware
	token, _ := c.Cookie("ac_cus")
	claims, err := auth.ExtractClaimAccessTokenCustomer(token)
	if err != nil {
		c.Status(401)
		return
	}
	customerId := claims.Uid
	var request models.APICommonQueryID
	if err := c.ShouldBindQuery(&request); err == nil {
		//	get the specific order detail
		var response models.OrderDetailResponse
		wg := sync.WaitGroup{}
		mu := sync.Mutex{}
		errChan := make(chan error)
		wg.Add(3)

		// get the order details
		go func() {
			defer wg.Done()
			var id uint64
			var status, statusDescription, createdAt, courier, TrackingCode string
			var itemCost, shippingCost, totalCost uint
			err := database.MysqlInstance.
				QueryRow(`SELECT id, coalesce(transaction_status, ''), coalesce(status_description, ''), DATE_FORMAT(created_at, '%d %M %Y'),
       			courier_code,  coalesce(courier_tracking_code, ''), item_cost, freight_cost, gross_amount
       			FROM orders WHERE customer_refer = UUID_TO_BIN(?) AND id = ?`, customerId, request.ID).
				Scan(&id, &status, &statusDescription, &createdAt, &courier, &TrackingCode, &itemCost, &shippingCost, &totalCost)
			if err != nil {
				if err == sql.ErrNoRows {
					errChan <- fmt.Errorf("order not found")
				}
				errChan <- err
				return
			}
			mu.Lock()
			response.ID = id
			response.Status = status
			response.StatusDescription = statusDescription
			response.CreatedAt = createdAt
			response.Courier = courier
			response.TrackingCode = TrackingCode
			response.ItemCost = itemCost
			response.ShippingCost = shippingCost
			response.TotalCost = totalCost
			mu.Unlock()
		}()

		// get the items
		go func() {
			defer wg.Done()
			var itemList []models.ItemListOrderDetail
			rows, err := database.MysqlInstance.
				Query(`
					select oi.id, BIN_TO_UUID(oi.product_refer), oi.on_buy_name, oi.on_buy_price, coalesce(pi.image, ''), oi.quantity, (if (r.id is null, false, true)) as reviewed
					from order_items oi
					         left join (select pi.product_refer, CONCAT(BIN_TO_UUID(pi.id), '.webp') as image
					                    from (select product_refer, min(id) as id
					                          from product_images
					                          group by product_refer) pi) pi on pi.product_refer = oi.product_refer
					         inner join orders o on oi.order_refer = o.id
					left join reviews r on oi.id = r.order_item_refer
					where oi.order_refer = ?
					  and o.customer_refer = UUID_TO_BIN(?);
				`, request.ID, customerId)
			if err != nil {
				errChan <- err
				return
			}
			defer rows.Close()
			for rows.Next() {
				var item models.ItemListOrderDetail
				err := rows.Scan(&item.OrderID, &item.ProductId, &item.ProductName, &item.ProductPrice, &item.ProductImage, &item.Quantity, &item.Reviewed)
				if err != nil {
					errChan <- err
					return
				}
				itemList = append(itemList, item)
			}
			mu.Lock()
			response.ItemList = itemList
			mu.Unlock()
		}()

		// get the address details
		go func() {
			defer wg.Done()
			var address models.AddressOrderResponse
			err := database.MysqlInstance.
				QueryRow(`
					select ca.recipient_name, ca.phone_number, ca.full_address, sa.full_name
					from orders o
					         left join customer_addresses ca on o.customer_address_refer = ca.id
					         left join shipping_areas sa on ca.shipping_area_refer = sa.id
					where o.id = ?
					  and o.customer_refer = UUID_TO_BIN(?);
				`, request.ID, customerId).
				Scan(&address.RecipientName, &address.PhoneNumber, &address.FullAddress, &address.ShippingArea)
			if err != nil {
				if err == sql.ErrNoRows {
					errChan <- fmt.Errorf("order not found")
					return
				}
				errChan <- err
				return
			}
			mu.Lock()
			response.AddressDetail = address
			mu.Unlock()
		}()

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				if err.Error() == "order not found" {
					c.Status(404)
				} else {
					c.Status(500)
				}
				return
			}
		}
		c.JSON(200, response)
		return
	}

	// otherwise get the all order list
	var response []models.OrderResponse
	// get the order list from database
	rows, err := database.MysqlInstance.
		Query(`
			SELECT o.id,
			       DATE_FORMAT(o.created_at, '%d %M %Y') AS created_at,
			       o.gross_amount,
			       COALESCE(o.transaction_status, ''),
			       oi.item_count,
			       COALESCE(pi.image, ''),
			       p.name
			FROM orders o
			         LEFT JOIN (
			    SELECT order_refer, COUNT(*) AS item_count
			    FROM order_items
			    GROUP BY order_refer
			) oi ON oi.order_refer = o.id
			         LEFT JOIN (
			    SELECT pi.product_refer, CONCAT(BIN_TO_UUID(pi.id), '.webp') AS image
			    FROM (
			             SELECT product_refer, MIN(id) AS id
			             FROM product_images
			             GROUP BY product_refer
			         ) pi
			) pi ON pi.product_refer = (SELECT product_refer FROM order_items WHERE order_refer = o.id LIMIT 1)
			         LEFT JOIN (
			    SELECT id, name
			    FROM products
			) p ON p.id = (SELECT product_refer FROM order_items WHERE order_refer = o.id LIMIT 1)
			WHERE o.customer_refer = UUID_TO_BIN(?);
			`, customerId)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item models.OrderResponse
		err := rows.Scan(&item.ID, &item.CreatedAt, &item.GrossAmount, &item.Status, &item.ItemCount, &item.Image, &item.ProductName)
		if err != nil {
			c.Status(500)
			return
		}
		response = append(response, item)
	}
	if len(response) == 0 {
		c.Status(404)
		return
	}
	c.JSON(200, response)
}
