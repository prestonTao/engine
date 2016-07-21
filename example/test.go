package main

import (
	"fmt"
	// "math/rand"
	"io"
	"net/http"
	"time"
)

func example2() {
	mux := &MyMux{}
	http.ListenAndServe(":80", mux)
}

type MyMux struct{}

func (p *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("haha")
	if r.URL.Path == "/" {
		go nimei(w, r)
		return
	}
	http.NotFound(w, r)
	return
}

func main() {
	example2()
}

func example1() {
	http.HandleFunc("/nimei", nimei)

	http.HandleFunc("/hello", hello)

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println("ListenAndServer: ", err.Error())
	}
	fmt.Println("webServer startup...")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hello start")
	go test(w)
	fmt.Println("hello end")
}

func nimei(w http.ResponseWriter, r *http.Request) {
	fmt.Println("nimei start")
	io.WriteString(w, "nimei")
	time.Sleep(time.Second * 5)
	fmt.Println("nimei end")
}

func test(w http.ResponseWriter) {
	time.Sleep(time.Second * 10)
	io.WriteString(w, "hello")
	fmt.Println("test end")
}
