build: swag
	go build -tags "doc"
build_prod:
	go build
swag:
	swag init  --markdownFiles .
run: build
	./swaggo-gin