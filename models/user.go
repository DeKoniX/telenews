package models

import (
	"time"
)

type User struct {
	ID        uint `gorm:"primary_key"`
	UserName  string
	ChatID    int64 `gorm:"unique;not null"`
	Sources   []Source
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user *User) Insert() error {
	b := db.NewRecord(user)
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

func (user *User) SelectByChatId(chatID int64) error {
	err := db.Where("chat_id = ?", chatID).First(&user).Error
	return err
}

func (user *User) SelectById(id uint) error {
	err := db.First(&user, id).Error
	return err
}

func (user *User) Delete() error {
	var sources []Source

	a := db.Model(&user).Association("Sources")
	if err := a.Find(&sources).Error; err != nil {
		return err
	}
	for _, source := range sources {
		if err := source.Delete(); err != nil {
			return err
		}
	}

	if err := db.Delete(user).Error; err != nil {
		return err
	}
	return nil
}
