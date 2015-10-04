package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"
	"strings"
)

//Response json structure
type Response struct {
	List struct {
		Meta struct {
			Count int    `json:"count"`
			Start int    `json:"start"`
			Type  string `json:"type"`
		} `json:"meta"`
		Resources []struct {
			Resource struct {
				Classname string `json:"classname"`
				Fields    struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					Utctime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

//ClientRequest structure
type ClientRequest struct {
	Budget             string
	StocksymbolPercent map[string]int
}

//Initializing global to persist the values
var (
	tradeID              string
	tradeIDtoresponseMap = make(map[string]map[string]map[string]string)
	tradeIDtobudgetMap   = make(map[string]string)
)

//JSONResponse structure
type JSONResponse struct{}

//GetStockValue function
func (j *JSONResponse) GetStockValue(args *ClientRequest, reply *map[string]map[string]map[string]string) error {

	var (
		jsonResp        Response
		amtForEachStock float64
		stockSymbols    []string
		responseMap     map[string]map[string]string
	)

	for key := range args.StocksymbolPercent {
		stockSymbols = append(stockSymbols, key)
	}

	response, err := http.Get("http://finance.yahoo.com/webservice/v1/symbols/" + strings.Join(stockSymbols, ",") + "/quote?format=json")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		json.Unmarshal([]byte(contents), &jsonResp)
	}

	length := jsonResp.List.Meta.Count

	stockSymbolPriceMap := map[string]float64{}

	for i := 0; i < length; i++ {
		stockSymbolPriceMap[jsonResp.List.Resources[i].Resource.Fields.Symbol], _ = strconv.ParseFloat(jsonResp.List.Resources[i].Resource.Fields.Price, 64)
	}

	responseMap = make(map[string]map[string]string)

	//Performing computations and storing in a map to return to the client
	for key, value := range args.StocksymbolPercent {
		for symbol, stockPrice := range stockSymbolPriceMap {
			if key == symbol {
				if _, ok := tradeIDtoresponseMap["1"]; ok {
					tradeIDtoresponseMapLength := len(tradeIDtoresponseMap)
					nestedMap := make(map[string]string)
					budgetInt, _ := strconv.Atoi(args.Budget)
					amtForEachStock = (float64(value) / float64(100)) * float64(budgetInt)
					stockPriceString := strconv.FormatFloat(stockPrice, 'f', 3, 64)
					numberofShare := int((amtForEachStock) / stockPrice)
					shareString := strconv.Itoa(numberofShare)
					nestedMap[shareString] = stockPriceString
					responseMap[key] = nestedMap
					tradeID = strconv.Itoa(tradeIDtoresponseMapLength + 1)
				} else {
					nestedMap := make(map[string]string)
					budgetInt, _ := strconv.Atoi(args.Budget)
					amtForEachStock = (float64(value) / float64(100)) * float64(budgetInt)
					stockPriceString := strconv.FormatFloat(stockPrice, 'f', 3, 64)
					numberofShare := int((amtForEachStock) / stockPrice)
					shareString := strconv.Itoa(numberofShare)
					nestedMap[shareString] = stockPriceString
					responseMap[key] = nestedMap
					tradeID = "1"
				}
			}
		}
	}

	tradeIDtobudgetMap[tradeID] = args.Budget
	tradeIDtoresponseMap[tradeID] = responseMap
	*reply = tradeIDtoresponseMap

	return nil
}

//GetProfileData to fetch the new stock price based on the trade ID
func (j *JSONResponse) GetProfileData(reqTradeID string, reply2 *map[string]map[string]map[string]string) error {

	var (
		s              Response
		stockSymbolArr []string
		profileDataMap = make(map[string]map[string]map[string]string)
		fetchStockArr  []string
		i              = 0
	)

	for reqTradeIDKey, value := range tradeIDtoresponseMap {
		if reqTradeIDKey == reqTradeID {
			fetchStockArr = make([]string, len(value), len(value)*2)
			for nestedKey := range value {
				fetchStockArr[i] = nestedKey
				i++
			}
		}
	}

	for i := 0; i < len(fetchStockArr); i++ {
		stockSymbolArr = append(stockSymbolArr, fetchStockArr[i])
	}

	response, err := http.Get("http://finance.yahoo.com/webservice/v1/symbols/" + strings.Join(stockSymbolArr, ",") + "/quote?format=json")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		json.Unmarshal([]byte(contents), &s)
	}
	length := s.List.Meta.Count

	stockSymbolPriceMap := map[string]float64{}

	for i := 0; i < length; i++ {
		stockSymbolPriceMap[s.List.Resources[i].Resource.Fields.Symbol], _ = strconv.ParseFloat(s.List.Resources[i].Resource.Fields.Price, 64)
	}
	stocktoPriceDataMap := make(map[string]map[string]string)

	//Peforming computations and storing in a map to return to the client
	for key, value := range stockSymbolPriceMap {
		for tradeidKey, tradeidvalues := range tradeIDtoresponseMap {
			if tradeidKey == reqTradeID {
				for stockKey, stockvaluesMap := range tradeidvalues {
					if key == stockKey {
						stockCounttoPrice := make(map[string]string)
						for numberofstocks, stockprice := range stockvaluesMap {
							iStockPrice, _ := strconv.ParseFloat(stockprice, 64)
							stockPriceRound := Round(iStockPrice, 0.5, 3)
							valueRound := Round(value, 0.5, 3)
							if valueRound > stockPriceRound {
								stockCounttoPrice[numberofstocks] = "+$" + strconv.FormatFloat(valueRound, 'f', 3, 64)
							} else if valueRound < stockPriceRound {
								stockCounttoPrice[numberofstocks] = "-$" + strconv.FormatFloat(valueRound, 'f', 3, 64)
							} else {
								stockCounttoPrice[numberofstocks] = "$" + strconv.FormatFloat(valueRound, 'f', 3, 64)
							}
						}
						stocktoPriceDataMap[key] = stockCounttoPrice
					}
				}
			}
		}
	}

	budget := tradeIDtobudgetMap[reqTradeID]
	profileDataMap[budget] = stocktoPriceDataMap

	*reply2 = profileDataMap
	return nil
}

//Round function to round the float values to 3 precision
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

func main() {
	cal := new(JSONResponse)
	server := rpc.NewServer()
	server.Register(cal)
	server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	listener, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		if conn, err := listener.Accept(); err != nil {
			log.Fatal("accept error: " + err.Error())
		} else {
			log.Printf("new connection established\n")
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}
}
