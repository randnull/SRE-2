package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"time"
)

func GetData(url string) (string, error) {
	repeatCountParam := 3
	allowedStatusCode := []int{500, 502, 503, 504}

	for attemptNumber := 0; attemptNumber < repeatCountParam; attemptNumber++ {
		log.Println("Attempt to get ", url)
		resp, err := http.Get(url)

		if err != nil {
			log.Println("Error with request")
			return "", err
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)

			if err != nil {
				log.Println("Error with parsing data")
				return "", nil
			}

			log.Printf("Success in %v attempts", attemptNumber+1)
			return string(body), nil
		}

		if !slices.Contains(allowedStatusCode, resp.StatusCode) {
			log.Printf("Error status code not allowed %v", resp.StatusCode)
			return "", err
		}

		log.Println("Failed. Get: ", resp.StatusCode)

		if attemptNumber != repeatCountParam-1 {
			sleepTime := attemptNumber + 1
			log.Printf("Retry. Sleeping for: %v seconds", sleepTime)
			time.Sleep(time.Second * time.Duration(sleepTime))
		}
	}

	return "", errors.New("error with request")
}

func main() {
	var SiteUrl string = "http://127.0.0.1:8000/example"

	answer, err := GetData(SiteUrl)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}
