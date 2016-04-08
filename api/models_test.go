package api

import (
	"os"
	"testing"
)

func init() {
	os.Setenv("GOENV", "test")
}

func TestSaveUserToDatabase(t *testing.T) {
	db := GetDB()
	var testingTable = []struct {
		user          User
		ExpectedError error
		Reason        string
	}{
		{User{FirstName: "testuser1",
			LastName:     "testuser1",
			Password:     "blasphemy",
			PrimaryEmail: "test@test.com"}, nil, "Everything should be okay with this user"},
		{User{FirstName: "testuser1",
			LastName:     "testuser1",
			Password:     "blasphemy",
			PrimaryEmail: "test@test.com"}, UserExistsError{"Unable to save user with PrimaryEmail test@test.com already exists in database"}, "This is a duplicate user to the last and therefore should not be saved"},
	}

	for _, testCase := range testingTable {
		err := testCase.user.Save(db)
		if err != testCase.ExpectedError {
			t.Errorf("Expected %s but got %s reason %s", testCase.ExpectedError, err, testCase.Reason)
		}
	}
}
