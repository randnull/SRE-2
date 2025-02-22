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
	}
	w.WriteHeader(http.StatusOK)
}
```

Возможные результаты:


