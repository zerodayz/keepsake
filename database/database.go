package database

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)
func main() {
	db, err := sql.Open("mysql", "gowiki:gowiki55@/gowiki")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		internal_id int,
		name varchar(50),
		username varchar(15),
		email varchar(255),
		password varchar(60)
		);`)
	if err != nil {
		panic(err.Error())
	}
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tokens (
		token_id int,
		token blob,
		username varchar(15),
		expires timestamp()
		);`)
	if err != nil {
		panic(err.Error())
	}
}