package main

import (
    "crypto/tls"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "io"
    "strconv"
    "strings"
    "golang.org/x/net/html"
)


func main() {
    
    linkQueue := make(chan string)
    uniqeListQueue := make(chan string)

    go func() { linkQueue <- os.Args[1] }()
    go removeDuplicate(linkQueue, uniqeListQueue)

    finished := make(chan bool)

    go func() {
        for uri := range uniqeListQueue {
            queueLink(uri, linkQueue)
        }
        finished <- true
    }()
  <-finished
}

func formatLink(link string) string {
    if strings.Contains(link, "#") {
        var pointer int
        for n, str := range link {
            if strconv.QuoteRune(str) == "'#'" {
                pointer = n
                break
            }
        }
        return link[:pointer]
    }
    return link
}

func filteredDuplicate(links *[]string, uniqueLink []string) {
    for _, data := range uniqueLink {

       var duplicate bool
       for _, string := range *links {
          if string == data {
            duplicate = true
            break
           }
        }
        if duplicate == false {
            *links = append(*links, data)
        }
    }
}

func removeDuplicate(in chan string, out chan string) {
    var seen = make(map[string]bool)
    for val := range in {
        if !seen[val] {
            seen[val] = true
            out <- val
        }
    }
}

func queueLink(uri string, linkQueue chan string) {
    fmt.Println("hitting url", uri)
    transport := &http.Transport{ TLSClientConfig: &tls.Config{ InsecureSkipVerify: true, },}
    client := http.Client{Transport: transport}
    resp, err := client.Get(uri)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    links := GetChildLinks(resp.Body)

    for _, link := range links {
        formatedLink := fixUrl(link, uri)
        if uri != "" {
            go func() { linkQueue <- formatedLink }()
        }
    }
}

func fixUrl(href, parent string) string {
    uri, err := url.Parse(href)
    if err != nil {
        return ""
    }
    parentUrl, err := url.Parse(parent)
    if err != nil {
        return ""
    }
    uri = parentUrl.ResolveReference(uri)
    return uri.String()
}


func GetChildLinks(httpBody io.Reader) []string {
    links := []string{}
    uniqueLink := []string{}
    page := html.NewTokenizer(httpBody)
    for {
        tokenType := page.Next()
        if tokenType == html.ErrorToken {
            return links
        }
        token := page.Token()
        if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
            for _, attr := range token.Attr {
                if attr.Key == "href" {
                    link := formatLink(attr.Val)
                    uniqueLink = append(uniqueLink, link)
                    filteredDuplicate(&links, uniqueLink)
                }
            }
        }
    }
}
