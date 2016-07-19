package main

import (
    "net/http"
    "encoding/json"
)

type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Kelvin float64 `json:"temp"`
    } `json:"main"`
}

func main() {
    http.HandleFunc("/", hello)
    http.ListenAndServe(":8080", nil)
}

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello!"))
}

func query(city string) (weatherData, error) {
    resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=b772eefc824c6bf22bba33f8fada5ed6&q=" + city)
    if err != nil {
        return weatherData{}, err
    }

    defer resp.Body.Close()

    var weatherInfo weatherData

    if err := json.NewDecoder(resp.Body).Decode(&weatherInfo); err != nil {
        return weatherData{}, err
    }

    return weatherInfo, nil
}