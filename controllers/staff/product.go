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

package staff

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/logging"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AddNewProduct(c *gin.Context) {
	var request models.ProductCreate
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	wg := sync.WaitGroup{}
	errChan := make(chan error, 1)
	// check if the category exists
	wg.Add(1)
	go func(id uint) {
		defer wg.Done()
		var exist int8
		err := database.MysqlInstance.QueryRow(
			"SELECT 1 FROM categories WHERE id = ? AND deleted_at IS NULL", id,
		).Scan(&exist)
		if exist != 1 {
			errChan <- errors.New("category not found")
			return
		}
		// the sql.ErrNoRows won't ever be called as the category exist check is already done
		if err != nil {
			errChan <- err
		}
	}(request.CategoryID)

	var id uuid.UUID
	// check if soft-delete product with the same name exists
	var existingProductID uuid.UUID
	err := database.MysqlInstance.
		QueryRow("SELECT BIN_TO_UUID(id) FROM products WHERE name = ? AND deleted_at IS NOT NULL", request.Name).
		Scan(&existingProductID)
	if err != nil && err != sql.ErrNoRows {
		go logging.InsertLog(logging.ERROR, "1-exist:"+err.Error())
		c.Status(500)
		return
	}

	// wait for the category check to finish
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			if strings.Contains(err.Error(), "category") {
				c.JSON(409, gin.H{"error": err.Error()})
			} else {
				c.Status(500)
				go logging.InsertLog(logging.ERROR, "2-cat:"+err.Error())
			}
			return
		}
	}

	if existingProductID != uuid.Nil {
		//	update the deleted_at to NULL
		_, err = database.MysqlInstance.
			Exec(
				"UPDATE products SET deleted_at = NULL, created_at = CURRENT_TIMESTAMP, updated_at = NULL, description = ?, price = ?, weight = ?, category_refer = ?, cumulative_review = 0, length = ?, width = ?, height = ?  WHERE id = UUID_TO_BIN(?)",
				request.Description, request.Price, request.Weight, request.CategoryID, request.Length, request.Width,
				request.Height, existingProductID,
			)
		if err != nil {
			c.Status(500)
			go logging.InsertLog(logging.ERROR, "2-cat:"+err.Error())
			return
		}
		id = existingProductID
	} else {
		id = uuid.New()
		// insert the new product
		_, err = database.MysqlInstance.
			Exec(
				"INSERT INTO products (id, name, description, price, weight, category_refer, length, width, height) VALUES (UUID_TO_BIN(?), ?, ?, ?, ?, ?, ?, ?, ?)",
				id, request.Name, request.Description, request.Price, request.Weight, request.CategoryID,
				request.Length, request.Width, request.Height,
			)
		if err != nil {
			//	check if the product name already exists
			if strings.Contains(err.Error(), "Duplicate entry") {
				c.JSON(409, gin.H{"error": "Product name already exists"})
				return
			}
			go logging.InsertLog(logging.ERROR, "3-insert:"+err.Error())
			c.Status(500)
			return
		}
	}
	// insert the new product into inventories
	if existingProductID != uuid.Nil {
		_, err = database.MysqlInstance.
			Exec(
				"UPDATE inventories SET quantity = ?, updated_at = CURRENT_TIMESTAMP WHERE product_refer = UUID_TO_BIN(?)",
				request.InitialStock, existingProductID,
			)
	} else {
		_, err = database.MysqlInstance.
			Exec(
				"INSERT INTO inventories (product_refer, quantity, updated_at) VALUE (UUID_TO_BIN(?), ?, CURRENT_TIMESTAMP)",
				id, request.InitialStock,
			)
	}
	if err != nil {
		go logging.InsertLog(logging.ERROR, "4-inventory:"+err.Error())
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"id": id})
}

