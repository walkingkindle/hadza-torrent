package main

import (
	"bufio"
	"fmt"
	"os"

	"torrent-client-go/magnet-parser"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Gimme magnet \n")

	sentence, err := reader.ReadString('\n')
	if err != nil {
		fmt.Print("Error when reading out the sentence")
		return
	}

	result, err := parser.ParseMagnet(sentence)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n", result)
}
