package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/risentveber/wallet-api/services/transfers"
)

func retry(pause time.Duration, maxCount uint, call func() error) error {
	var err error
	for i := uint(0); i < maxCount; i++ {
		err = call()
		if err == nil {
			return nil
		}
		log.Print("retry error", err)
		time.Sleep(pause)
	}

	return err
}

func main() {
	port := flag.String("port", "8080", "port")
	dbInfo := flag.String("db", "", "db connections credentials")
	level := flag.String("loglevel", "info", "log level")
	flag.Parse()

	fmt.Println(*port, *dbInfo, *level)

	db, err := sql.Open("postgres", *dbInfo)
	if err != nil {
		log.Printf("fatal error %s", err)

		return
	}
	defer db.Close()
	err = retry(1*time.Second, 100, db.Ping)
	if err != nil {
		log.Printf("fatal error %s", err)

		return
	}

	repo := transfers.NewRepository(db)
	service := transfers.NewService(repo)
	endpoints := transfers.NewEndpoints(service)
	httpHandler := transfers.NewHTTPHandler(endpoints)

	fmt.Printf("Starting server at port %s\n", *port)
	if err := http.ListenAndServe(":"+*port, httpHandler); err != nil {
		log.Printf("fatal error %s", err)
	}
}
