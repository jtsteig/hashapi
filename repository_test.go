package main

import (
	"database/sql"
	"testing"
)

func TestSqlIteHashStoreHappyPath(t *testing.T) {
	filename := "c:\\temp\\testdb.db"
	hashTable := "hashes"
	db, _ := sql.Open("sqlite3", filename)
	hashStore, initErr := CreateHashRepoWithClean(db, hashTable)

	if initErr != nil {
		t.Errorf("Failed to init db: %q", initErr)
	}

	countID, createError := hashStore.CreateEmptyHashEntry()
	if createError != nil {
		t.Errorf("Failed to create hash entry: %q", createError)
	}
	expected := HashStat{"testHash1", countID, 444}
	if storeError := hashStore.UpdateHashWithValues(expected.CountID, expected.HashValue, expected.HashTimeInMilliseconds); storeError != nil {
		t.Errorf("Error updating the hash values after creation: %q", storeError)
	}

	result, hashErr := hashStore.GetHashStat(countID)
	if hashErr != nil {
		t.Errorf("Error getting hashStats: %q", hashErr)
	}

	if expected.HashValue != result.HashValue {
		t.Errorf("Got incorrect hash value: %q and expected %q", expected.HashValue, result.HashValue)
	}
	if expected.CountID != result.CountID {
		t.Errorf("Got incorrect countId: %q and expected %q", expected.CountID, result.CountID)
	}
	if expected.HashTimeInMilliseconds != result.HashTimeInMilliseconds {
		t.Errorf("Got incorrect hashtime value: %q and expected %q", expected.HashTimeInMilliseconds, result.HashTimeInMilliseconds)
	}

	hashStore.CreateEmptyHashEntry()
	hashStore.CreateEmptyHashEntry()
	hashStore.CreateEmptyHashEntry()
	hashStore.CreateEmptyHashEntry()

	totalResults, totalErr := hashStore.GetTotalStats()
	if totalErr != nil {
		t.Errorf("Error getting totalStats: %q", totalErr)
	}
	if totalResults.Count != 5 {
		t.Errorf("Error getting the totalCount. Expected %d but got %d", 5, totalResults.Count)
	}

	dropErr := hashStore.ClearStore()
	if dropErr != nil {
		t.Errorf("Failed to drop table %q", dropErr)
	}
	closeErr := hashStore.Close()
	if closeErr != nil {
		t.Errorf("Failed to close db: %q", closeErr)
	}
}
