package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
)

//ClientRequest structure
type ClientRequest struct {
	Budget             string
	StocksymbolPercent map[string]int
}

var reply = make(map[string]map[string]map[string]string)

func main() {

	var (
		input                  int
		inputLength            int
		tradeID                string
		stkSymbolPercent       string
		budget                 string
		R                      ClientRequest
		totalsum               float64
		tradeIDtoUninvestedAmt = make(map[string]float64)
	)

	for {
		client, err := net.Dial("tcp", "127.0.0.1:1234")
		if err != nil {
			log.Fatal("dialing:", err)
		}
		fmt.Println("Press 1 to buy stocks")
		fmt.Println("Press 2 to check your profile")
		fmt.Println("Press 3 to exit")
		n, err := fmt.Scanln(&input)

		if n < 1 || err != nil {
			fmt.Println("invalid input")
			return
		}

		//Switch case to perform BuyStock and check profile
		switch input {
		case 1:
			{
				fmt.Println("Enter the stock symbol with percentage. For eg: GOOG:80,AAPL:20")
				n, err := fmt.Scanln(&stkSymbolPercent)
				fmt.Println("Enter your budget")
				m, errr := fmt.Scanln(&budget)
				if n < 1 || err != nil || m < 1 || errr != nil {
					fmt.Println("invalid input")
					return
				}

				symbolPercent := strings.Split(stkSymbolPercent, ",")
				inputLength = len(symbolPercent)
				stockSymbolAndPercentage := make([]string, inputLength, inputLength*2)

				for i := 0; i < len(symbolPercent); i++ {
					stockSymbolAndPercentage[i] = symbolPercent[i]
				}

				R.StocksymbolPercent = make(map[string]int)
				R.Budget = budget

				for i := 0; i < len(stockSymbolAndPercentage); i++ {
					stockPercent := strings.Split(stockSymbolAndPercentage[i], ":")
					stock, percent := stockPercent[0], stockPercent[1]
					iPercent, err := strconv.Atoi(percent)

					if err != nil {
						log.Fatal("Conversion from string to int error:", err)
					}

					R.StocksymbolPercent[stock] = iPercent
				}

				var sum = 0

				for _, value := range R.StocksymbolPercent {
					sum = sum + value
				}

				if len(R.StocksymbolPercent) != len(stockSymbolAndPercentage) {
					log.Fatal("You are trying to fetch price of 2 same stock symbols..Give a valid input")
				}

				if sum != 100 {
					log.Fatal("Total of percentage is less than 100")
				}

				args := &ClientRequest{R.Budget, R.StocksymbolPercent}
				c := jsonrpc.NewClient(client)
				err = c.Call("JSONResponse.GetStockValue", args, &reply)

				if err != nil {
					log.Fatal("Response error:", err)
				}

				length := len(reply)
				sLen := strconv.Itoa(length)

				totalsum = 0
				loopCount := 0

				//iterating through the response to display the data to the user.
				for key, value := range reply {
					if key == sLen {
						fmt.Println("TradeID: ", key)
						fmt.Print("Stocks: ")
						for nestedkey, nestedvalueMap := range value {
							fmt.Print(nestedkey)
							for childKey, childValue := range nestedvalueMap {
								fmt.Print(":", childKey)
								fmt.Print(":$", childValue)
								iChildKey, _ := strconv.Atoi(childKey)
								iChildValue, _ := strconv.ParseFloat(childValue, 64)
								totalsum = (float64(iChildKey) * float64(iChildValue)) + totalsum
							} // close child loop
							loopCount++
							if len(value) != loopCount {
								fmt.Print(", ")
							}
						} // close nested loop
						ibudget, _ := strconv.Atoi(budget)
						uninvestedAmt := (float64(ibudget) - totalsum)
						tradeIDtoUninvestedAmt[key] = Round(uninvestedAmt, 0.5, 3)
						fmt.Println("\nUnvested Amount: $", tradeIDtoUninvestedAmt[key])
					} // End of loop

				} // End for

			}
		case 2:
			{
				var (
					totalMarketValue float64
					reply2           map[string]map[string]map[string]string
					reqBudget        string
					stringPrice      string
				)

				fmt.Println("Enter the trading ID to view your profile")
				n, err := fmt.Scanln(&tradeID)
				if n < 1 || err != nil {
					fmt.Println("invalid input")
					return
				}

				c := jsonrpc.NewClient(client)
				err = c.Call("JSONResponse.GetProfileData", tradeID, &reply2)

				if err != nil {
					log.Fatal("Response error:", err)
				}

				totalMarketValue = 0
				loopcount := 0

				//iterating through the reply2 map to display the profile data to the user
				for reply2Key, reply2ValueMap := range reply2 {
					reqBudget = reply2Key
					fmt.Print("Stocks: ")
					for reply2ValueMapKey, reply2ValueMapValue := range reply2ValueMap {
						fmt.Print(reply2ValueMapKey)
						for share, price := range reply2ValueMapValue {
							ishare, _ := strconv.Atoi(share)

							if strings.Contains(price, "+$") {
								stringPrice = StripChar(price, "+$")
							} else if strings.Contains(price, "-$") {
								stringPrice = StripChar(price, "-$")
							} else {
								stringPrice = StripChar(price, "$")
							}

							fprice, _ := strconv.ParseFloat(stringPrice, 64)
							totalMarketValue = (float64(ishare) * fprice) + totalMarketValue
							fmt.Print(":", share)
							fmt.Print(":", price)
						}
						loopcount++
						if loopcount != len(reply2ValueMap) {
							fmt.Print(", ")
						}
					}
				}

				ireqBudget, _ := strconv.Atoi(reqBudget)
				leftoutAmt := float64(ireqBudget) - totalMarketValue
				fmt.Println("")
				fmt.Println("Current Market Value: $", totalMarketValue)
				fmt.Println("Unvested Amount: $", Round(leftoutAmt, 0.5, 3))

			}
		case 3:
			{
				os.Exit(2)
			}
		}
	}
}

//StripChar function to strip the given character from the string
func StripChar(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

//Round function for rounding the value
func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
