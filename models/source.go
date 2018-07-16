package models

import "github.com/jinzhu/gorm"

type SourceType string

var (
	Twitter SourceType = "twitter"
	VKGroup SourceType = "vk_group"
	RSS     SourceType = "rss"
)

type Source struct {
	gorm.Model
	Type   SourceType
	URL    string
	Items  []Item
	UserID uint
}

func (source *Source) Insert(user User) (_ int, _ error) {
	a := db.Model(&user).Association("Sources")
	if err := a.Append(source).Error; err != nil {
		return 0, err
	}
	return a.Count(), nil
}

func (source Source) Select(user User) (sources []Source, _ error) {
	a := db.Model(&user).Association("Sources")
	if err := a.Find(&sources).Error; err != nil {
		return sources, err
	}
	return sources, nil
}

func (source Source) SelectByType(user User, sourceType SourceType) (sources []Source, err error) {
	err = db.Where("user_id = ? AND type = ?", user.ID, sourceType).Find(&sources).Error
	return sources, err
}

func (source *Source) SelectByURLAndType(user User, url string, sourceType SourceType) error {
	err := db.Where("user_id = ? AND url = ? AND type = ?", user.ID, url, sourceType).First(&source).Error
	return err
}

func (source Source) Delete() error {
	if err := db.Delete(source).Error; err != nil {
		return err
	}
	return nil
}
