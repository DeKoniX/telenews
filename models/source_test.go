package models

import "testing"

func TestSource(t *testing.T) {
	connectDB()

	// testUser: username: TestUser, chatID: 123
	user, err := createUser("TestUser", 123)
	if err != nil {
		t.Error("[ERR] Create User: ", user, err)
	}

	// testUser: username: TestUser2, chatID: 1234
	user2, err := createUser("TestUser2", 1234)
	if err != nil {
		t.Error("[ERR] Create User2: ", user, err)
	}

	// testSource: url: test_url_1, Twitter
	appSource1, err := appendSource(user, "test_url_1", Twitter)
	if err != nil {
		t.Error("[ERR] Append Source user 1: ", user, err, appSource1)
	}

	// testSource: url: test_url_2, VKWall
	appSource2, err := appendSource(user2, "test_url_2", VKWall)
	if err != nil {
		t.Error("[ERR] Append Source user 2: ", user2, err, appSource2)
	}

	sources, err := Source{}.Select(user)
	if err != nil {
		t.Error("[ERR] Get Sources user: ", user, err, sources)
	}
	if sources[0].Query != appSource1.Query && sources[0].Type != appSource1.Type && err != nil {
		t.Error("[ERR] Do not match sources: ", user, err, sources, appSource1)
	}

	sources2, err := Source{}.SelectByType(user, Twitter)
	if sources2[0].Query == appSource1.Query && sources2[0].Type == appSource1.Type && err != nil {
		t.Error("[ERR] Do not match sources by type: ", user2, err, sources2, appSource1)
	}

	testSource := Source{}
	err = testSource.SelectByQueryAndType(user2, appSource2.Query, appSource2.Type)
	if testSource.Query != appSource2.Query && testSource.Type != appSource2.Type && err != nil {
		t.Error("[ERR] Do not match sources by url and type: ", user2, err, testSource, appSource2)
	}

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

	err = user2.Delete()
	if err != nil {
		t.Error("[ERR] Delete User 2: ", user, err)
	}
}

func appendSource(user User, url string, sourceType SourceType) (source Source, err error) {
	source = Source{Query: url, Type: sourceType}
	_, err = source.Insert(user)

	return source, err
}
