package main

import (
	"fmt"
	"net/http"

	badger "github.com/dgraph-io/badger/v3"
)

// badger
// postman
//

var count int

func MyServer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "test %d!!!\n", count)
	count = count + 1

	for k, v := range r.URL.Query() {
		if k == "content" {
			Buffer := []rune(v[0])
			var BufferLen = len(Buffer)
			if Buffer[len(Buffer)-1] == []rune("?")[0] || Buffer[len(Buffer)-1] == []rune("？")[0] {
				if Buffer[len(Buffer)-2] == []rune("吗")[0] {
					BufferLen = len(Buffer) - 2
				} else {
					BufferLen = len(Buffer) - 1
				}
			}
			if Buffer[0] == []rune("你")[0] {
				Buffer[0] = []rune("我")[0]
			}
			fmt.Fprintln(w, string(Buffer[:BufferLen]), "！")
		}
	}
}

func main() {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	http.HandleFunc("/", MyServer)

	if err := http.ListenAndServe(":3030", nil); err != nil {
		fmt.Printf("服务器连接出错！")
	}
}
