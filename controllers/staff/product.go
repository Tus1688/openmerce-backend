package staff

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/Tus1688/openmerce-backend/database"
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
	id := uuid.New()
	// insert the new product
	_, err := database.MysqlInstance.
		Exec("INSERT INTO products (id, name, description, price, weight, category_refer) VALUES (UUID_TO_BIN(?), ?, ?, ?, ?, ?)",
			id, request.Name, request.Description, request.Price, request.Weight, request.CategoryID)
	if err != nil {
		//	check if the product name already exists
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(409, gin.H{"error": "Product name already exists"})
			return
		}
		c.Status(500)
		return
	}
	// insert the new product into inventories
	_, err = database.MysqlInstance.
		Exec("INSERT INTO inventories (product_refer, quantity, updated_at) VALUE (UUID_TO_BIN(?), ?, CURRENT_TIMESTAMP)",
			id, request.InitialStock)
	if err != nil {
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
	err := database.MysqlInstance.QueryRow("SELECT 1 FROM products WHERE id = UUID_TO_BIN(?)", request.ProductID).Scan(&exist)
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
		return
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("picture", request.Picture.Filename)
	if err != nil {
		c.Status(500)
		return
	}
	_, err = io.Copy(part, image)
	if err != nil {
		c.Status(500)
		return
	}
	err = writer.Close()
	if err != nil {
		c.Status(500)
		return
	}
	res, err := http.Post(url, writer.FormDataContentType(), body)
	if err != nil {
		c.Status(500)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 201 {
		c.Status(500)
		return
	}
	// we are going to get "id": uuid from the response
	var response struct {
		File string `json:"file"`
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		c.Status(500)
		return
	}
	//	insert the response.File into product_images
	_, err = database.MysqlInstance.
		Exec("INSERT INTO product_images (id, product_refer) VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?))",
			strings.Replace(response.File, ".webp", "", 1), request.ProductID)
	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"file": response.File})
}
