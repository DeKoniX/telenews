package models

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	err := connectDB()
	if err != nil {
		t.Error("[ERR] DB Connect: ", err)
	}
}

func connectDB() error {
	var envDBAddress = os.Getenv("DBADDRESS")
	var envDBName = os.Getenv("DBNAME")
	var envDBUser = os.Getenv("DBUSER")
	var envDBPass = os.Getenv("DBPASS")

	err := Init(envDBAddress, envDBUser, envDBPass, envDBName)
	db.LogMode(true) //LOGMODE
	return err
}
