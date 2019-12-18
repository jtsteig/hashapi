package main

import "database/sql"

// CreateHashRepoWithClean creates a new repository and clears the storage, this is intended for testing.
func CreateHashRepoWithClean(db *sql.DB, hashTableName string) (*HashRepository, error) {
	hashStore := HashRepository{
		db:            db,
		hashTableName: hashTableName,
	}
	hashStore.ClearStore()
	err := hashStore.InitTables()
	if err != nil {
		return &HashRepository{}, err
	}
	return &hashStore, nil
}
