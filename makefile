S3_BUCKET_NAME = conspirago-documents
LAMBDA_FUNCTION_NAME = ConspiraGo-Lambda
AWS_REGION = us-east-1
GO_BINARY_NAME = mylambda

create-bucket:
	aws s3api create-bucket --bucket $(S3_BUCKET_NAME) --region $(AWS_REGION)

build-lambda:
	GOOS=linux GOARCH=amd64 go build -o $(GO_BINARY_NAME) main.go

deploy-lambda: build-lambda
	aws lambda create-function --function-name $(LAMBDA_FUNCTION_NAME) \
		--runtime go1.x --role [ARN_DO_SEU_ROLE] \
		--handler $(GO_BINARY_NAME) \
		--zip-file fileb://$(GO_BINARY_NAME) --region $(AWS_REGION)

configure-lambda:
	aws lambda create-event-source-mapping --function-name $(LAMBDA_FUNCTION_NAME) \
		--batch-size 10 --event-source-arn arn:aws:s3:::$(S3_BUCKET_NAME) \
		--starting-position TRIM_HORIZON

.PHONY: create-bucket build-lambda deploy-lambda configure-lambda
