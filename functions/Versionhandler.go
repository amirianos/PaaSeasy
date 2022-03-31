package functions

import (
	"database/sql"
	"fmt"
)

var db *sql.DB

func Versionhandler(firsttime bool, path string, lastversion string, dbuser string, dbpassword string, dbname string) {
	var err error
	db, err = sql.Open("mysql", dbuser+":"+dbpassword+"@tcp(localhost:3306)/"+dbname)
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	if firsttime {
		var insertStmt *sql.Stmt
		insertStmt, err = db.Prepare("INSERT INTO lastversion (path, lastversion) VALUES (?, ?);")
		if err != nil {
			fmt.Println("error preparing statement:", err)
			return
		}
		defer insertStmt.Close()
		var result sql.Result
		//  func (s *Stmt) Exec(args ...interface{}) (Result, error)
		result, err = insertStmt.Exec(path, lastversion)
		rowsAff, _ := result.RowsAffected()
		lastIns, _ := result.LastInsertId()
		fmt.Println("rowsAff:", rowsAff)
		fmt.Println("lastIns:", lastIns)
		fmt.Println("err:", err)
		if err != nil {
			fmt.Println("error inserting new user")
			return
		}
	} else {
		fmt.Println("not now")

	}
}
