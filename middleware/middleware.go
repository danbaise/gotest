package main

import (
	"net/http"
	"io"
	"log"
	"context"

)

type MiddleWare interface{
	Chain(h http.Handler) http.Handler
}

type Chain struct{
	final http.Handler
	middlewares []MiddleWare
}

func (c *Chain)New(raw http.Handler,middlewares ...MiddleWare) *Chain{
	return &Chain{
		final:raw,
		middlewares:middlewares,
	}
}

func (c *Chain)Then(middleware MiddleWare) *Chain{
	c.middlewares=append(c.middlewares,middleware)
	return c
}

func (c *Chain)ServeHTTP(w http.ResponseWriter,r *http.Request) {
	final:=c.final
	for _,v := range c.middlewares{
		final=v.Chain(final)
	}
	final.ServeHTTP(w,r)
}

type test3 struct {
	name string
}
func (t *test3) Chain(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w,"test3\r\n")

		if v := r.Context().Value("username"); v != nil {
			io.WriteString(w, v.(string)+"\r\n")
		}
		if t.name=="test3"{
			io.WriteString(w, "struct\r\n")
		}
		base.ServeHTTP(w, r)
	})
}

type test4 struct {}

func (t *test4) Chain(base http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w,"test4\r\n")
		ctx := context.WithValue(r.Context(), "username", "zj")
		base.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main(){
	http.Handle("/chain", new(Chain).New(http.HandlerFunc(Hello)).Then(&test3{name:"test3"}).Then(new(test4)))
	http.Handle("/", http.HandlerFunc(Hello))
	http.Handle("/test1", test1(http.HandlerFunc(Hello)))
	log.Fatal(http.ListenAndServe(":1234",nil))
}

func Hello(w http.ResponseWriter, r *http.Request){
	if v := r.Context().Value("username"); v != nil {
		io.WriteString(w, "456\r\n")
	}
	io.WriteString(w,"hello world")
}

func test1(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		io.WriteString(w,"simple middleware\r\n")
		h.ServeHTTP(w,r)
	})
}
