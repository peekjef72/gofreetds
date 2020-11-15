package main

import (
	"database/sql"
	"fmt"
	//	"github.com/peekjef72/gofreetds"
	//	"os"
)

//Example how to use gofreetds as mssql driver for standard sql interface.
//More information on how to use sql interface:
//  http://golang.org/pkg/database/sql/
//  https://code.google.com/p/go-wiki/wiki/SQLInterface
func main() {
	//get connection string
	//connStr := os.Getenv("GOFREETDS_CONN_STR")
	//connStr := "Server=TBSM_VAL;User Id=root;Password=root;compatibility=sybase;"
	// connStr := "Server=mssql;User Id=SA;Password=gNZPA3vly4Yh2wL;Database=alerts;compatibility=openserver;"
	connStr := "host=localhost:1433;User Id=SA;Password=gNZPA3vly4Yh2wL;Database=alerts;compatibility=openserver;"
	if connStr == "" {
		panic("Set connection string for the pubs database in GOFREETDS_CONN_STR environment variable!\n")
	}

	//get db connection
	db, err := sql.Open("mssql", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//use it
	//row := db.QueryRow("SELECT au_fname, au_lname name FROM authors WHERE au_id = ?", "172-32-1176")
	//var firstName, lastName string
	//err = row.Scan(&firstName, &lastName)
	sql := "SELECT count(*) as num from status"
	row := db.QueryRow(sql)
	fmt.Printf("query: %s\n", sql)
	var num int
	err = row.Scan(&num)
	if err != nil {
		panic(err)
	}

	//show results
	//fmt.Printf("author for id: %s is %s %s\n", auId, firstName, lastName)
	fmt.Printf("count: %d\n", num)
}
