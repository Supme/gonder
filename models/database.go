package models

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mcuadros/go-version"
	"gonder/bindata"
	"strings"
)

func ConnectDb() error {
	var err error
	Db, err = sql.Open("mysql", Config.dbString)
	if err != nil {
		return fmt.Errorf("open database error: %s", err)
	}
	err = Db.Ping()
	if err != nil {
		return fmt.Errorf("ping database error: %s", err)
	}
	Db.SetMaxIdleConns(Config.dbConnections)
	Db.SetMaxOpenConns(Config.dbConnections)
	return nil
}

func CheckDb() error {
	if _, err := Db.Exec("SELECT 1 FROM `auth_user`"); err != nil {
		return fmt.Errorf("database is empty, use -i key for create from a template")
	}
	// with version 0.16.0 there was a table of versions
	_, err := Db.Exec("SELECT 1 FROM `version`")
	if err != nil {
		err = dbQueryFrom("sql/update/0.16.0.sql")
		if err != nil {
			return err
		}
	}
	row := Db.QueryRow("SELECT `number` FROM `version` ORDER BY `at` DESC LIMIT 1 ")
	var dbVersion string
	err = row.Scan(&dbVersion)
	if err != nil {
		return fmt.Errorf("get database version: %s", err)
	}

	// in the future, here, if necessary, will check the version of the application and the database

	// update the version in the database when it changes
	if version.Compare(Version, dbVersion, ">") {
		if _, err := Db.Exec("INSERT INTO `version` (`number`) VALUES (?)", Version); err != nil {
			return fmt.Errorf("insert new version to database: %s", err)
		}
	}

	return nil
}

// InitDb initialize database
func InitDb(withoutConfirm bool) error {
	if !withoutConfirm {
		var confirm string
		fmt.Print("Initial database (y/N)? ")
		if _, err := fmt.Scanln(&confirm); err != nil {
			return err
		}
		if strings.ToLower(confirm) != "y" {
			return nil
		}
	}

	return dbQueryFrom("sql/dump.sql")
}

func dbQueryFrom(filename string) error {
	sqlDump, err := bindata.ReadFileOrAsset(filename)
	if err != nil {
		return err
	}
	query := strings.Split(string(sqlDump), ";")
	tx, err := Db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	for i := range query {
		if strings.TrimSpace(query[i]) == "" {
			continue
		}
		_, err = tx.Exec(query[i])
		if err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
