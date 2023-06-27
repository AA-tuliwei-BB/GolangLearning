package main_test

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func HttpGet(par map[string]string) string {
	ApiUrl := "localhost:3030/"
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
		"opt":      "send",
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
	if res != "succeed" {
		return "signup error"
	}
	res = testLogin(username, passwd)
	split_res := strings.Split(res, " ")
	if split_res[0] != "Successfully" {
		return "login error"
	}
	token := split_res[len(split_res)-1]
	str1 := RandomString(15)
	str2 := RandomString(15)
	testSend(username, token, str1)
	testSend(username, token, str2)
	res = testQuery(username, token)
	split_res = strings.Split(res, "\n")
	if split_res[0] != str1 || split_res[1] != str2 {
		return "send or query error"
	}
	return "ok"
}

func TestRun(t *testing.T) {
	log.Println(testRun())
	log.Println(testRun())
	log.Println(testRun())
}
