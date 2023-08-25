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

	sliced, err := transparency.Slice(context.Background(), img, services.Bound{
		Left:   416,
		Right:  1272,
		Top:    2607,
		Bottom: 2745,
	})

	if err != nil {
		panic(err)
	}

	transparent, err := transparency.RemoveTransparency(context.Background(), sliced)

	if err != nil {
		panic(err)
	}

	// do something with reader
	b, _ := ioutil.ReadAll(transparent)
	err = os.WriteFile("/home/luciano/Desktop/teste1.png", b, 0644)

	if err != nil {
		panic(err)
	}
}
