package main

import (
    "net/http"
    "encoding/json"
    "strings"
    "log"
    "time"
)

type weatherProvider interface {
    queryTemperature(city string) (float64, error)
}

type multiWeatherProvider []weatherProvider

type openWeatherMap struct{
    apiKey string
}

type weatherUnderground struct{
    apiKey string
}

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

func (w multiWeatherProvider) getTemperature(city string) (float64, error) {
    tempSum := 0.0

    for _, provider := range w {
        temp, err := provider.queryTemperature(city)
        if err != nil {
            return 0, nil
        }

        tempSum += temp
    }

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