func AddImage(c *gin.Context) {
	var request models.ProductImage
	if err := c.ShouldBind(&request); err != nil {
		c.Status(400)
		return
	}
	// check if the product exists
	var exist int8
	err := database.MysqlInstance.QueryRow(
		"SELECT 1 FROM products WHERE id = UUID_TO_BIN(?)", request.ProductID,
	).Scan(&exist)
	if err != nil {
		c.Status(404)
		return
	}
	//	Insert the image by sending request into NginxFSBaseUrl
	url := NginxFSBaseUrl + "/handler"
	//	upload the image to NginxFS
	image, err := request.Picture.Open()
	if err != nil {
		c.Status(500)
		go logging.InsertLog(logging.ERROR, "1-addimg:"+err.Error())
		return
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("picture", request.Picture.Filename)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "2-addimg:"+err.Error())
		c.Status(500)
		return
	}
	_, err = io.Copy(part, image)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "3-addimg:"+err.Error())
		c.Status(500)
		return
	}
	err = writer.Close()
	if err != nil {
		go logging.InsertLog(logging.ERROR, "4-addimg:"+err.Error())
		c.Status(500)
		return
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "5-addimg:"+err.Error())
		c.Status(500)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", NginxFSAuthorization)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "6-addimg:"+err.Error())
		c.Status(500)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 201 {
		go logging.InsertLog(logging.ERROR, "7-addimg:"+err.Error())
		c.Status(500)
		return
	}
	// we are going to get "id": uuid from the response
	var response struct {
		File string `json:"file"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		go logging.InsertLog(logging.ERROR, "8-addimg:"+err.Error())
		c.Status(500)
		return
	}
	//	insert the response.File into product_images
	_, err = database.MysqlInstance.
		Exec(
			"INSERT INTO product_images (id, product_refer) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?))",
			strings.Replace(response.File, ".webp", "", 1), request.ProductID,
		)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "9-addimg:"+err.Error())
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"file": response.File})
}

func DeleteProduct(c *gin.Context) {
	var request models.APICommonQueryUUID
	if err := c.ShouldBindQuery(&request); err != nil {
		c.Status(400)
		return
	}
	//	try to delete the product by set deleted_at to current timestamp
	res, err := database.MysqlInstance.Exec(
		"UPDATE products SET deleted_at = CURRENT_TIMESTAMP WHERE id = UUID_TO_BIN(?) AND deleted_at IS NULL",
		request.ID,
	)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "1-delprod"+err.Error())
		c.Status(500)
		return
	}
	//	check if the product exists
	affected, err := res.RowsAffected()
	if err != nil {
		go logging.InsertLog(logging.ERROR, "2-delprod"+err.Error())
		c.Status(500)
		return
	}
	if affected == 0 {
		c.Status(404)
		return
	}
	//	try to delete product_images
	var imageUrls []string
	rows, err := database.MysqlInstance.Query(
		"SELECT BIN_TO_UUID(id) FROM product_images WHERE product_refer = UUID_TO_BIN(?)", request.ID,
	)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "3-delprod"+err.Error())
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var imageUrl string
		if err := rows.Scan(&imageUrl); err != nil {
			c.Status(500)
			go logging.InsertLog(logging.ERROR, "4-delprod"+err.Error())
			return
		}
		imageUrls = append(imageUrls, imageUrl)
	}

	//	delete the images from NginxFS
	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, imageUrl := range imageUrls {
		wg.Add(1)
		go func(targetUrl string) {
			defer wg.Done()
			url := NginxFSBaseUrl + "/handler?file=" + targetUrl + ".webp"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			req.Header.Set("Authorization", NginxFSAuthorization)
			if err != nil {
				errChan <- err
				return
			}
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer res.Body.Close()
			if res.StatusCode != 200 && res.StatusCode != 404 {
				// 404 considered as success as it maybe deleted by other request
				errChan <- errors.New("failed to delete image from NginxFS with status code " + strconv.Itoa(res.StatusCode))
				return
			}
			//	delete the image from product_images
			_, err = database.MysqlInstance.Exec("DELETE FROM product_images WHERE id = UUID_TO_BIN(?)", targetUrl)
			if err != nil {
				errChan <- err
				return
			}
		}(imageUrl)
	}
	// Delete the product from cart_items
	wg.Add(1)
	go func(ProductId string) {
		defer wg.Done()
		_, err := database.MysqlInstance.Exec("DELETE FROM cart_items WHERE product_refer = UUID_TO_BIN(?)", ProductId)
		if err != nil {
			errChan <- err
			return
		}
	}(request.ID)

	// Delete the product from wishlists
	wg.Add(1)
	go func(ProductId string) {
		defer wg.Done()
		_, err := database.MysqlInstance.Exec("DELETE FROM wishlists WHERE product_refer = UUID_TO_BIN(?)", ProductId)
		if err != nil {
			errChan <- err
			return
		}
	}(request.ID)
	//TODO: delete product reviews

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			c.Status(500)
			go logging.InsertLog(logging.ERROR, "5-delprod"+err.Error())
			return
		}
	}
	go ClearFreightCache(request.ID)
	c.Status(200)
}

func DeleteImage(c *gin.Context) {
	var request models.ProductImageDelete
	if err := c.ShouldBind(&request); err != nil {
		c.Status(400)
		return
	}
	//	check if the product exists
	var exist int8
	// strips FileName replace ".webp" with "" to get the id
	err := database.MysqlInstance.
		QueryRow(
			"SELECT 1 FROM product_images WHERE id = UUID_TO_BIN(?) AND product_refer = UUID_TO_BIN(?)",
			strings.Replace(request.FileName, ".webp", "", 1), request.ProductID,
		).Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		go logging.InsertLog(logging.ERROR, "1-delimg"+err.Error())
		c.Status(500)
		return
	}
	if exist != 1 {
		c.Status(404)
		return
	}
	url := NginxFSBaseUrl + "/handler?file=" + request.FileName
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("Authorization", NginxFSAuthorization)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "2-delimg"+err.Error())
		c.Status(500)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "3-delimg"+err.Error())
		c.Status(500)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 && res.StatusCode != 404 {
		go logging.InsertLog(logging.ERROR, "4-delimg"+err.Error())
		// 404 considered as success as it maybe deleted by other request
		c.Status(500)
		return
	}
	//	delete the image from product_images
	_, err = database.MysqlInstance.
		Exec(
			"DELETE FROM product_images WHERE id = UUID_TO_BIN(?) AND product_refer = UUID_TO_BIN(?)",
			strings.Replace(request.FileName, ".webp", "", 1), request.ProductID,
		)
	if err != nil {
		go logging.InsertLog(logging.ERROR, "5-delimg"+err.Error())
		c.Status(500)
		return
	}
	c.Status(200)
}

func UpdateProduct(c *gin.Context) {
	var request models.ProductUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		c.Status(400)
		return
	}
	// somethingToUpdate is used to check if there is any field to update
	// if there is no field to update, it will return 400
	var somethingToUpdate bool
	query := "UPDATE products p"
	var args []interface{}
	if request.Stock != 0 {
		query += ", inventories i SET i.quantity = ?, i.updated_at = CURRENT_TIMESTAMP, "
		args = append(args, request.Stock)
		somethingToUpdate = true
	} else {
		query += " SET "
	}
	query += "p.updated_at = CURRENT_TIMESTAMP"
	wg := sync.WaitGroup{}
	errChan := make(chan error, 2)
	if request.Name != "" {
		wg.Add(1)
		// this goroutine is used to check if the product name is already exist
		// in deleted product, if it existed, it will update the deleted product name
		// into LEFT(name, 60) + "_deleted_" + current datetime (YYYY-MM-DD HH:MM)
		go func(name string) {
			defer wg.Done()
			var deletedName string
			_ = database.MysqlInstance.QueryRow(
				"SELECT LEFT(name, 60) FROM products WHERE name = ? AND deleted_at IS NOT NULL", name,
			).Scan(&deletedName)
			if deletedName != "" {
				//	update the deleted product name into deletedName + "_deleted" + current datetime (YYYY-MM-DD HH:MM)
				_, err := database.MysqlInstance.
					Exec(
						"UPDATE products SET name = CONCAT(?, '_deleted_', DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i')) WHERE name = ? AND deleted_at IS NOT NULL",
						deletedName, name,
					)
				if err != nil {
					errChan <- err
					return
				}
			}
		}(request.Name)
		query += ", p.name = ?"
		args = append(args, request.Name)
		somethingToUpdate = true
	}
	if request.Description != "" {
		query += ", p.description = ?"
		args = append(args, request.Description)
		somethingToUpdate = true
	}
	if request.Price != 0 {
		query += ", p.price = ?"
		args = append(args, request.Price)
		somethingToUpdate = true
	}
	if request.CategoryID != 0 {
		wg.Add(1)
		// this goroutine check if the category exist
		go func(id uint) {
			defer wg.Done()
			var exist int8
			err := database.MysqlInstance.QueryRow(
				"SELECT 1 FROM categories WHERE id = ? AND deleted_at IS NULL", id,
			).Scan(&exist)
			if exist != 1 {
				errChan <- errors.New("category not found")
				return
			}
			// the sql.ErrNoRows won't ever be called as the category exist check is already done
			if err != nil {
				errChan <- err
			}
		}(request.CategoryID)
		query += ", p.category_refer = ?"
		args = append(args, request.CategoryID)
		somethingToUpdate = true
	}
	if request.Weight != 0 {
		query += ", p.weight = ? "
		args = append(args, request.Weight)
		somethingToUpdate = true
	}
	if request.Length != 0 {
		query += ", p.length = ? "
		args = append(args, request.Length)
		somethingToUpdate = true
	}
	if request.Width != 0 {
		query += ", p.width = ? "
		args = append(args, request.Width)
		somethingToUpdate = true
	}
	if request.Height != 0 {
		query += ", p.height = ? "
		args = append(args, request.Height)
		somethingToUpdate = true
	}
	if request.Stock != 0 {
		query += " WHERE i.product_refer = UUID_TO_BIN(?) AND "
		args = append(args, request.ID)
		somethingToUpdate = true
	} else {
		query += " WHERE "
	}
	if !somethingToUpdate {
		c.Status(400)
		return
	}
	query += "p.id = UUID_TO_BIN(?) AND p.deleted_at IS NULL"
	args = append(args, request.ID)
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			if strings.Contains(err.Error(), "category") {
				c.JSON(409, gin.H{"error": err.Error()})
			} else {
				c.Status(500)
			}
			return
		}
	}
	res, err := database.MysqlInstance.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			c.JSON(409, gin.H{"error": "Product name already exist"})
			return
		}
		c.Status(500)
		return
	}
	affected, err := res.RowsAffected()
	if err != nil {
		c.Status(500)
		return
	}
	if affected == 0 {
		// if nothing updated and there is multiple update request at the same time, it will return 404 even though the
		// product exist as there is nothing to update and CURRENT_TIMESTAMP increment by 1 second
		c.Status(404)
		return
	}
	go ClearFreightCache(request.ID)
	c.Status(200)
}

// ClearFreightCache is used to clear the cache on redisInstance[5] which match the productID*
func ClearFreightCache(productID string) {
	ctx := context.Background()
	// clear the cache on redisInstance[5] which match the productID*
	iter := database.RedisInstance[5].Scan(ctx, 0, productID+"*", 0).Iterator()
	for iter.Next(ctx) {
		err := database.RedisInstance[5].Del(ctx, iter.Val()).Err()
		if err != nil {
			log.Print(err)
		}
	}
	if err := iter.Err(); err != nil {
		log.Print(err)
	}
}

func GetProduct(c *gin.Context) {
	var response []models.HomepageProduct
	rows, err := database.MysqlInstance.
		Query(
			`
			SELECT
			    BIN_TO_UUID(p.id) AS id,
			    p.name,
			    p.price,
			    COALESCE(CONCAT(BIN_TO_UUID(pi.id), '.webp'), '') AS image,
			    p.cumulative_review,
			    COUNT(oi.id) AS sold_count
			FROM
			    products p
			        LEFT JOIN (
			        SELECT
			            id,
			            product_refer,
			            ROW_NUMBER() OVER (PARTITION BY product_refer ORDER BY created_at) AS rn
			        FROM
			            product_images
			    ) pi ON p.id = pi.product_refer AND pi.rn = 1
			        LEFT JOIN
			    order_items oi ON oi.product_refer = p.id
			        LEFT JOIN
			    orders o ON oi.order_refer = o.id
			        AND (o.transaction_status = 'settlement'
			            OR o.transaction_status = 'capture')
			WHERE
			    p.deleted_at IS NULL
			GROUP BY
			    p.id,
			    image`,
		)
	if err != nil {
		c.Status(500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var product models.HomepageProduct
		err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.ImageUrl, &product.Rating, &product.Sold)
		if err != nil {
			c.Status(500)
			return
		}
		response = append(response, product)
	}
	c.JSON(200, response)
}
