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
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	apiKey  = "YN1TUR6BQZWOUNTG"
	symbol1 = "VOO"
	symbol2 = "SPGI"
)

func fetchData(apiKey string, symbol string) ([]float64, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s", symbol, apiKey)
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

	timeSeriesData, ok := data["Time Series (Daily)"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to parse time series data")
	}

	var prices []float64
	currentDate := time.Now()
	count := 0

	for dateStr, entry := range timeSeriesData {
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

		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %v", err)
		}

		if currentDate.Sub(date).Hours() <= 365*24 {
			prices = append(prices, price)
			count++
		}
	}

	return prices, nil
}

func main() {
	Y, err := fetchData(apiKey, symbol1)
	if err != nil {
		log.Fatal("Failed to fetch real-time data:", err)
	}

	X, err := fetchData(apiKey, symbol2)
	if err != nil {
		log.Fatal("Failed to fetch real-time data:", err)
	}

	plotHistoricalDataAndSpread(X, Y)
	exportToCSV(X, "data1.csv")
	exportToCSV(Y, "data2.csv")
}

func plotHistoricalDataAndSpread(data1, data2 []float64) {
	page := components.NewPage()

	plotLineChart(page, "Stock Prices", data1, data2, "VOO", "SPGI")

	spreadData := calculateSpread(data1, data2)
	plotLineChart(page, "Stock Price Spread", spreadData, nil, "Spread", "")

	sma10 := calculateSMA(spreadData, 10)
	sma30 := calculateSMA(spreadData, 30)

	plotSMAChart(page, "SMAs on Spread", spreadData, sma10, sma30)
}

func plotLineChart(page *components.Page, title string, data1, data2 []float64, seriesNames ...string) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
	)

	xData := make([]string, len(data1))
	for i := range xData {
		xData[i] = strconv.Itoa(i + 1)
	}

	seriesData := make([][]opts.LineData, len(seriesNames))
	dataArrays := [][]float64{data1, data2}

	for i := range seriesNames {
		if i < len(dataArrays) && dataArrays[i] != nil {
			seriesData[i] = make([]opts.LineData, len(dataArrays[i]))
			for j, val := range dataArrays[i] {
				seriesData[i][j] = opts.LineData{Value: val}
			}
		}
	}

	line.SetXAxis(xData)

	for i, seriesName := range seriesNames {
		if seriesData[i] != nil {
			line.AddSeries(seriesName, seriesData[i])
		}
	}

	setCommonChartOptions(line)

	page.AddCharts(line)
}
func plotSMAChart(page *components.Page, title string, data []float64, sma10, sma30 []float64) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
		}),
	)

	xData := make([]string, len(data)-30+1)
	for i := range xData {
		xData[i] = strconv.Itoa(i) // Start from day 31
	}

	yData := make([][]opts.LineData, 3)
	yData[0] = make([]opts.LineData, len(data))
	yData[1] = make([]opts.LineData, len(sma10))
	yData[2] = make([]opts.LineData, len(sma30))

	for i, val := range data {
		yData[0][i] = opts.LineData{Value: val}
	}

	for i, val := range sma10 {
		yData[1][i] = opts.LineData{Value: val}
	}

	for i, val := range sma30 {
		yData[2][i] = opts.LineData{Value: val}
	}

	line.SetXAxis(len(data)).
		AddSeries("Spread", yData[0]).
		AddSeries("10-day SMA", yData[1]).
		AddSeries("30-day SMA", yData[2])

	page.AddCharts(line)
	f, err := os.Create("stock_data_and_spread.html")
	if err != nil {
		log.Fatal(err)
	}
	page.Render(io.MultiWriter(f))
	fmt.Println("Created Graph to show SMA")
}

func setCommonChartOptions(chart *charts.Line) {
	chart.SetGlobalOptions(
		charts.WithYAxisOpts(opts.YAxis{}),
		charts.WithDataZoomOpts(opts.DataZoom{
			YAxisIndex: []int{0},
		}),
	)
}

func calculateSpread(data1, data2 []float64) []float64 {
	if len(data1) != len(data2) {
		return nil
	}

	spread := make([]float64, len(data1))
	for i := 0; i < len(data1); i++ {
		spread[i] = data1[i] - data2[i]
	}

	return spread
}

func calculateSMA(data []float64, window int) []float64 {
	if window <= 0 || window > len(data) {
		return nil
	}

	sma := make([]float64, len(data)-window+1)

	for i := 0; i <= len(data)-window; i++ {
		sum := 0.0
		for j := i; j < i+window; j++ {
			sum += data[j]
		}
		sma[i] = sum / float64(window)
	}

	return sma
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

func getMean(data []float64) float64 {
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	return sum / float64(len(data))
}

func getStandardDeviation(data []float64) float64 {
	mean := getMean(data)
	variance := 0.0
	for _, value := range data {
		deviation := value - mean
		variance += deviation * deviation
	}
	variance /= float64(len(data))
	return math.Sqrt(variance)
}

func calculateZScore(x, mean, stddev float64) float64 {
	return (x - mean) / stddev
}
