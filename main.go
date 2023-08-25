package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/RossiEric/golang-aws-dojo-ConspiraGo/internal/services"
)

func main() {
	img, err := os.Open("/home/luciano/Desktop/IMG_20230825_135633.jpg")

	if err != nil {
		panic(err)
	}

	defer img.Close()

	transparency := services.NewImage()

	splited, err := transparency.Slice(context.Background(), img, services.Bound{
		Top:    2606,
		Bottom: 2744,
		Left:   416,
		Right:  1274,
	})

	if err != nil {
		panic(err)
	}

	reader, err := transparency.RemoveTransparency(context.Background(), splited)

	if err != nil {
		panic(err)
	}

	b, _ := ioutil.ReadAll(reader)
	err = os.WriteFile("/home/luciano/Desktop/rg1.png", b, 0644)

	if err != nil {
		panic(err)
	}
}
