package main

import (
	"chat/server"
	"fmt"
	"log"
	"net/http"
)

func test() {
	path := "database_file"
	err := server.Open(path)
	if err != nil {
		fmt.Println("error", err)
	}
	server.Close()
}

func MyServer(w http.ResponseWriter, r *http.Request) {
	var opt, username, message, token, passwd, parameter, function string
	passwd = ""
	token = ""
	parameter = ""
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
		case "parameter":
			parameter = v[0]
		case "function":
			function = v[0]
		default:
		}
	}

	switch opt {
	case "signup":
		log.Printf("user: %s sign up with passwd: %s\n", username, passwd)
		fmt.Fprintln(w, server.UserSignup(username, passwd))
	case "login":
		log.Printf("user: %s login with passwd: %s;", username, passwd)
		res, token := server.UserLogin(username, passwd)
		log.Printf("get token: %s\n", token)
		fmt.Fprintln(w, res, token)
	case "logout":
		log.Printf("user: %s logout\n", username)
		fmt.Fprintln(w, server.UserLogout(username))

	case "send":
		if function == "cof" {
			log.Printf("user: %s, token: %s send: \"%s\"\n", username, token, message)
			fmt.Fprintln(w, server.SendMessage(username, token, message))
		} else if function == "chat" {
			log.Printf("user: %s, token: %s chat with %s: \"%s\"\n", username, token, parameter, message)
			fmt.Fprintln(w, server.SendChat(username, token, message, parameter))
		}

	case "query":
		if function == "cof" {
			log.Printf("user: %s, token: %s query cof\n", username, token)
			result := server.QueryMessage(username, token)
			for _, v := range result {
				fmt.Fprintln(w, v)
			}
		} else if function == "chat" {
			log.Printf("user: %s, token: %s query chat with %s\n", username, token, parameter)
			result := server.QueryChat(username, token, parameter)
			for _, v := range result {
				fmt.Fprintln(w, v)
			}
		}
	case "makefriend":
		log.Printf("%s make friend with %s", username, parameter)
		fmt.Fprintln(w, server.MakeFriend(username, parameter))
	}
}

func main() {
	path := "../database/database_file"
	err := server.Open(path)
	if err != nil {
		fmt.Println("error", err)
	}
	defer server.Close()

	http.HandleFunc("/", MyServer)

	if err := http.ListenAndServe(":3030", nil); err != nil {
		fmt.Printf("服务器连接出错！")
	}
}
