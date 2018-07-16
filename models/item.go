package models

import (
	"crypto/md5"
	"fmt"
	"io"

	"strconv"

	"github.com/jinzhu/gorm"
)

type Item struct {
	gorm.Model
	Title    string
	Text     string
	Hash     string `gorm:"unique;not null"`
	SourceID uint
}

func (item *Item) Insert(source Source) (_ int, _ error) {
	a := db.Model(&source).Association("Items")

	item.Hash = genHash(source.UserID, item.Title, item.Text)

	itemTest := Item{}
	itemTest.SelectByHash(item.Hash)
	if itemTest.Hash == item.Hash {
		return a.Count(), alreadyExists
	} else {
		if err := a.Append(item).Error; err != nil {
			return 0, err
		}
	}
	return a.Count(), nil
}

func (item Item) Select(source Source) (items []Item, _ int, _ error) {
	a := db.Model(source).Association("Items")
	if a.Count() == 0 {
		return items, 0, nil
	}
	if err := a.Find(&items).Error; err != nil {
		return items, 0, err
	}
	return items, a.Count(), nil
}

func (item *Item) SelectByHash(hash string) (err error) {
	err = db.Where("hash = ?", hash).First(&item).Error
	return err
}

func (item Item) Delete() error {
	if err := db.Delete(item).Error; err != nil {
		return err
	}
	return nil
}

func genHash(userID uint, title, text string) (hash string) {
	h := md5.New()
	io.WriteString(h, strconv.Itoa(int(userID)))
	io.WriteString(h, title)
	io.WriteString(h, text)
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return hash
}
