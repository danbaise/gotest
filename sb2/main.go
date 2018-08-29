package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

func text(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `<doctype html>
        <html>
        <head>
          <title>Hello World</title>
        </head>
        <body>
        <p>
          <a href="/welcome">Welcome</a> |  <a href="/message">Message</a>
        </p>
        </body>
</html>`
	fmt.Fprintln(w, html)
}

func tpl(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles(filepath.Dir(os.Args[0]) + "/tpl/index.html"))
	c := &Data{"这是一个测试模版"}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	t.Execute(w, c)
}

type Data struct {
	Title string
}

func main() {

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(index))
	mux.Handle("/tpl", http.HandlerFunc(tpl))
	mux.HandleFunc("/text", text)
	mux.Handle("/sbadmin2/", http.StripPrefix("/sbadmin2/", http.FileServer(http.Dir(filepath.Dir(os.Args[0])+"/sbadmin2"))))

	http.ListenAndServe(":80", mux)
}
