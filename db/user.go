package db

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	UserName  string
	ChatID    int64 `gorm:"unique;not null"`
	Favorites []Favorite
}

func (user User) Insert() error {
	b := db.NewRecord(user) // обновляю, или создаю нового пользователя
	if b == true {
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	} else {
		if err := db.Save(&user).Error; err != nil {
			return err
		}
	}
	return nil
}

func (user *User) Delete() error {
	if err := db.Delete(user).Error; err != nil {
		return err
	}
	return nil
}
