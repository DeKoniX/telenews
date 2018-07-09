package db

import (
	"fmt"
	"os"
	"testing"
)

// TODO: соеденение, добавление пользователя, добавление к пользователю избранного, добавление к избранномму несколько статей, проверка
func TestInit(t *testing.T) {
	var err error

	var envDBAddress = os.Getenv("DBADDRESS")
	var envDBName = os.Getenv("DBNAME")
	var envDBUser = os.Getenv("DBUSER")
	var envDBPass = os.Getenv("DBPASS")

	err = connectDB(envDBAddress, envDBUser, envDBPass, envDBName)
	if err != nil {
		t.Error("[ERR] Connect DB: ", err)
	}

	// при добавлении пользователя структура пользователя не меняется - для работы ассоциаций менять надо !
	user := User{UserName: "test", ChatID: 634}
	db.First(&user)
	//err = user.Insert()
	//if err != nil {
	//	t.Error("[ERR] Create User: ", err)
	//}

	favorit := Favorite{
		Type: Twitter,
		URL:  "https://test_url2.tg",
	}
	favorit.Insert(&user)
	fav, _, err := Favorite{}.Select(&user)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(">>", fav)
	//for _, fav := range fav {
	//	fmt.Println(fav)
	//	err := fav.Delete(&user)
	//	if err != nil {
	//		t.Error(err)
	//	}
	//}
}

func connectDB(address, user, pass, dbname string) (err error) {
	err = Init(address, user, pass, dbname)
	db.LogMode(true) //LOGMODE
	return err
}
