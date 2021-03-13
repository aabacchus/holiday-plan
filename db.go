package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const sqlHostname = "127.0.0.1:3306"

func sqlDBConnection(username, password, dbName string) (*sql.DB, error) {
	dbPath := fmt.Sprintf("%s:%s@tcp(%s)/", username, password, sqlHostname)
	db, err := sql.Open("mysql", dbPath)
	if err != nil {
		return nil, err
	}
	//defer db.Close()

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+dbName)
	if err != nil {
		return nil, err
	}
	_, err = res.RowsAffected()
	if err != nil {
		return nil, err
	}
	db.Close()

	dbPath = fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, sqlHostname, dbName)
	db, err = sql.Open("mysql", dbPath)
	if err != nil {
		return nil, err
	}
	//defer db.Close()

	// max_connections overall is 151
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 2)

	ctx, cancelfunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Errors %s pinging DB", err)
		return db, err
	}

	return db, nil
}

// sqlWriteToDB writes a Markers into tablename in db and returns the number of rows affected
// and an error.
func sqlWriteToDB(db *sql.DB, tableName string, m Markers) (int64, error) {
	// use a prepared SQL statement
	query := "INSERT INTO " + tableName + "(name, latit, longit) VALUES"
	var queryFields []string
	var params []interface{}
	for _, mark := range m.Markers {
		queryFields = append(queryFields, "(?, ?, ?)")
		params = append(params, mark.Name, mark.Lat, mark.Long)
	}
	queryVals := strings.Join(queryFields, ",")
	query = query + queryVals

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, params...)
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func sqlQueryTable(db *sql.DB, tableName string) (Markers, error) {
	var m Markers
	rows, err := db.Query("SELECT name, latit, longit FROM waterfalls LIMIT 1")
	if err != nil {
		return Markers{}, err
	}
	defer rows.Close()

	// go through each row
	for rows.Next() {
		var mark Marker
		err := rows.Scan(&mark.Name, &mark.Lat, &mark.Long)
		if err != nil {
			return m, err
		}
		m.Markers = append(m.Markers, mark)
	}
	err = rows.Err()
	return m, err
}

func sqlCreateTable(tablename string, db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS ` + tablename + `(
			id INT PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(50) NOT NULL,
			latit DOUBLE(16,6) NOT NULL,
			longit DOUBLE(16,6) NOT NULL)`
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	res, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if *verbose {
		log.Printf("Rows affected when creating table: %d", rows)
	}
	return nil
}
