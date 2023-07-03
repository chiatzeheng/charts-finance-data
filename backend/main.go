package main

import (
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

const (
	apiKey  = "YN1TUR6BQZWOUNTG"
	symbol1 = "AAPL"
	symbol2 = "GOOG"
)

func fetchRealTimeData(apiKey string, symbol string) ([]float64, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_WEEKLY_ADJUSTED&symbol=%s&apikey=%s", symbol, apiKey)
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

	timeSeriesData, ok := data["Weekly Adjusted Time Series"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse time series data")
	}

	var prices []float64
	for _, entry := range timeSeriesData {
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
	}

	return prices, nil
}

func main() {
	applData, err := fetchRealTimeData(apiKey, symbol1)
	if err != nil {
		log.Fatal("Failed to fetch real-time data:", err)
	}

	msftData, err := fetchRealTimeData(apiKey, symbol2)
	if err != nil {
		log.Fatal("Failed to fetch real-time data:", err)
	}

	// fmt.Println("AAPL:", applData, "\nMSFT:", msftData)

	graph(applData, msftData)

	calculateCorrelation(applData)
	calculateCorrelation(msftData)
}

func graph(data1, data2 []float64) {

	line := charts.NewLine()

	// Set the title of the chart
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Stock Prices",
		}),
	)

	// Create x-axis values (week numbers)
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
		AddSeries("AAPL", yData1).
		AddSeries("MSFT", yData2)

	// Set the line color for AAPL series
	line.SetSeriesOptions(
		charts.WithLineStyleOpts(opts.LineStyle{
			Color: "#FF0000", // Red color for AAPL line
			Width: 2,         // Line width
		}),
	)

	// Set the line color for MSFT series
	line.SetSeriesOptions(
		charts.WithLineStyleOpts(opts.LineStyle{
			Color: "#0000FF", // Blue color for MSFT line
			Width: 2,         // Line width
		}),
	)

	// Generate the HTML file with the chart
	page := components.NewPage()
	page.AddCharts(line)
	f, err := os.Create("stock_prices.html")
	if err != nil {
		log.Fatal(err)
	}
	page.Render(io.MultiWriter(f))

	fmt.Println("Graph saved to stock_prices.html")
}

func calculateCorrelation(tradingData []float64) {
	// Calculate the mean value of the stock prices
	mean := getMeanPrice(tradingData)

	// Calculate the standard deviation of the stock prices
	stdDev := getStandardDeviation(tradingData)

	// Create a slice to store trading signals
	signals := make([]int, len(tradingData))

	// Generate trading signals based on mean reversion strategy
	for i, val := range tradingData {
		if val > mean+stdDev {
			// Price is above the mean + standard deviation, consider selling
			signals[i] = -1
		} else if val < mean-stdDev {
			// Price is below the mean - standard deviation, consider buying
			signals[i] = 1
		} else {
			// Price is within the mean +/- standard deviation range, do nothing
			signals[i] = 0
		}
	}

	// Implement the trading strategy based on the signals
	// For simplicity, let's assume we start with no position
	position := 0

	for i := 1; i < len(tradingData); i++ {
		if signals[i] == 1 && position == 0 {
			// Generate a buy signal and open a long position
			position = 1
			fmt.Printf("Buy at index %d\n", i)
		} else if signals[i] == -1 && position == 1 {
			// Generate a sell signal and close the long position
			position = 0
			fmt.Printf("Sell at index %d\n", i)
		}
	}
}

func getMeanPrice(prices []float64) float64 {
	sum := 0.0
	for _, price := range prices {
		sum += price
	}
	return sum / float64(len(prices))
}

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
