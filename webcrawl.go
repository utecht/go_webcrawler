package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "regexp"
    "os"
    "strconv"
)

type Fetcher interface {
    Fetch(url string) (body string, urls []string, err error)
}

type realFetcher string

func (f realFetcher) Fetch(url string) (body string, urls []string, err error){
    e := false
    e, body = GetURL(url)
    if e {
        urls = ParseURLs(url, body)
    } else {
        err = fmt.Errorf(body)
    }
    return
}

func GetURL(url string) (bool, string) {
    resp, err := http.Get(url)
    if err != nil {
       return false, err.Error()
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
       return false, err.Error()
    }
    return true, string(body)
}

func ParseURLs(url string, body string) []string {
    //re := regexp.MustCompile("<a.*href=\"(.*)\".*>")
    //re := regexp.MustCompile(`<a\s+href=("([^"]+)"|'([^']+)').*?>.*</a>`)
    re := regexp.MustCompile(`<a\s+href=("([^"]+)"|'([^']+)').*?>`)
    results := re.FindAllStringSubmatch(body, -1)
    ret := make([]string, len(results))
    for i, r := range results {
        //fmt.Println(r)
        if len(r[2]) > 4 && r[2][:4] == "http"{
            ret[i] = r[2]
        }else {
            ret[i] = url + r[2]
        }
    }
    return ret
}

func Crawl(url string, depth int, fetcher Fetcher, ret chan string, visited chan map[string]int){
    defer close(ret)

    v := <-visited
    _, ok := v[url]
    if ok {
       v[url]++
       visited <-v
       return
    } else {
        v[url] = 1
        visited <-v
    }

    if depth <= 0 {
        return
    }

    body, urls, err := fetcher.Fetch(url)
    if err != nil {
        ret <-err.Error()
        return
    }

    ret <-fmt.Sprintf("found: %s %d", url, len(body))

    result := make([]chan string, len(urls))
    for i, u := range urls {
        result[i] = make(chan string)
        go Crawl(u, depth-1, fetcher, result[i], visited)
    }

    for i := range result {
        for s := range result[i] {
            ret <-s
        }
    }
    return
}

func main(){
    if len(os.Args) < 3 {
        fmt.Println("Please pass in a starting website and depth")
        return
    }
    start := os.Args[1]
    x, err := strconv.ParseInt(os.Args[2], 10, 0)
    if err != nil {
        return
    }
    result := make(chan string)
    visited := make(chan map[string]int, 1)
    go Crawl(start, int(x), fetcher, result, visited)
    visited <- make(map[string]int)
    for s := range result {
        fmt.Println(s)
    }
    fmt.Printf("\n\nFound %d total links\n\n", len(<-visited))
}

var fetcher = realFetcher("test")
