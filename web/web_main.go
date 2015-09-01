package main

import (
    "net/http"
    "log"
    "html/template"
    "strings"
    "fmt"
)

var WEB_HOME = "./nse_web"

func homePage(w http.ResponseWriter, r *http.Request){
    home := strings.Join([]string{WEB_HOME,"home.html"}, "/")
    t,err:= template.ParseFiles(home)
    var data interface {}
    if err != nil{
        log.Fatal(err)
        panic(err)
    }
    t.Execute(w,data)
}
func serveStaticFiles(w http.ResponseWriter, r *http.Request){
    fname := strings.Join([]string{WEB_HOME, r.URL.Path[1:]}, "/")
    fmt.Println("url: ", r.URL, fname)
    
    http.ServeFile(w,r, fname)
}

func main(){

    http.HandleFunc("/",homePage)
    http.HandleFunc("/static/", serveStaticFiles)
    err := http.ListenAndServe(":9090", nil)
    if err != nil{
        log.Fatal("ListenAndServe:", err)
        panic(err)
    }
}