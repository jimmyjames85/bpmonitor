# bpmonitor


## Developing

Setup:

	docker-compose up -d setup webserver

Run locally with

	source ./example.env && go run cmd/bpmonitor/main.go

Login with username `bp` and password `monitor` via:

	http://localhost:8080/
