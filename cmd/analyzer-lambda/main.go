package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func handleRequest(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3Entity := record.S3

		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			panic(fmt.Sprintf("failed to load SDK config, %v", err))
		}

		client := s3.NewFromConfig(cfg)
		getObjResp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: &s3Entity.Bucket.Name,
			Key:    &s3Entity.Object.Key,
		})

		if err != nil {
			panic(fmt.Sprintf("failed to get s3 object, %v", err))
		}

		fmt.Println(getObjResp.ETag)

		// TODO: textract stuff
	}
}

func main() {
	lambda.Start(handleRequest)
}
