package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type WeatherData struct {
	ReceiverTime      string  `json:"receiverTime"`
	ReceiverTimestamp int64   `json:"receiverTimestamp"`
	TemperatureIndoor float64 `json:"temperatureIndoor"`
	HumidityIndoor    float64 `json:"humidityIndoor"`
	PressureAbsolute  float64 `json:"pressureAbsolute"`
	PressureRelative  float64 `json:"pressureRelative"`
	Temperature       float64 `json:"temperature"`
	Humidity          float64 `json:"humidity"`
	DewPoint          float64 `json:"dewPoint"`
	WindDir           float64 `json:"windDir"`
	WindDirCardinal   string  `json:"windDirCardinal"`
	WindSpeed         float64 `json:"windSpeed"`
	WindGust          float64 `json:"windGust"`
	WindChill         float64 `json:"windChill"`
	SolarRadiation    float64 `json:"solarRadiation"`
	Uv                float64 `json:"uv"`
	Uvi               float64 `json:"uvi"`
	PrecipHourlyRate  float64 `json:"precipHourlyRate"`
	PrecipDaily       float64 `json:"precipDaily"`
	PrecipWeekly      float64 `json:"precipWeekly"`
	PrecipMonthly     float64 `json:"precipMonthly"`
	PrecipYearly      float64 `json:"precipYearly"`
	HeatIndex         float64 `json:"heatIndex"`
}

var pwsIp string
var fetchInterval int
var debugEnabled bool

func fetchDocumentFromPws() *goquery.Document {

	pwsUrl := fmt.Sprintf("http://%s/livedata.htm", pwsIp)
	// Read the HTML file
	resp, err := http.Get(pwsUrl)
	if err != nil {
		log.Printf("Failed to connect to PWS: %s\n", err)
		return nil
	}
	defer resp.Body.Close()

	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to fetch data from PWS: %s\n", err)
		return nil
	}

	// Parse the HTML file with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlData)))

	if err != nil {
		log.Printf("Failed to process HTML: %s\n", err)
		return nil
	}
	if debugEnabled {
		log.Printf("Fetched data from PWS\n")
	}
	return doc
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
func parseHtml(doc *goquery.Document) WeatherData {

	var weatherData WeatherData

	doc.Find("table").Each(func(i int, s *goquery.Selection) {

		rows := s.Find("tr")

		// parse the table rows and extract the data
		rows.Each(func(i int, s *goquery.Selection) {
			value := s.Find("input").AttrOr("value", "")

			switch i {
			case 8:
				weatherData.ReceiverTime = value
			case 12:
				weatherData.TemperatureIndoor = parseFloat(value)
			case 13:
				weatherData.HumidityIndoor = parseFloat(value)
			case 14:
				weatherData.PressureAbsolute = parseFloat(value)
			case 15:
				weatherData.PressureRelative = parseFloat(value)
			case 16:
				weatherData.Temperature = parseFloat(value)
			case 17:
				weatherData.Humidity = parseFloat(value)
			case 18:
				weatherData.WindDir = parseFloat(value)
			case 19:
				weatherData.WindSpeed = parseFloat(value)
			case 20:
				weatherData.WindGust = parseFloat(value)
			case 21:
				weatherData.SolarRadiation = parseFloat(value)
			case 22:
				weatherData.Uv = parseFloat(value)
			case 23:
				weatherData.Uvi = parseFloat(value)
			case 24:
				weatherData.PrecipHourlyRate = parseFloat(value)
			case 25:
				weatherData.PrecipDaily = parseFloat(value)
			case 26:
				weatherData.PrecipWeekly = parseFloat(value)
			case 27:
				weatherData.PrecipMonthly = parseFloat(value)
			case 28:
				weatherData.PrecipYearly = parseFloat(value)
			default:
				return
			}
		})
	})

	return weatherData
}

func constructUrl(data WeatherData) string {

	currentTime := time.Now()

	urlVars := []string{
		fmt.Sprintf("user=%s", os.Getenv("USERNAME")),
		fmt.Sprintf("pass=%s", os.Getenv("PASSWORD")),
		fmt.Sprintf("ev=%d", currentTime.Year()),
		fmt.Sprintf("ho=%d", currentTime.Month()),
		fmt.Sprintf("nap=%d", currentTime.Day()),
		fmt.Sprintf("ora=%d", currentTime.Hour()),
		fmt.Sprintf("perc=%d", currentTime.Minute()),
		fmt.Sprintf("mp=%d", currentTime.Second()),
		fmt.Sprintf("hom=%.1f", data.Temperature),
		fmt.Sprintf("rh=%.0f", data.Humidity),
		fmt.Sprintf("szelirany=%.0f", data.WindDir),
		fmt.Sprintf("szelero=%.1f", data.WindSpeed),
		fmt.Sprintf("szellokes=%.1f", data.WindGust),
		fmt.Sprintf("p=%.1f", data.PressureRelative),
		fmt.Sprintf("csap=%.2f", data.PrecipDaily),
		fmt.Sprintf("csap1h=%.2f", data.PrecipHourlyRate),
		"tipus=WH2600",
	}

	return strings.Join(urlVars, "&")

}

func sendToidokep(data WeatherData) error {

	url := fmt.Sprintf("https://pro.idokep.hu/sendws.php?%s", constructUrl(data))

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Error posting: %s\n", err)
		return err
	}

	if debugEnabled {
		log.Printf("Success %#v\n", resp.Status)
	}
	return nil
}

func main() {

	debugEnabled = os.Getenv("DEBUG_ENABLED") != ""

	pwsIp = os.Getenv("PWS_IP")
	if pwsIp == "" {
		log.Fatalf("PWS_IP env var undefined")
	}
	fetchIntervalStr := os.Getenv("FETCH_INTERVAL")
	if fetchIntervalStr == "" {
		log.Fatalf("FETCH_INTERVAL env var undefined")
	}
	var err error
	fetchInterval, err = strconv.Atoi(fetchIntervalStr)
	if err != nil {
		log.Fatalf("Invalid FETCH_INTERVAL value: %s\n", fetchIntervalStr)
	}

	for {
		doc := fetchDocumentFromPws()

		if doc != nil {
			weatherData := parseHtml(doc)
			if debugEnabled {
				log.Printf("%#v\n", weatherData)
			}
			sendToidokep(weatherData)
		}
		time.Sleep(time.Duration(fetchInterval) * time.Second)
	}

}
