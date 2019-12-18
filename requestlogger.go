package main

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

func PrintErrorLog(requestId string, message string) {
	log.Printf(":%q:ERROR:%q", requestId, message)
}

func PrintInfoLog(requestId string, message string) {
	log.Printf(":%q:INFO:%q", requestId, message)
}

func PrintFatalLog(requestId string, message string) {
	log.Fatalf(":%q:FATAL:%q", requestId, message)
}
