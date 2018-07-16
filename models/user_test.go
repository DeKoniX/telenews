package models

import "testing"

func TestUser(t *testing.T) {
	connectDB()

	// testUser: username: TestUser, chatID: 123
	user, err := createUser("TestUser", 123)
	if err != nil {
		t.Error("[ERR] Create User: ", user, err)
	}

	user2, err := getUser(123)
	if err != nil {
		t.Error("[ERR] Get User: ", user2, err)
	}

	err = user.Delete()
	if err != nil {
		t.Error("[ERR] Delete User: ", user, err)
	}

	user, err = getUser(123)
	if err == nil {
		t.Error("[ERR] Get Delete User: ", user, err)
	}
}

func createUser(username string, chatID int64) (user User, err error) {
	user = User{UserName: username, ChatID: chatID}
	err = user.Insert()

	return user, err
}

func getUser(chatID int64) (user User, err error) {
	err = user.SelectByChatId(chatID)

	return user, err
}
