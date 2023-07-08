package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// For websocket 1 API call every 3 mins is allowed

//I want to get Data from the past 6 years
//Market is open 252 in a given year

const (
	apiKey  = "YN1TUR6BQZWOUNTG"
	symbol1 = "SPY"
	symbol2 = "NFLX"
)

func fetchRealTimeData(apiKey string, symbol string) ([]float64, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_MONTHLY&symbol=%s&interval=5min&outputsize=compact&apikey=%s", symbol, apiKey)
	fmt.Println("URL:", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	timeSeriesData, ok := data["Monthly Time Series"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse time series data")
	}

	var prices []float64
	count := 0
	for _, entry := range timeSeriesData {
		if count >= 48 { // Considering 12 months in a year for 4 years
			break
		}

		entryData, ok := entry.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse entry data")
		}

		closePrice, ok := entryData["4. close"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to parse close price")
		}

		price, err := strconv.ParseFloat(closePrice, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to convert close price to float")
		}

		prices = append(prices, price)
		count++
	}

	return prices, nil
}

func main() {
	stockdata1, err := fetchRealTimeData(apiKey, symbol1)
	if err != nil {
		log.Fatal("Failed to fetch real-time data:", err)
	}

	stockdata2, err := fetchRealTimeData(apiKey, symbol2)
	if err != nil {
		log.Fatal("Failed to fetch real-time data:", err)
	}

	exportedData := calculateSpread(stockdata1, stockdata2)
	exportToCSV(exportedData, "spread_of_stock_prices.csv")
	exportToCSV(stockdata1, "stock_price_1.csv")
	exportToCSV(stockdata2, "stock_price_2.csv")
	fmt.Println("Data exported to CSV successfully.")

	// fmt.Println("Stock Data 1:", stockdata1)
	// fmt.Println("Stock Data 2:", stockdata2)

	// graphofhistoricalData(stockdata1, stockdata2)

	// spreadGraph(stockdata1, stockdata2)

}

func graphofhistoricalData(data1, data2 []float64) {
	line := charts.NewLine()

	// Set the title of the chart
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Stock Prices",
		}),
	)

	// Create x-axis values (indices)
	xData := make([]string, len(data1))
	for i := range xData {
		xData[i] = strconv.Itoa(i + 1)
	}

	// Create y-axis values for AAPL stock prices
	yData1 := make([]opts.LineData, len(data1))
	for i, val := range data1 {
		yData1[i] = opts.LineData{Value: val}
	}

	// Create y-axis values for MSFT stock prices
	yData2 := make([]opts.LineData, len(data2))
	for i, val := range data2 {
		yData2[i] = opts.LineData{Value: val}
	}

	// Add the data series to the chart
	line.SetXAxis(xData).
		AddSeries("SPY", yData1).
		AddSeries("NFLX", yData2).
		SetGlobalOptions(
			charts.WithYAxisOpts(opts.YAxis{}),
			charts.WithDataZoomOpts(opts.DataZoom{
				YAxisIndex: []int{0}, // Enable data zoom for the y-axis
			}),
		)
	page := components.NewPage()
	page.AddCharts(line)
	f, err := os.Create("stock_prices.html")
	if err != nil {
		log.Fatal(err)
	}
	page.Render(io.MultiWriter(f))

	fmt.Println("Created Graph to show relationship between stock price")
}

func calculateSpread(data1, data2 []float64) []float64 {
	if len(data1) != len(data2) {
		// Ensure the input data slices have the same length
		return nil
	}

	spread := make([]float64, len(data1))
	for i := 0; i < len(data1); i++ {
		spread[i] = data1[i] - data2[i]
	}

	return spread
}

func spreadGraph(data1, data2 []float64) {
	line := charts.NewLine()

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Stock Price Spread",
		}),
	)

	// Calculate the spread between the stock prices
	spreadData := calculateSpread(data1, data2)

	xData := make([]string, len(spreadData))
	for i := range xData {
		xData[i] = strconv.Itoa(i + 1)
	}

	yData := make([]opts.LineData, len(spreadData))
	for i, val := range spreadData {
		yData[i] = opts.LineData{Value: val}
	}

	// Add the data series to the chart
	line.SetXAxis(xData).
		AddSeries("Spread", yData).
		SetGlobalOptions(
			charts.WithYAxisOpts(opts.YAxis{}),
			charts.WithDataZoomOpts(opts.DataZoom{
				YAxisIndex: []int{0},
			}),
		)
	page := components.NewPage()
	page.AddCharts(line)
	f, err := os.Create("stock_price_spread.html")
	if err != nil {
		log.Fatal(err)
	}
	page.Render(io.MultiWriter(f))

	fmt.Println("Created Graph to show stock price spread")
}

// Average of all the prices
func getMeanPrice(prices []float64) float64 {
	sum := 0.0
	for _, price := range prices {
		sum += price
	}
	return sum / float64(len(prices))
}

// Denoted by Sqrt of Summation of value of each price minus the
// mean price squared divided by the number of prices
func getStandardDeviation(prices []float64) float64 {
	mean := getMeanPrice(prices)
	sum := 0.0
	for _, price := range prices {
		deviation := price - mean
		sum += deviation * deviation
	}
	variance := sum / float64(len(prices))
	return math.Sqrt(variance)
}

func exportToCSV(data []float64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range data {
		err := writer.Write([]string{strconv.FormatFloat(value, 'f', -1, 64)})
		if err != nil {
			return err
		}
	}

	return nil
}
