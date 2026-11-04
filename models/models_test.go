package models

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	connectDB(t)
}

func connectDB(t *testing.T) {
	t.Helper()
	address, name, user := os.Getenv("DBADDRESS"), os.Getenv("DBNAME"), os.Getenv("DBUSER")
	if address == "" || name == "" || user == "" {
		t.Skip("set DBADDRESS, DBNAME and DBUSER to run PostgreSQL integration tests")
	}
	if err := Init(address, user, os.Getenv("DBPASS"), name); err != nil {
		t.Fatalf("connect to test PostgreSQL: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test PostgreSQL: %v", err)
		}
	})
}
