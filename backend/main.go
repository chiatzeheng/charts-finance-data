// API KEY = PKDLZ6ZT0R0KSFX9OIDZ
// SECRET = y9WUIUwRWc6AUW51UZdDNZEX4o0DYh58zvxU1Lm8

package main

import (
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/v2/alpaca"
)


type Request struct {
	i int
}


func main() {
	apiKey := "PKDLZ6ZT0R0KSFX9OIDZ"
	apiSecret := "y9WUIUwRWc6AUW51UZdDNZEX4o0DYh58zvxU1Lm8"
	baseURL := "https://paper-api.alpaca.markets"

	// Instantiating new Alpaca paper trading client
	client := alpaca.NewClient(alpaca.ClientOpts{
		// Alternatively, you can set your API key and secret using environment
		// variables named APCA_API_KEY_ID and APCA_API_SECRET_KEY respectively
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
		BaseURL:   baseURL, // Remove for live trading
	})



	nsdq_position, err := client.GetPosition("NSDQ")
	if err != nil {
		fmt.Println("No NSDQ position.")
	} else {
		fmt.Printf("NSDQ position: %v shares.\n", nsdq_position.Qty)
	}

	positions, err := client.ListPositions()
	if err != nil {
		fmt.Println("No positions found.")
	} else {
		// Print the quantity of shares for each position.
		for _, position := range positions {
			fmt.Printf("%v shares in %s", position.Qty, position.Symbol)
		}
	}

}
