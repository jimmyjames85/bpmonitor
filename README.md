# bpmonitor


## Developing

	docker-compose up -d setup
	source ./example.env
	go run cmd/bpmonitor/main.go

	# if you want to run a webserver hosting files in web/
	docker-compose up -d webserver
