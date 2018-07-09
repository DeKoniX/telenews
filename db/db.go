package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB

func Init(address, username, password, dbname string) (err error) {
	pgurl := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", username, password, address, dbname)
	fmt.Println(pgurl)

	db, err = gorm.Open("postgres", pgurl)
	if err != nil {
		return err
	}
	db.LogMode(true) //LOGMODE

	db.AutoMigrate(&User{}, &Favorite{}, &Item{})

	return nil
}
