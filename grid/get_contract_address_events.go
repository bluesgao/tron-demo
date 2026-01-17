package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {

	url := "https://api.trongrid.io/v1/contracts/TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t/events"

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))

}
