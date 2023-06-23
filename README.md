## Openmerce's Backend
This is the backend of Openmerce, an open source e-commerce platform for Indonesia. This project is built using 
- Mysql
- Redis
- Gin Gonic

## How to deploy this project
Refer to this [Repository](https://github.com/Tus1688/openmerce-deployment) for docker swarm deployment guide

## Development
### Prerequisites
- Docker compose 
- Go 1.20 or higher
- Freight service (we use private freight service, so you need to build it yourself)
- Midtrans account (for payment gateway)
- Mailgun account (for email service)

### How to run
1. Clone this repository
2. Run `docker-compose up -d` to start mysql, redis, go-nginx-fs, and (your own freight service, so make sure to build it first)
3. Run `go run main.go` to start the server

### Note
- [Go-nginx-fs](https://github.com/Tus1688/go-nginx-fs) (for image server)
- Create your own freight service
- Create your own .env file (refer to .env.example)

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details