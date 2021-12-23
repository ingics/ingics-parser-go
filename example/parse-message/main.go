package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ingics/ingics-parser-go/ibs"
	"github.com/ingics/ingics-parser-go/igs"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Error: message string is required.")
		fmt.Printf("Usage: %v <message>\n", os.Args[0])
		fmt.Printf("Example: %v \"$GPRP,D8714D784B4F,F008D1789200,-45,02010612FF0D0083BC3101006D0B31000000140F0600,1630382368.698\"\n", os.Args[0])
		os.Exit(1)
	}

	if m := igs.Parse(os.Args[1]); m != nil {
		fmt.Printf("Type:    %v\n", m.MsgType())
		fmt.Printf("Beacon:  %v\n", m.Beacon())
		fmt.Printf("Gateway: %v\n", m.Gateway())
		fmt.Printf("RSSI:    %v\n", m.RSSI())
		fmt.Printf("Payload: %v\n", m.Payload())
		if t := m.Timestamp(); t != nil {
			fmt.Printf("Time:    %v\n", t)
		}

		if bytes, err := hex.DecodeString(m.Payload()); err == nil {
			p := ibs.Parse(bytes)
			fmt.Printf("Payload(parsed): %v\n", p)
		}
	} else {
		fmt.Println("Error: Invalid input message")
		fmt.Println(os.Args[1])
	}
}
