package db

import "github.com/jinzhu/gorm"

type FavoriteType string

var (
	Twitter FavoriteType = "twitter"
	VK      FavoriteType = "vk"
	RSS     FavoriteType = "rss"
)

type Favorite struct {
	gorm.Model
	Type   FavoriteType
	URL    string
	Items  []Item
	UserID uint
}

func (favorite Favorite) Insert(user *User) (_ int, _ error) {
	a := db.Model(user).Association("Favorites")
	if err := a.Append(favorite).Error; err != nil {
		return 0, err
	}
	return a.Count(), nil
}

func (favorite Favorite) Select(user *User) (favorites []Favorite, _ int, _ error) {
	a := db.Model(user).Association("Favorites")
	if a.Count() == 0 {
		return favorites, 0, nil
	}
	if err := a.Find(&favorites).Error; err != nil {
		return favorites, 0, err
	}
	return favorites, a.Count(), nil
}

func (favorite Favorite) Delete() error {
	if err := db.Delete(favorite).Error; err != nil {
		return err
	}
	return nil
}
