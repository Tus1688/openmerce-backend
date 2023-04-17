package mailgun

import (
	"net/http"
	"net/url"
	"strings"
)

type mailgun struct {
	APIKey string
	Domain string
}

type MailgunSend struct {
	FromName    string
	FromAddress string
	To          string
	Subject     string
	Body        string
}

func SendEmail(send MailgunSend) error {
	baseUrl := "https://api.mailgun.net/v3/" + creds.Domain + "/messages"
	data := url.Values{
		"from":    {send.FromName + " <" + send.FromAddress + "@" + creds.Domain + ">"},
		"to":      {send.To},
		"subject": {send.Subject},
		"text":    {send.Body},
	}
	req, err := http.NewRequest("POST", baseUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", creds.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		return err
	}
	return nil
}
