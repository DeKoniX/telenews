package models

import (
	"testing"
)

func TestItem(t *testing.T) {
	connectDB()

	// testUser: username: TestUser, chatID: 123
	user, err := createUser("TestUser", 123)
	if err != nil {
		t.Error("[ERR] Create User: ", user, err)
	}

	// testSource: url: test_url_1, Twitter
	appSource1, err := appendSource(user, "test_url_1", Twitter)
	if err != nil {
		t.Error("[ERR] Append Source 1: ", user, err, appSource1)
	}

	// testSource: url: test_url_2, VKWall
	appSource2, err := appendSource(user, "test_url_2", VKWall)
	if err != nil {
		t.Error("[ERR] Append Source 2: ", user, err, appSource2)
	}

	appItem1, err := appendItem(appSource1, "Title 1", "Text 1")
	if err != nil {
		t.Error("[ERR] Append Item source 1: ", appSource1, appItem1, err)
	}

	appItem2, err := appendItem(appSource2, "Title 1", "Text 1")
	if err != nil {
		t.Error("[ERR] Append Item source 2: ", appSource2, appItem2, err)
	}

	appExistItem, err := appendItem(appSource1, "Title 1", "Text 1")
	if err != ErrAlreadyExists {
		t.Error("[ERR] Append exist Item: ", appSource1, appExistItem, err)
	}

	appItems, count, err := Item{}.Select(appSource1)
	if err != nil {
		t.Error("[ERR] SelectByUser Items: ", appItems, count, err)
	}
	if count != 1 {
		t.Error("[WAR] not the right amount Item: ", appItems, count)
	}

	// ---------------------------------
	err = appSource1.Delete()
	if err != nil {
		t.Error("[ERR] Delete Source 1: ", appSource1, err)
	}

	err = appSource2.Delete()
	if err != nil {
		t.Error("[ERR] Delete Source 2: ", appSource2, err)
	}

	err = user.Delete()
	if err != nil {
		t.Error("[ERR] Delete User: ", user, err)
	}

	err = appItem1.Delete()
	if err != nil {
		t.Error("[ERR] Delete Item 1: ", user, err)
	}
}

func appendItem(source Source, title, text string) (item Item, err error) {
	item = Item{Title: title, Text: text}
	_, err = item.Insert(source)

	return item, err
}
