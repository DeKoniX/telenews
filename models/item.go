package models

import (
	"github.com/jinzhu/gorm"
)

type Item struct {
	gorm.Model
	Title    string
	Text     string
	Link     string
	Hash     string `gorm:"unique;not null"`
	SourceID uint
}

func (item *Item) Insert(source Source) (_ int, _ error) {
	a := db.Model(&source).Association("Items")

	itemTest := Item{}
	itemTest.SelectByHash(item.Hash)
	if itemTest.Hash == item.Hash {
		return a.Count(), ErrAlreadyExists
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
