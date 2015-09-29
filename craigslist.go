package craigslist

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Client struct {
	HttpClient      *http.Client
	UseStoredOffset bool
}

type SearchList struct {
	Searches []Search
}

type Search struct {
	Id         int64
	Title      string
	Location   string
	Url        string
	PostDate   string
	HasPicture bool
}

func (this *Client) RequestPage(url string) (document *goquery.Document, err error) {

	if this.HttpClient == nil { //If client is not set, create an empty one.
		this.HttpClient = &http.Client{}
	}

	//Create a request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	//Do request, get response
	resp, err := this.HttpClient.Do(req)
	if err != nil {
		return
	}

	//Turn to document
	document, err = goquery.NewDocumentFromResponse(resp)

	if err != nil {
		return
	}
	return
}

func (this *Client) GetSearchList(location string, category string) (searchList SearchList, err error) {

	//Fetch Search List
	doc, err := this.RequestPage("http://" + location + ".craigslist.org/search/" + category)
	if err != nil {
		return
	}

	offset, err := this.readSearchListFile(location, category)
	if err != nil {
		return
	}

	//Iterate result rows
	newOffset := offset
	doc.Find(".content .row").Each(func(i int, row *goquery.Selection) {
		search := Search{}
		id, _ := row.Attr("data-pid")
		search.Id, err = strconv.ParseInt(id, 10, 32)

		url, _ := row.Find(".hdrlnk").First().Attr("href")
		search.Url = "http://" + location + ".craigslist.com" + url
		search.Title = row.Find(".hdrlnk").First().Text()

		search.Location = strings.TrimSpace(row.Find(".l2 .pnr small").First().Text())
		search.PostDate, _ = row.Find(".pl time").First().Attr("datetime")
		hasPic := row.Find(".l2 .pnr .px .p").First().Text()
		if hasPic == " pic" {
			search.HasPicture = true
		}

		if search.Id > offset {
			if newOffset < search.Id {
				newOffset = search.Id
			}
			searchList.Searches = append(searchList.Searches, search)
		}
	})

	err = this.writeSearchListFile(location, category, newOffset)
	if err != nil {
		return
	}
	return
}

func (this *Client) readSearchListFile(location string, category string) (offset int64, err error) {
	if !this.UseStoredOffset {
		return
	}
	var data []byte

	if _, err = os.Stat(location + "_" + category + ".dat"); os.IsNotExist(err) {
		//File does not exist, and that's fine, just return.
		err = nil
		return
	}
	data, err = ioutil.ReadFile(location + "_" + category + ".dat")
	offset, err = strconv.ParseInt(string(data), 10, 32)
	return
}

func (this *Client) writeSearchListFile(location string, category string, offset int64) (err error) {
	if !this.UseStoredOffset {
		return
	}

	//Write results to file
	err = ioutil.WriteFile(location+"_"+category+".dat", []byte(fmt.Sprintf("%d", offset)), 0644)
	if err != nil {
		return
	}

	return
}

func (this *Client) SearchPageForKeywords(url string, keywords []string) (keywordsFound []string, err error) {
	//Fetch Search List
	doc, err := this.RequestPage(url)
	content := doc.Find("#postingbody").First().Text()
	content = strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(content, strings.TrimSpace(strings.ToLower(keyword))) {
			keywordsFound = append(keywordsFound, keyword)
		}
	}
	return
}
