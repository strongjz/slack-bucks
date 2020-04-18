OUTPUT = golang_app_serverless # Referenced as Handler in template.yaml
TEMPLATE = template.yaml
S3_BUCKET = golang-example-serverless-app
AWS_PROFILE ?= strongjz-tech
VERSION ?= 0.0.4

.PHONY: vendor
vendor:
	dep ensure

.PHONY: test
test:
	go test ./...

.PHONY: clean
clean:
	rm -f $(OUTPUT)

install:
	go get .

build: install
	GOOS=linux GOARCH=amd64 go build -o $(OUTPUT) main.go

# compile the code to run in Lambda (local or real)
.PHONY: lambda
lambda: build

zip: build
	zip $(VERSION)-app.zip $(OUTPUT)

upload: zip
	echo "Uploading buck to ${S3_BUCKET}"
	aws s3 cp $(VERSION)-app.zip s3://$(S3_BUCKET)/$(VERSION)/app.zip

.PHONY: api
api-debug: build
	sam local start-api --debug --env-vars env.json --profile $(AWS_PROFILE)

.PHONY: api
api: build
	sam local start-api --env-vars env.json --profile $(AWS_PROFILE)
