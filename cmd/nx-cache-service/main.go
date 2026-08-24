package main

import (
	"fmt"
	"net/http"
)

func main() {
	const PORT = 8080

	fmt.Println("Starting the NX Cache Service on port", PORT)

	_ = http.ListenAndServe("localhost:8080", nil)
}
