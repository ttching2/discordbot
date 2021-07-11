package strawpoll_test

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func thingy(v string) {
	fmt.Println(v)
}

func TestThing(t *testing.T) {
	req, err := http.NewRequest("GET", "https://manganato.com/manga-hd984612", nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	node, _ := html.Parse(r.Body)
	body := node.FirstChild.FirstChild.NextSibling
	bodySite := body.LastChild.PrevSibling.PrevSibling.PrevSibling.PrevSibling.PrevSibling.PrevSibling.PrevSibling
	containerMain := bodySite.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling
	containerMainLeft := containerMain.FirstChild.NextSibling
	chapterList := containerMainLeft.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling
	chapterRows := chapterList.FirstChild.NextSibling.NextSibling.NextSibling
	firstRow := chapterRows.FirstChild.NextSibling
	time := firstRow.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling
	fmt.Println(findTime(time))
	thing := thing{body, 0, false}
	thing.traverse(body)
	fmt.Println(thing.newChapter)
}

type thing struct {
	n          *html.Node
	row        int
	newChapter bool
}

func (t *thing) traverse(n *html.Node) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == "div" {
			for _, attr := range child.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "chapter-row") {
					if t.row == 2 {
						t.row++
						t.newChapter = findTime(child)
						return
					}
					t.row++
				}
			}
		}
		t.traverse(child)
	}
}

func findTime(n *html.Node) bool {
		for _, attr := range n.Attr {
			if attr.Key == "title" {
				releaseTime, err := time.Parse("Jan 2,2006 15:04", attr.Val)
				if err != nil {
					fmt.Print(err)
				} else {	
					now := time.Now().Add(time.Hour * 3)
					return now.Sub(releaseTime).Hours() <= 1
				}
			}
		}
	
	return false
}
