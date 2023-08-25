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

	transparency := services.NewTransparency()

	reader, err := transparency.Process(context.Background(), img)

	if err != nil {
		panic(err)
	}

	// do something with reader
	b, _ := ioutil.ReadAll(reader)
	err = os.WriteFile("/home/luciano/Desktop/rg1.png", b, 0644)

	if err != nil {
		panic(err)
	}
}
