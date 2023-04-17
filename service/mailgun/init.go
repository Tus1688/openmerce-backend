package mailgun

import "os"

var creds mailgun

func ReadEnv() {
	creds = mailgun{
		APIKey: os.Getenv("MAILGUN_API_KEY"),
		Domain: os.Getenv("MAILGUN_DOMAIN"),
	}
}
