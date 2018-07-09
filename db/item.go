package db

import "github.com/jinzhu/gorm"

type Item struct {
	gorm.Model
	Title      string
	Text       string
	Hash       string `gorm:"unique;not null"`
	FavoriteID uint
}

func (item Item) Insert(favorite *Favorite) (_ int, _ error) {
	a := db.Model(favorite).Association("Items")
	if err := a.Append(item).Error; err != nil {
		return 0, err
	}
	return a.Count(), nil
}

func (item Item) Select(favorite *Favorite) (items []Item, _ int, _ error) {
	a := db.Model(favorite).Association("Favorites")
	if a.Count() == 0 {
		return items, 0, nil
	}
	if err := a.Find(&items).Error; err != nil {
		return items, 0, err
	}
	return items, a.Count(), nil
}

func (item Item) Delete() error {
	if err := db.Delete(item).Error; err != nil {
		return err
	}
	return nil
}
