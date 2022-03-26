package db

import (
	"database/sql"
	"github.com/Nerinyan/Nerinyan-APIV2/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pterm/pterm"
)

var Maria1 *sql.DB
var Maria2 *sql.DB

func ConnectMaria() {

	db, err := sql.Open("mysql", config.Config.Sql.Url)
	if Maria1 = db; db != nil && err == nil {
		Maria1.SetMaxOpenConns(100)
		pterm.Success.Println("RDBMS1 connected")

		if _, err = Maria1.Exec("SET SQL_SAFE_UPDATES = 0;"); err != nil {
			pterm.Error.Println("RDBMS1: SET SQL_SAFE_UPDATES FAIL.")
			panic(err)
		}
		//pterm.Success.Println("RDBMS Connected.")
	} else {
		pterm.Error.Println("RDBMS1 Connect Fail", err)
		panic(err)
	}

	if config.Config.Sql2.Url == "" {
		return
	}
	db2, err := sql.Open("mysql", config.Config.Sql2.Url)
	if Maria2 = db2; db2 != nil && err == nil {
		Maria2.SetMaxOpenConns(100)
		pterm.Success.Println("RDBMS2 connected")

		if _, err = Maria2.Exec("SET SQL_SAFE_UPDATES = 0;"); err != nil {
			pterm.Error.Println("RDBMS2: SET SQL_SAFE_UPDATES FAIL.")
			panic(err)
		}
		//pterm.Success.Println("RDBMS Connected.")
	} else {
		pterm.Error.Println("RDBMS2 Connect Fail", err)
		panic(err)
	}
}
