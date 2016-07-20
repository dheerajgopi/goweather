package main

import (
    "net/http"
    "encoding/json"
    "strings"
    "log"
    "time"
)

// Interface for weather info provider APIs.
type weatherProvider interface {
    queryTemperature(city string) (float64, error)
}

// Type for storing array of weatherProvider interface types
type multiWeatherProvider []weatherProvider

// Type for "open weather map" api
type openWeatherMap struct{
    apiKey string
}

// Type for "wunderground" api
type weatherUnderground struct{
    apiKey string
}

// Method for querying temperature of a city from openweathermap.org
func (w openWeatherMap) queryTemperature(city string) (float64, error) {
    resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + w.apiKey + "&q=" + city)
    if err != nil {
        return 0, err
    }

    defer resp.Body.Close()

    var weatherInfo struct {
        Main struct {
            Kelvin float64 `json:"temp"`
        } `json:"main"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
        return 0, err
    }

    log.Printf("openWeatherMap: %s: %.2f", city, weatherInfo.Main.Kelvin)
    return weatherInfo.Main.Kelvin, nil
}

// Method for querying temperature of a city from wunderground.com/weeather/api
func (w weatherUnderground) queryTemperature(city string) (float64, error) {
    resp, err := http.Get("http://api.wunderground.com/api/" + w.apiKey + "/conditions/q/" + city + ".json")
    if err != nil {
        return 0, err
    }

    defer resp.Body.Close()

    var weatherInfo struct {
        Observation struct {
            Celsius float64 `json:"temp_c"`
        } `json:"current_observation"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
        return 0, err
    }

    kelvin := weatherInfo.Observation.Celsius + 273.15
    log.Printf("weatherUnderground: %s: %.2f", city, kelvin)
    return kelvin, nil
}

// Method to calculate avg of temperatures of a city as provided by each of
// the weather info providers in the multiWeatherProvider type
func (w multiWeatherProvider) getTemperature(city string) (float64, error) {
    temps := make(chan float64, len(w))
    errs := make(chan error, len(w))

    // For each provider spawn a goroutine
    for _, provider := range w {
        go func(wp weatherProvider) {
            t, err := wp.queryTemperature(city)
            if err != nil {
                errs <- err
                return
            }
            temps <- t
        }(provider)
    }

    tempSum := 0.0

    // Collect temperature or error from the providers
    for i := 0; i < len(w); i++ {
        select {
        case temp := <-temps:
            tempSum += temp
        case err := <-errs:
            return 0, err
        }
    }

    // return the avg
    return tempSum / float64(len(w)), nil
}

func main() {
    providers := multiWeatherProvider{
        openWeatherMap{apiKey: "b772eefc824c6bf22bba33f8fada5ed6"},
        weatherUnderground{apiKey: "40d0ecd5b170cf6b"},
    }

    http.HandleFunc("/weather/", func (w http.ResponseWriter, r *http.Request) {
        begin := time.Now()
        city := strings.SplitN(r.URL.Path, "/", 3)[2]

        temp, err := providers.getTemperature(city)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        json.NewEncoder(w).Encode(map[string] interface{} {
            "city": city,
            "temp": temp,
            "timeTook": time.Since(begin).String(),
        })
    })

    http.ListenAndServe(":8080", nil)
}
