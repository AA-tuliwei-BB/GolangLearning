package main

// login request: localhost:port?opt=login&username=$username
// return: "success!" or "failed!"

// logout request: localhost:port?opt=logout&username=$username
// return: "success!" or "failed!"

// send message request: localhost:port?opt=send&username=$username&message=$message
// return: "success!" or "failed!"

// query request: localhost:port?opt=query
// return "Error!" or $message

// database: username-time-message

import (
	"bytes"
	"chat/database"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
)

var LoginDb, HeaderDb, MessageDb database.DataBase

func Open(path string) error {
	err := LoginDb.Open(path + "/Login.db")
	if err != nil {
		return err
	}
	err = HeaderDb.Open(path + "/Header.db")
	if err != nil {
		LoginDb.Close()
		return err
	}
	err = MessageDb.Open(path + "/Message.db")
	if err != nil {
		LoginDb.Close()
		HeaderDb.Close()
		return err
	}
	if flag, _ := MessageDb.Has([]byte("counter")); !flag {
		MessageDb.Write([]byte("counter"), []byte{0, 0, 0, 0}[:4])
	}
	return nil
}

func Close() {
	LoginDb.Close()
	HeaderDb.Close()
	MessageDb.Close()
}

func CheckLogin(username string) bool {
	if flag, err := LoginDb.Has([]byte(username)); !flag || err != nil {
		if err != nil {
			log.Println("error", err)
		}
		return false
	}
	if LoginDb.Get([]byte(username)) == "0" {
		return false
	} else {
		return true
	}
}

func UserLogin(username string) string {
	check := CheckLogin(username)
	if check {
		return "User already logged in"
	} else {
		err := LoginDb.Modify([]byte(username), []byte("1"))
		if err != nil {
			return err.Error()
		} else {
			return "Successfully log in"
		}
	}
}

func UserLogout(username string) string {
	check := CheckLogin(username)
	log.Println("test point")
	if !check {
		return "User not Login!"
	} else {
		err := LoginDb.Modify([]byte(username), []byte("0"))
		if err != nil {
			return err.Error()
		} else {
			return "Goodbye!"
		}
	}
}

func QueryMessage(username string) []string {
	var result []string
	check := CheckLogin(username)
	if !check {
		result = append(result, "User not login")
		return result
	}
	if flag, err := HeaderDb.Has([]byte(username)); !flag || err != nil {
		result = append(result, "No message")
		return result
	}
	pos := []byte(HeaderDb.Get([]byte(username)))[:4]
	for pos[0] != 0 || pos[1] != 0 || pos[2] != 0 || pos[3] != 0 {
		content := MessageDb.Get(pos)
		log.Println("pos", pos)
		log.Println("content", content)
		result = append(result, content[4:])
		pos = []byte(content)[:4]
	}
	return result
}

func SendMessage(username string, message string) string {
	check := CheckLogin(username)
	if !check {
		return "User not login"
	}
	flag, err := HeaderDb.Has([]byte(username))
	if err != nil {
		return err.Error()
	}
	if !flag {
		HeaderDb.Write([]byte(username), []byte{0, 0, 0, 0}[:4])
	}
	var count int32
	ByteBuffer := bytes.NewBuffer([]byte(MessageDb.Get([]byte("counter")))[:4])
	binary.Read(ByteBuffer, binary.BigEndian, &count)
	count = count + 1
	binary.Write(ByteBuffer, binary.BigEndian, count)
	MessageDb.Modify([]byte("counter"), ByteBuffer.Bytes())
	content := []byte(HeaderDb.Get([]byte(username)) + message)
	log.Println("content", content)
	log.Println("bytes", ByteBuffer.Bytes())
	HeaderDb.Modify([]byte(username), ByteBuffer.Bytes())
	MessageDb.Write(ByteBuffer.Bytes(), content)
	return "successfully send message"
}

func test() {
	path := "database_file"
	err := Open(path)
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println(UserLogin("user1"))
	fmt.Println(UserLogin("user2"))
	fmt.Println(UserLogout("user2"))
	fmt.Println(SendMessage("user1", "hello?"))
	fmt.Println(SendMessage("user1", "hello!"))
	fmt.Println(QueryMessage("user1"))
	Close()
}

func MyServer(w http.ResponseWriter, r *http.Request) {
	var opt, username, message string
	for k, v := range r.URL.Query() {
		switch k {
		case "opt":
			opt = v[0]
		case "username":
			username = v[0]
		case "message":
			message = v[0]
		default:
		}
	}
	switch opt {
	case "login":
		fmt.Fprintln(w, UserLogin(username))
	case "logout":
		fmt.Fprintln(w, UserLogout(username))
	case "send":
		fmt.Fprintln(w, SendMessage(username, message))
	case "query":
		result := QueryMessage(username)
		for _, v := range result {
			fmt.Fprintln(w, v)
		}
	}
}

func main() {
	path := "database_file"
	err := Open(path)
	if err != nil {
		fmt.Println("error", err)
	}
	defer Close()

	http.HandleFunc("/", MyServer)

	if err := http.ListenAndServe(":3030", nil); err != nil {
		fmt.Printf("服务器连接出错！")
	}
}
