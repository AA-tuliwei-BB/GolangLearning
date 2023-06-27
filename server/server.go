package server

// login request: localhost:port?opt=login&username=$username
// return: "success!" or "failed!"

// logout request: localhost:port?opt=logout&username=$username
// return: "success!" or "failed!"

// send message request: localhost:port?opt=send&username=$username&message=$message
// return: "success!" or "failed!"

// query request: localhost:port?opt=query
// return "Error!" or $message

import (
	"chat/database"
	listdatabase "chat/listDatabase"
	"crypto/md5"
	"encoding/hex"
	"log"
	"sync"
)

var LoginDb, PasswdDb database.DataBase
var MessageDb listdatabase.ListDb

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
	return nil
}

func Close() {
	LoginDb.Close()
	PasswdDb.Close()
	MessageDb.Close()
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

func UserSignup(username string, passwd string) string {
	check, err := PasswdDb.Has([]byte(username))
	if err != nil {
		log.Println("Error on check Signup", err)
		return "error"
	}
	if check {
		return "User exists"
	} else {
		PasswdDb.Write([]byte(username), []byte(passwd))
		return "Succeed"
	}
}

func UserCheckPasswd(username string, passwd string) bool {
	check, err := PasswdDb.Has([]byte(username))
	if err != nil {
		log.Println("Error on check passwd", err)
		return false
	}
	if check {
		return passwd == PasswdDb.Get([]byte(username))
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

func QueryMessage(username string, token string) []string {
	var result = []string{}
	check := CheckToken(username, token)
	if !check {
		result = append(result, "invalid token")
		return result
	}
	check = CheckLogin(username)
	if !check {
		result = append(result, "User not login")
		return result
	}
	result = MessageDb.Query(username)
	return result
}

var message_lock sync.Mutex

func SendMessage(username string, token string, message string) string {
	check := CheckToken(username, token)
	if !check {
		return "invalid token"
	}
	check = CheckLogin(username)
	if !check {
		return "User not login"
	}
	err := MessageDb.Insert(username, message)
	if err != nil {
		return err.Error()
	}
	return "Successfully send message"
}
