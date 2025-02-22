package main

import (
	"log"
	"math/rand"
	"net/http"
)

func ExampleHandler(w http.ResponseWriter, req *http.Request) {
	if rand.Int()%10 < 5 {
		w.WriteHeader(http.StatusBadGateway)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/example", ExampleHandler)
	log.Println("Start Listen 0.0.0.0:8000!")
	http.ListenAndServe(":8000", nil)
}
