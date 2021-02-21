package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

var wg sync.WaitGroup

//SitemapIndex Indexes the sitemap
type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

//News is an object to show the news or something
type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keyword  string
	Location string
}

type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

func newsRoutine(c chan News, Location string) {
	defer wg.Done()
	var n News

	Location = strings.TrimSpace(Location)
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()

	c <- n
}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s SitemapIndex

	newsMap := make(map[string]NewsMap)

	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemaps/index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	resp.Body.Close()

	queue := make(chan News, 30)

	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}

	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx, _ := range elem.Keywords {
			newsMap[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "A TITLE", News: newsMap}

	t, _ := template.ParseFiles("basictemplating.gohtml")

	t.Execute(w, p)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Woah GO</h1>")
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/agg", newsAggHandler)
	http.ListenAndServe(":8000", nil)
}

// package main

// import (
// 	"encoding/xml"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// )

// func main() {
//
// 	for idx, data := range newsMap {
// 		fmt.Println("/n/n/n", idx)
// 		fmt.Println("/n", data.Keyword)
// 		fmt.Println("/n", data.Location)
// 	}
// 	resp.Body.Close()
// }
