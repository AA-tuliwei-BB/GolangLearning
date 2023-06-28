package server

import (
	"chat/database"
	listdatabase "chat/listDatabase"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

type MessageInDb struct {
	Time    int64
	Message string
}

type Message struct {
	Time     int64
	Message  string
	Username string
}

var LoginDb, PasswdDb database.DataBase
var MessageDb, FriendDb listdatabase.ListDb
var ChatDb listdatabase.ListDb

func Open(path string) error {
	err := LoginDb.Open(path + "/Login.db")
	if err != nil {
		return err
	}
	err = PasswdDb.Open(path + "/Passwd.db")
	if err != nil {
		LoginDb.Close()
		return err
	}
	err = MessageDb.Open(path + "/MessageDb")
	if err != nil {
		LoginDb.Close()
		PasswdDb.Close()
		return err
	}
	err = FriendDb.Open(path + "/FriendDb")
	if err != nil {
		LoginDb.Close()
		PasswdDb.Close()
		MessageDb.Close()
		return err
	}
	err = ChatDb.Open(path + "/ChatDb")
	if err != nil {
		LoginDb.Close()
		PasswdDb.Close()
		MessageDb.Close()
		FriendDb.Close()
		return err
	}
	return nil
}

func Close() {
	LoginDb.Close()
	PasswdDb.Close()
	MessageDb.Close()
	FriendDb.Close()
	ChatDb.Close()
}

func GetToken(username string) string {
	secret := "secret"
	b := md5.Sum([]byte(username + secret))
	return hex.EncodeToString(b[:])
}

func CheckToken(username string, token string) bool {
	secret := "secret"
	b := md5.Sum([]byte(username + secret))
	return hex.EncodeToString(b[:]) == token
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

var make_friend_lock sync.Mutex
var signup_lock sync.RWMutex

func CheckExist(username string) bool {
	signup_lock.RLock()
	check, err := PasswdDb.Has([]byte(username))
	signup_lock.RUnlock()
	return check && err == nil
}

func IsFriend(username string, friend string) bool {
	friends := FriendDb.Query(username)
	for _, v := range friends {
		if v == friend {
			return true
		}
	}
	return false
}

func MakeFriend(username string, friend string) string {
	make_friend_lock.Lock()
	if IsFriend(username, friend) {
		return "They are already friends"
	}
	FriendDb.Insert(username, friend)
	if username != friend {
		FriendDb.Insert(friend, username)
	}
	make_friend_lock.Unlock()
	return "Succeed"
}

func UserSignup(username string, passwd string) string {
	signup_lock.Lock()
	check, err := PasswdDb.Has([]byte(username))
	if err != nil {
		log.Println("Error on check Signup", err)
		signup_lock.Unlock()
		return "error"
	}
	if check {
		signup_lock.Unlock()
		return "User exists"
	} else {
		PasswdDb.Write([]byte(username), []byte(passwd))
		MakeFriend(username, username)
		signup_lock.Unlock()
		return "Succeed"
	}
}

func UserCheckPasswd(username string, passwd string) bool {
	signup_lock.RLock()
	check, err := PasswdDb.Has([]byte(username))
	signup_lock.RUnlock()
	if err != nil {
		log.Println("Error on check passwd", err)
		return false
	}
	if check {
		signup_lock.RLock()
		result := (passwd == PasswdDb.Get([]byte(username)))
		signup_lock.RUnlock()
		return result
	} else {
		return false
	}
}

func UserLogin(username string, passwd string) (string, string) { // return (log, token)
	if !UserCheckPasswd(username, passwd) {
		return "Password incorrect", ""
	}
	check := CheckLogin(username)
	if check {
		return "User already logged in", ""
	} else {
		err := LoginDb.Modify([]byte(username), []byte("1"))
		if err != nil {
			return err.Error(), ""
		} else {
			return "Successfully log in", GetToken(username)
		}
	}
}

func UserLogout(username string) string {
	check := CheckLogin(username)
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

func checkQueryAndSend(username string, token string) string {
	check := CheckToken(username, token)
	if !check {
		return "invalid token"
	}
	check = CheckLogin(username)
	if !check {
		return "User not login"
	}
	return ""
}

func Query(db *listdatabase.ListDb, friends []string) []string {
	all_messages := []Message{}
	for _, v := range friends {
		messages := db.Query(v)
		if messages[0] == "No message" && len(messages) == 1 {
			continue
		}
		for _, message := range messages {
			tmp := MessageInDb{}
			json.Unmarshal([]byte(message), &tmp)
			tmp2 := Message{tmp.Time, tmp.Message, v}
			all_messages = append(all_messages, tmp2)
		}
	}
	sort.Slice(all_messages, func(i, j int) bool {
		return all_messages[i].Time > all_messages[j].Time
	})
	result := []string{}
	for _, v := range all_messages {
		buffer := fmt.Sprint(v.Time, " ", v.Username, " ", v.Message)
		result = append(result, buffer)
	}
	return result
}

func SendMessage(username string, token string, message string) string {
	check := checkQueryAndSend(username, token)
	if check != "" {
		return check
	}
	tmp := MessageInDb{time.Now().UnixMilli(), message}
	data, err := json.Marshal(tmp)
	err = MessageDb.Insert(username, string(data))
	if err != nil {
		return err.Error()
	}
	return "Successfully send message"
}

func QueryMessage(username string, token string) []string {
	check := checkQueryAndSend(username, token)
	if check != "" {
		return []string{check}
	}
	friends := FriendDb.Query(username)
	if len(friends) == 1 && friends[0] == "No message" {
		friends[0] = username
	}
	return Query(&MessageDb, friends)
}

func SendChat(username string, token string, message string, user2 string) string {
	check := checkQueryAndSend(username, token)
	if check != "" {
		return check
	}
	if !CheckExist(user2) {
		return user2 + " does not exist"
	}
	tmp := MessageInDb{time.Now().UnixMilli(), message}
	data, err := json.Marshal(tmp)
	err = ChatDb.Insert(username+"&"+user2, string(data))
	if err != nil {
		return err.Error()
	}
	return "Successfully send message"
}

func QueryChat(username string, token string, user2 string) []string {
	check := checkQueryAndSend(username, token)
	if check != "" {
		return []string{check}
	}
	if !CheckExist(user2) {
		return []string{user2 + " does not exist"}
	}
	friends := []string{username + "&" + user2, user2 + "&" + username}
	return Query(&ChatDb, friends)
}
