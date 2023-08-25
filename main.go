package main

import (
	"conspirago/golang-aws-dojo-ConspiraGo/internal/services"
	"context"
	"io/ioutil"
	"os"
)

func main() {
	img, err := os.Open("/home/luciano/Desktop/rg1.jpg")

	if err != nil {
		panic(err)
	}

	defer img.Close()

	transparency := services.NewImage()

	reader, err := transparency.RemoveTransparency(context.Background(), img)

	if err != nil {
		panic(err)
	}

	splited, err := transparency.Slice(context.Background(), reader, 0, 44, 0, 126)

	if err != nil {
		panic(err)
	}

	// do something with reader
	b, _ := ioutil.ReadAll(splited)
	err = os.WriteFile("/home/luciano/Desktop/rg1.png", b, 0644)

	if err != nil {
		panic(err)
	}
}
