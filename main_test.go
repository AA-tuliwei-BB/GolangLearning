package main_test

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func HttpGet(par map[string]string) string {
	ApiUrl := "http://127.0.0.1:3030/"
	data := url.Values{}
	for k, v := range par {
		data.Set(k, v)
	}
	u, err := url.ParseRequestURI(ApiUrl)
	if err != nil {
		log.Println("testLogin, url:", err)
	}
	u.RawQuery = data.Encode()
	resp, err := http.Get(u.String())
	if err != nil {
		log.Println("testLogin, Get:", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	return string(b)
}

func testLogin(username string, passwd string) string {
	par := map[string]string{
		"opt":      "login",
		"username": username,
		"passwd":   passwd,
	}
	return HttpGet(par)
}

func testLogout(username string) string {
	par := map[string]string{
		"opt":      "logout",
		"username": username,
	}
	return HttpGet(par)
}

func testSignup(username string, passwd string) string {
	par := map[string]string{
		"opt":      "signup",
		"username": username,
		"passwd":   passwd,
	}
	return HttpGet(par)
}

func testSend(username string, token string, message string) string {
	par := map[string]string{
		"opt":      "send",
		"username": username,
		"token":    token,
		"message":  message,
	}
	return HttpGet(par)
}

func testQuery(username string, token string) string {
	par := map[string]string{
		"opt":      "query",
		"username": username,
		"token":    token,
	}
	return HttpGet(par)
}

func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func testRun() string {
	username := RandomString(10)
	passwd := RandomString(10)
	res := testSignup(username, passwd)
	if res != "Succeed\n" {
		return "signup error"
	}
	res = testLogin(username, passwd)
	split_res := strings.Split(res, " ")
	if split_res[0] != "Successfully" {
		return "login error"
	}
	token := split_res[len(split_res)-1]
	token = strings.TrimRight(token, "\n")
	str1 := RandomString(15)
	str2 := RandomString(15)
	res = testSend(username, token, str1)
	if strings.Split(res, " ")[0] != "Successfully" {
		log.Println(username, ": send error")
		return "send error!!"
	}
	res = testSend(username, token, str2)
	if strings.Split(res, " ")[0] != "Successfully" {
		log.Println(username, ": send error")
		return "send error"
	}
	res = testQuery(username, token)
	split_res = strings.Split(res, "\n")
	if split_res[1] != str1 || split_res[0] != str2 {
		log.Println(username, token, str1, str2, split_res)
		return "query error"
	}
	return "ok7"
}

func TestRun(t *testing.T) {
	log.Println(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())
	//	for i := 1; i <= 4; i = i + 1 {
	//	go testRun()
	//}
	log.Println(testRun())
}
