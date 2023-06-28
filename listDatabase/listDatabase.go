package listdatabase

import (
	"bytes"
	"chat/database"
	"encoding/binary"
	"sync"
)

type ListDb struct {
	head, list database.DataBase
	lock       sync.Mutex
}

func (db *ListDb) Open(path string) error {
	err := db.head.Open(path + "/head.db")
	if err != nil {
		return err
	}
	err = db.list.Open(path + "/list.db")
	if err != nil {
		db.head.Close()
		return err
	}
	if flag, _ := db.list.Has([]byte("counter")); !flag {
		db.list.Write([]byte("counter"), []byte{0, 0, 0, 0}[:4])
	}
	return nil
}

func (db *ListDb) Close() {
	db.head.Close()
	db.list.Close()
}

func (db *ListDb) Query(key string) []string {
	var result = []string{}
	if flag, err := db.head.Has([]byte(key)); !flag || err != nil {
		result = append(result, "No message")
		return result
	}
	pos := []byte(db.head.Get([]byte(key)))[:4]
	for pos[0] != 0 || pos[1] != 0 || pos[2] != 0 || pos[3] != 0 {
		content := db.list.Get(pos)
		result = append(result, content[4:])
		pos = []byte(content)[:4]
	}
	return result
}

func (db *ListDb) Insert(key string, value string) error {
	flag, err := db.head.Has([]byte(key))
	if err != nil {
		return err
	}
	if !flag {
		db.head.Write([]byte(key), []byte{0, 0, 0, 0}[:4])
	}
	var count int32
	db.lock.Lock()
	ByteBuffer := bytes.NewBuffer([]byte(db.list.Get([]byte("counter")))[:4])
	binary.Read(ByteBuffer, binary.BigEndian, &count)
	count = count + 1
	//log.Println(count)
	binary.Write(ByteBuffer, binary.BigEndian, count)
	db.list.Modify([]byte("counter"), ByteBuffer.Bytes())
	content := []byte(db.head.Get([]byte(key)) + value)
	db.head.Modify([]byte(key), ByteBuffer.Bytes())
	db.list.Write(ByteBuffer.Bytes(), content)
	db.lock.Unlock()
	return nil
}
