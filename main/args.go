package main

import (
	"errors"
	"flag"
)

func ParseLocationFromArgs() (result string, err error) {
	filePath := flag.String("file", "", "Specify the file location")

	flag.Parse()

	if *filePath == "" {
		return "", errors.New("Could not parse the location. Please specify the location of the torrent file using the --file flag")
	}

	parsedLocation := *filePath

	return parsedLocation, nil
}
