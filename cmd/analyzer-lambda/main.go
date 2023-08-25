package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/RossiEric/golang-aws-dojo-ConspiraGo/internal/services"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/textract"
)

var textractSession *textract.Textract
var s3Client *s3.S3

func init() {
	textractSession = textract.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})))

	s3Client = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})))
}

func main() {
	//lambda.Start(Handler)
	process(context.Background(), "conspirago-documents", "rg.jpg")
}

func process(ctx context.Context, bucket string, key string) {
	input := &textract.AnalyzeDocumentInput{
		Document: &textract.Document{
			S3Object: &textract.S3Object{
				Bucket: &bucket,
				Name:   &key,
			},
		},
		FeatureTypes: []*string{
			aws.String("SIGNATURES"),
		},
	}

	result, err := textractSession.AnalyzeDocument(input)

	if err != nil {
		panic(err)
	}

	file, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		panic(err)
	}

	im, _, err := image.Decode(file.Body)

	if err != nil {
		panic(err)
	}

	imageWidth := im.Bounds().Max.X
	imageHeight := im.Bounds().Max.Y

	service := services.NewImage()

	for _, block := range result.Blocks {
		if *block.BlockType == "SIGNATURE" {
			x := int(*block.Geometry.BoundingBox.Left * float64(imageWidth))
			y := int(*block.Geometry.BoundingBox.Top * float64(imageHeight))
			width := int(*block.Geometry.BoundingBox.Width * float64(imageWidth))
			height := int(*block.Geometry.BoundingBox.Height * float64(imageHeight))

			var buffer bytes.Buffer

			err = png.Encode(&buffer, im)

			if err != nil {
				panic(err)
			}

			sliced, err := service.Slice(ctx, bytes.NewReader(buffer.Bytes()), services.Bound{
				Top:    y,
				Left:   x,
				Right:  width + x,
				Bottom: height + y,
			})

			if err != nil {
				panic(err)
			}

			transparent, err := service.RemoveTransparency(ctx, sliced)

			if err != nil {
				panic(err)
			}

			b, _ := ioutil.ReadAll(transparent)

			err = os.WriteFile("output.png", b, 0644)

			if err != nil {
				panic(err)
			}

			fmt.Println("Signature cropped and saved!")
		}
	}
}

func Handler(ctx context.Context, s3Event events.S3Event) {
	for _, record := range s3Event.Records {
		s3Entity := record.S3

		process(ctx, s3Entity.Bucket.Name, s3Entity.Object.Key)
	}
}
