S3_BUCKET_NAME = conspirago-documents
LAMBDA_FUNCTION_NAME = ConspiraGo-Lambda
AWS_REGION = us-east-1
GO_BINARY_NAME = mylambda

create-bucket:
	aws s3api create-bucket --bucket $(S3_BUCKET_NAME) --region $(AWS_REGION)

build-lambda:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/Handler ./cmd/analyzer-lambda/main.go
	zip -r $(GO_BINARY_NAME).zip bin/Handler


deploy-lambda: build-lambda
	aws lambda update-function-code --function-name $(LAMBDA_FUNCTION_NAME) \
		--zip-file fileb://$(GO_BINARY_NAME).zip > /dev/null

configure-lambda:
	aws lambda create-event-source-mapping --function-name $(LAMBDA_FUNCTION_NAME) \
		--batch-size 10 --event-source-arn arn:aws:s3:::$(S3_BUCKET_NAME) \
		--starting-position TRIM_HORIZON

.PHONY: create-bucket build-lambda deploy-lambda configure-lambda
