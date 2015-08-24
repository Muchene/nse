package main

import (
	"database/sql"
	"fmt"
	"bufio"
	_ "github.com/lib/pq"
	"nse/pricescraper"
	"time"
	"os"
	"strings"
)

type Company pricescraper.Company

func initCompaniesTable(prices map[string]pricescraper.Company, db *sql.DB) {
	sStmt := "INSERT INTO companies (company_id,company_name,type) VALUES  ($1, $2, $3)"
	stmt, err := db.Prepare(sStmt)
	if err != nil {
		panic(err)
	}
	id := 0
	for _, company := range prices {
		res, err := stmt.Exec(id, company.Name, company.CType)
		id = id + 1
		if res == nil || err != nil {
			panic(err)
		}
	}
	stmt.Close()
}

func insertPrices(companyRows *sql.Rows, prices map[string]pricescraper.Company, db *sql.DB) {
	sStmt := "INSERT INTO price (company_id, time, price) VALUES ($1, $2, $3)"
	stmt, err := db.Prepare(sStmt)
	defer stmt.Close()
	if err != nil {
		fmt.Println("Error insertPrices error preparing insert statement")
		panic(err)
	}
	for companyRows.Next() {
		var id uint64
		var name string
		var cType string
		companyRows.Scan(&id, &name, &cType)
		res, err := stmt.Exec(id, time.Now(), prices[name].LastPrice)
		if res == nil || err != nil {
			fmt.Printf("Error inserting price %s\n", name)
		}
	}
}


func checkAndPanic(err error){
	if err != nil{
		panic(err)
	}
}

func parseSettings(fname string) map[string]string{
	sett := make(map[string]string)
	dat,err := os.Open(fname)
	defer dat.Close()
	checkAndPanic(err)
	s := bufio.NewScanner(dat)
	for s.Scan(){
		line := s.Text()
		lineEls := strings.Split(line, "=")
		if(len(lineEls) < 2){
			continue
		}
		sett[lineEls[0]] = lineEls[1]
	}

	return sett
}

func main() {
 	settings := map[string] string{
 				"user":"postgres",
 				"dbname":"stocks",
 				"password": "123456",
 				"sslmode": "disable",
 				"host": "localhost",}

	if(len(os.Args)>1){
		user_settings := parseSettings(os.Args[1])
		settings = user_settings
	}
		
	connectionStr := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s", 
		settings["host"], settings["user"], settings["dbname"], 
		settings["password"], settings["sslmode"])
	
	db, err := sql.Open("postgres", connectionStr)
	checkAndPanic(err)
	defer db.Close()

	//Maybe the first time this database was initialized so check the number of companies exist...
	rows, err := db.Query("SELECT COUNT(company_id) FROM companies")
	rows.Next()
	var numRows int
	rows.Scan(&numRows)
	checkAndPanic(err)

	//scrape the page for the prices (and the names...)
	prices := pricescraper.GrabData()
	if numRows == 0 {
		initCompaniesTable(prices, db)
	}
	rows.Close()

	rows, err = db.Query("SELECT * FROM companies")
	checkAndPanic(err)
	insertPrices(rows, prices, db)
	rows.Close()
}
