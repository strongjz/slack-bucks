OUTPUT = main # Referenced as Handler in template.yaml
TEMPLATE = template.yaml

.PHONY: vendor
vendor:
	dep ensure

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f $(OUTPUT) $(PACKAGED_TEMPLATE)

.PHONY: install
install:
	go get ./...

main: ./main.go
	go build -o $(OUTPUT) main.go

# compile the code to run in Lambda (local or real)
.PHONY: lambda
lambda:
	GOOS=linux GOARCH=amd64 $(MAKE) main

.PHONY: build
build: clean lambda


.PHONY: api
api-debug: build
	sam local start-api --debug --env-vars env.json

.PHONY: api
api: build
	sam local start-api --env-vars env.json
