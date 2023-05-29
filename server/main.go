package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

type Response struct {
	Rates map[string]float64
}

type Result struct {
	Currency string
	Rate     float64
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPassword, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	apiKey := os.Getenv("OPEN_EXCHANGE_RATES_API_KEY")
	currencies := []string{"JPY", "EUR", "GBP"}

	sem := make(chan struct{}, 2)
	for {
		results := make(chan Result, len(currencies))
		var wg sync.WaitGroup

		for _, currency := range currencies {
			wg.Add(1)
			go func(currency string) {
				defer wg.Done()
				resp, err := http.Get(fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s&symbols=%s", apiKey, currency))
				if err != nil {
					log.Println(err)
					return
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					resp.Body.Close()
					return
				}
				resp.Body.Close()

				var data Response
				err = json.Unmarshal(body, &data)
				if err != nil {
					log.Println(err)
					return
				}

				results <- Result{Currency: currency, Rate: data.Rates[currency]}
			}(currency)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for result := range results {
			wg.Add(1)
			go func(result Result) {
				sem <- struct{}{}
				defer func() {
					<-sem
					wg.Done()
				}()

				_, err = db.Exec("INSERT INTO rates (date, currency, rate) VALUES (?, ?, ?)", time.Now(), result.Currency, result.Rate)
				if err != nil {
					log.Println(err)
				}
				fmt.Printf("1 USD is equal to %.2f %s\n", result.Rate, result.Currency)
			}(result)
		}

		wg.Wait()
		time.Sleep(1 * time.Hour)
	}
}
