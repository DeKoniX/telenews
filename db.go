package main

import (
	"database/sql"
	"time"

	"crypto/md5"
	"fmt"
	"io"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func dbInit() (dataBase DB, err error) {
	dataBase.db, err = sql.Open("sqlite3", "./telenews.sqlite")
	if err != nil {
		return dataBase, err
	}

	sqlStmt := `
  CREATE TABLE news (
    id  INTEGER NOT NULL PRIMARY KEY,
    name STRING,
    hash STRING NOT NULL UNIQUE,
    date DATE NOT NULL
  );
  CREATE INDEX hash_index ON news(hash);
  CREATE TABLE users (
    id INTEGER NOT NULL PRIMARY KEY,
    username STRING NOT NULL,
    chat_id INTEGER NOT NULL UNIQUE
  );
`

	_, _ = dataBase.db.Exec(sqlStmt)

	return dataBase, nil
}

func (dataBase *DB) InsertInfo(idNews, name string, date time.Time) (id int64, err error) {
	timeNow := time.Now()
	hash := GenHash(idNews, name, date)

	tx, err := dataBase.db.Begin()
	if err != nil {
		return id, err
	}

	stmt, err := tx.Prepare("INSERT INTO news(name, hash, date) values(?, ?, ?)")
	if err != nil {
		return id, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(name, hash, timeNow)
	if err != nil {
		return id, err
	}
	tx.Commit()

	id, err = result.LastInsertId()
	if err != nil {
		return id, err
	}

	return id, nil
}

type RowsInfo struct {
	Id          int
	Name        string
	Hash        string
	Url         string
	Description string
	Date        time.Time
}

func (dataBase *DB) SelectInfo(hash string) (select_rows []RowsInfo, err error) {
	rows, err := dataBase.db.Query("SELECT * FROM news WHERE hash=?", hash)
	if err != nil {
		return select_rows, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, hash string
		var date time.Time
		err = rows.Scan(&id, &name, &hash, &date)
		if err != nil {
			return select_rows, err
		}

		select_rows = append(select_rows, RowsInfo{
			Id:   id,
			Name: name,
			Hash: hash,
			Date: date,
		})
	}

	return select_rows, nil
}

func (dataBase *DB) InsertUser(username string, chatID int64) (id int64, err error) {
	tx, err := dataBase.db.Begin()
	if err != nil {
		return id, err
	}

	stmt, err := tx.Prepare("INSERT INTO users(username, chat_id) values(?, ?)")
	if err != nil {
		return id, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(username, chatID)
	if err != nil {
		return id, err
	}
	tx.Commit()

	id, err = result.LastInsertId()
	if err != nil {
		return id, err
	}

	return id, nil
}

type RowsUser struct {
	Id       int
	Username string
	ChatID   int64
}

func (dataBase *DB) SelectUsers() (select_rows []RowsUser, err error) {
	rows, err := dataBase.db.Query("SELECT * FROM users")
	if err != nil {
		return select_rows, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username string
		var chatID int64
		err = rows.Scan(&id, &username, &chatID)
		if err != nil {
			return select_rows, err
		}

		select_rows = append(select_rows, RowsUser{
			Id:       id,
			Username: username,
			ChatID:   chatID,
		})
	}

	return select_rows, nil
}

func (dataBase *DB) DeleteUser(chatID int64) (err error) {
	_, err = dataBase.db.Exec("DELETE FROM users WHERE chat_id=?", chatID)
	return err
}

//ToDo: Удаление по таймингу

func GenHash(id, name string, date time.Time) (hash string) {
	h := md5.New()
	if id == "" {
		io.WriteString(h, name)
		io.WriteString(h, date.String())
	} else {
		io.WriteString(h, id)
	}
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return hash
}
