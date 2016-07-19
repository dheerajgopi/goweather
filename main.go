package main

import (
    "net/http"
    "encoding/json"
    "strings"
)

type weatherData struct {
    Name string `json:"name"`
    Main struct {
        Kelvin float64 `json:"temp"`
    } `json:"main"`
}

func main() {
    http.HandleFunc("/", hello)
    http.HandleFunc("/weather/", weatherJson)
    http.ListenAndServe(":8080", nil)
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

func hello(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello!"))
}

func weatherJson(w http.ResponseWriter, r *http.Request) {
    city := strings.SplitN(r.URL.Path, "/", 3)[2]

    data, err := query(city)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    json.NewEncoder(w).Encode(data)
}
