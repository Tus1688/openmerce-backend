package staff

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/Tus1688/openmerce-backend/database"
	"github.com/Tus1688/openmerce-backend/models"
	"github.com/gin-gonic/gin"
)

func AddHomeBanner(c *gin.Context) {
	var request models.InsertBanner
	if err := c.ShouldBind(&request); err != nil {
		c.Status(400)
		return
	}
	url := NginxFSBaseUrl + "/handler"
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
	if _, err := io.Copy(part, image); err != nil {
		c.Status(500)
		return
	}
	if err := writer.Close(); err != nil {
		c.Status(500)
		return
	}
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		c.Status(500)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", NginxFSAuthorization)
	res, err := http.DefaultClient.Do(req)
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
	// insert to database
	_, err = database.MysqlInstance.Exec("INSERT INTO homepage_banner (file_name, href) VALUES (?, ?)", response.File, request.Href)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(201)
}

func DeleteHomeBanner(c *gin.Context) {
	var request models.APICommonQueryID
	if err := c.ShouldBind(&request); err != nil {
		c.Status(400)
		return
	}
	// get file name from database
	var fileName string
	err := database.MysqlInstance.
		QueryRow("SELECT file_name FROM homepage_banner WHERE id = ?", request.ID).
		Scan(&fileName)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Status(404)
			return
		}
		c.Status(500)
		return
	}

	// delete from the image server
	url := NginxFSBaseUrl + "/handler?file=" + fileName
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		c.Status(500)
		return
	}
	req.Header.Set("Authorization", NginxFSAuthorization)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Status(500)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 && res.StatusCode != 404 {
		// 404 considered as success because the file is already deleted
		c.Status(500)
		return
	}
	_, err = database.MysqlInstance.
		Exec("DELETE FROM homepage_banner WHERE id = ?", request.ID)
	if err != nil {
		c.Status(500)
		return
	}
	c.Status(200)
}
