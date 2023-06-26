package database

import (
	"log"
	"os"
	"strings"

	badger "github.com/dgraph-io/badger/v3"
)

type DataBase struct {
	db *badger.DB
}

func (db *DataBase) Open(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0754)
	}
	opts := badger.DefaultOptions(path)
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false
	opts.ValueThreshold = 255
	opts.CompactL0OnClose = true
	var err error
	db.db, err = badger.Open(opts)
	if err != nil {
		log.Println("badger open failed", "path", path, "err", err)
		return err
	}
	return nil
}

func (db *DataBase) Close() {
	err := db.db.Close()
	if err == nil {
		log.Println("database closed", "err", err)
	} else {
		log.Println("failed to close database", "err", err)
	}
}

func (db *DataBase) Write(key []byte, value []byte) error {
	wb := db.db.NewWriteBatch()
	defer wb.Cancel()
	err := wb.SetEntry(badger.NewEntry(key, value).WithMeta(0))
	if err != nil {
		log.Println("Failed to write data to cache.", "key", string(key), "value", string(value), "err", err)
	}
	err = wb.Flush()
	if err != nil {
		log.Println("Failed to flush data to cache.", "key", string(key), "value", string(value), "err", err)
	}
	return err
}

func (db *DataBase) Get(key []byte) string {
	var ival []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		ival, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Println("Failed to read data from the cache.", "key", string(key), "error", err)
	}
	return string(ival)
}

func (db *DataBase) Has(key []byte) (bool, error) {
	var exist bool = false
	err := db.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return err
		} else {
			exist = true
		}
		return err
	})
	// align with leveldb, if the key doesn't exist, leveldb returns nil
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			err = nil
		}
	}
	return exist, err
}

func (db *DataBase) Delete(key []byte) error {
	wb := db.db.NewWriteBatch()
	defer wb.Cancel()
	return wb.Delete(key)
}

func (db *DataBase) Modify(key []byte, value []byte) error {
	exist, err := db.Has(key)
	if err != nil {
		return err
	}
	if exist {
		err = db.Delete(key)
		if err != nil {
			return err
		}
	}
	err = db.Write(key, value)
	return err
}
