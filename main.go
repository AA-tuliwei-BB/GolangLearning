package main

import (
	"fmt"
	"net/http"
)

func test() {
	path := "database_file"
	err := Open(path)
	if err != nil {
		fmt.Println("error", err)
	}
	Close()
}

func MyServer(w http.ResponseWriter, r *http.Request) {
	var opt, username, message, token, passwd string
	passwd = ""
	token = ""
	for k, v := range r.URL.Query() {
		switch k {
		case "opt":
			opt = v[0]
		case "username":
			username = v[0]
		case "message":
			message = v[0]
		case "token":
			token = v[0]
		case "passwd":
			passwd = v[0]
		default:
		}
	}

	switch opt {
	case "signup":
		fmt.Fprintln(w, UserSignup(username, passwd))
	case "login":
		log, token := UserLogin(username, passwd)
		fmt.Fprintln(w, log, token)
	case "logout":
		fmt.Fprintln(w, UserLogout(username))
	case "send":
		fmt.Fprintln(w, SendMessage(username, token, message))
	case "query":
		result := QueryMessage(username, token)
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
