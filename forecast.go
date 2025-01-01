package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type coord struct {
        latitude float64
        longitude float64
}


// I found the api url to convert the location name into geographic coordinate system, e.g. latitude and longitude
// it is needed for the later API that I am going to use
func getCoordinates (location string) (coordinates coord, err error) {
        limit := 1 // i'll limit the search to only return one agreed location
        api_url := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=%d&appid=%s", location, limit, API_KEY)

        resp, err := http.Get(api_url)

        if err != nil {
                fmt.Println("error status code: ", resp.StatusCode)
                return coord{}, errors.New(err.Error())
        }

        // It is preferable to process other status code
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        var objbody []map[string]interface{}

        if e := json.Unmarshal(body, &objbody); e != nil {
                fmt.Println("Unmarshal error: ", e.Error())
                return coord{}, errors.New(e.Error())
        }

        coordinates.latitude = (objbody[0]["lat"]).(float64)
        coordinates.longitude = (objbody[0]["lon"]).(float64)
        
        return coordinates, nil
}

var API_KEY = os.Getenv("API_KEY")

func main () {
        coordinates, err := getCoordinates("Jakarta")

        if err != nil {
                fmt.Println("Error in my get city coordinate call")
                return
        }

        // i was planning to use the "Daily Forecast 16 Days API," but I just realized that it is for paid plans
        // so i will utilize the "Call 5 day / 3 hour forecast data" API
        forecast_api_url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?lat=%f&lon=%f&units=metric&appid=%s", coordinates.latitude, coordinates.longitude, API_KEY)

        resp, err := http.Get(forecast_api_url)

        if err != nil {
                fmt.Println("Error in my forecast api call: ", err.Error())
        }

        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)

        // bear with me...
        var objbody map[string] interface{}


        if e := json.Unmarshal(body, &objbody); e != nil {
                fmt.Println(e)
        }

        obj := objbody["list"].([]interface{})  // this would be the list that contains 40 weather forecasts
                                                // (one forecast every three hours, for 5 days)
                                                // i will pick forecast-0, forecast-7, forecast-15, ..., forecast-31

        // using datetime in go is kinda painful, but the blog from Noval Agung really helped
        // https://dasarpemrogramangolang.novalagung.com/A-time-parsing-format.html
        expected_remainder := 7

        layout_format := "2006-01-02 15:04:05"
        fmt.Println("Weather Forecast for Jakarta")
        for idx, forecast := range obj {
                if idx % 8 == expected_remainder {
                        forecast_obj := forecast.(map[string]interface{})
                        datetime := forecast_obj["dt_txt"].(string)
                        date, _ := time.Parse(layout_format, datetime)

                        date1 := date.Format("Mon, 02 Jan 2006")

                        temperature := forecast_obj["main"].(map[string]interface{})["temp"]
                        temperature_forecast := fmt.Sprintf("%s: %.2f\u00BAC", date1, temperature)
                        fmt.Println(temperature_forecast)
                }
        }
}
