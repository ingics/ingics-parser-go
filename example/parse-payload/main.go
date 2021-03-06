package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ingics/ingics-parser-go/ibs"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Error: Payload hex string is required.")
		fmt.Printf("Usage: %v <payload1> [payload2] ...\n", os.Args[0])
		fmt.Printf("Example: %v 02010612FF0D0083BC280100770B3000000014010000\n", os.Args[0])
		os.Exit(1)
	}

	for _, payloadHex := range os.Args[1:] {
		if payloadBytes, err := hex.DecodeString(payloadHex); err == nil {
			payload := ibs.Parse(payloadBytes)
			fmt.Println(payload)
		} else {
			fmt.Printf("Invalid hex string: %v", payloadHex)
			fmt.Println(err)
		}
	}
}
