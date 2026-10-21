package main

import (
	"fmt"

	"github.com/AmiraliQdi/golang-ghostops/internal/log"
)

func main() {
	rawLog := "2026-10-03T11:51:21Z WARN payment-service Rate limit approaching for client IP 192.168.1.50"
	parsedLog, err := log.New(rawLog)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(parsedLog)
}
