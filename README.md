# Задача 1: Надежный клиент для внешнего API

Решение находится в папке task1

GetData:
```
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
```

Пример работы:

Запускаем тестовый сервер (server.go)
Сервер случайно (50%) отдает либо BadRequest, либо OK.

```
func ExampleHandler(w http.ResponseWriter, req *http.Request) {
	if rand.Int()%10 < 5 {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	fmt.Fprintf(w, "Message from server!")
	w.WriteHeader(http.StatusOK)
}
```

Возможные результаты:

Успешно со второй попытки:
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex1.png)

Не успешно:
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex2.png)

Успешно сразу:
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex3.png)

# Задача 2: Реализация Circuit Breaker для клиента API

Решение находится в папке task2

GetDataWithCircuitBreaker:
```
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
```

Результат работы:

Подробный результат работы приложен в файле log.txt в папке task2 (task2/log.txt)

Обычная работа (ошибки не превышают 2 подряд):
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex7.png)

3 ошибки подряд:
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex6.png)

HalfOpen + ошибка после него:
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex5.png)

HalfOpen + успех после него:
![Иллюстрация к проекту](https://github.com/randnull/SRE-2/blob/main/images/ex4.png)
