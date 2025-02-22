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

type CircuitBreaker struct {
	currentState       string
	countOfFailed      int
	lastCloseTimestamp time.Time
}

var CBreaker = CircuitBreaker{
	currentState:  ClosedState,
	countOfFailed: 0,
}

var allowedStatusCode = []int{500, 502, 503, 504}

var ClosedState = "Closed"
var HalfOpenState = "HalfOpen"
var OpenState = "Open"

func (c *CircuitBreaker) CheckCircuitBreakerStatus() bool {
	if c.currentState == ClosedState {
		return true
	} else if c.currentState == OpenState {
		currentTime := time.Now()
		if int(currentTime.Sub(c.lastCloseTimestamp)/time.Second) > 10 {
			log.Printf("[CB] Change CircuitBreaker state from %v to HalfOpen", c.currentState)
			c.countOfFailed = 0
			c.currentState = HalfOpenState
			return true
		}
		return false
	} else if c.currentState == HalfOpenState {
		return true
	}
	return false
}

func (c *CircuitBreaker) ProceedError() {
	c.countOfFailed += 1
	if c.currentState == HalfOpenState || c.countOfFailed >= 3 {
		log.Printf("[CB] Change CircuitBreaker state from %v to Open", c.currentState)
		c.currentState = OpenState
		c.lastCloseTimestamp = time.Now()
	}
}

func GetDataWithCircuitBreaker(url string) (string, error) {
	if !CBreaker.CheckCircuitBreakerStatus() {
		return "", errors.New("[INFO] request not allowed. Open State")
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)

	// Это не считаем, так как ошибка не входит в наши статус-коды
	if err != nil {
		log.Println("[INFO] Error with request")
		return "", err
	}

	if resp.StatusCode == 200 {
		log.Println("[INFO] Success request!")

		if CBreaker.currentState != ClosedState {
			log.Printf("[CB] Change CircuitBreaker state from %v to Closed", CBreaker.currentState)
			CBreaker.currentState = ClosedState
		}
		CBreaker.countOfFailed = 0

		body, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Println("[INFO] Error with parsing data")
			return "", nil
		}

		return string(body), nil
	}

	// Это не считаем, так как ошибка не входит в наши статус-коды
	if !slices.Contains(allowedStatusCode, resp.StatusCode) {
		log.Printf("[INFO] Error status code not allowed %v", resp.StatusCode)
		return "", err
	}

	CBreaker.ProceedError()

	return "", errors.New("[ERROR] error with request")
}

func main() {
	var SiteUrl string = "http://127.0.0.1:8000/example"

	for i := 0; i < 100; i++ {
		answer, err := GetDataWithCircuitBreaker(SiteUrl)

		fmt.Println("[INFO] count of errors:", CBreaker.countOfFailed)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(1 * time.Second)

		fmt.Println(answer)
	}

}
