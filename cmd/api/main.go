package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8080", "port")
	dbCreds := flag.String("db", "", "db connections credentials")
	level := flag.String("loglevel", "info", "log level")
	flag.Parse()

	fmt.Println(*port, *dbCreds, *level)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello!")
	})

	fmt.Printf("Starting server at port %s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}
