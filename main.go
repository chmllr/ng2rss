package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
)

const url = "https://www.nationalgeographic.com/bin/services/core/public/query/content.json?contentTypes=adventure/components/pagetypes/story/article,adventure/components/pagetypes/story/gallery,adventure/components/pagetypes/story/interactive,adventure/components/pagetypes/story/multipage,animals/components/pagetypes/story/article,animals/components/pagetypes/story/gallery,animals/components/pagetypes/story/interactive,animals/components/pagetypes/story/multipage,archaeologyandhistory/components/pagetypes/story/article,archaeologyandhistory/components/pagetypes/story/gallery,archaeologyandhistory/components/pagetypes/story/interactive,archaeologyandhistory/components/pagetypes/story/multipage,environment/components/pagetypes/story/article,environment/components/pagetypes/story/gallery,environment/components/pagetypes/story/interactive,environment/components/pagetypes/story/multipage,magazine/components/pagetypes/story/article,magazine/components/pagetypes/story/gallery,magazine/components/pagetypes/story/interactive,magazine/components/pagetypes/story/multipage,news/components/pagetypes/story/article,news/components/pagetypes/story/gallery,news/components/pagetypes/story/interactive,news/components/pagetypes/story/multipage,peopleandculture/components/pagetypes/story/article,peopleandculture/components/pagetypes/story/gallery,peopleandculture/components/pagetypes/story/interactive,peopleandculture/components/pagetypes/story/multipage,photography/components/pagetypes/story/article,photography/components/pagetypes/story/gallery,photography/components/pagetypes/story/interactive,photography/components/pagetypes/story/multipage,science/components/pagetypes/story/article,science/components/pagetypes/story/gallery,science/components/pagetypes/story/interactive,science/components/pagetypes/story/multipage,travel/components/pagetypes/story/article,travel/components/pagetypes/story/gallery,travel/components/pagetypes/story/interactive,travel/components/pagetypes/story/multipage&sort=newest&operator=or&includedTags=&excludedTags=ngs_genres:reference,ngs_series:expedition_antarctica,ngs_visibility:omit_from_hp&excludedGuids=beda7baa-e63b-4276-8122-34e47a4e653e&pageSize=12&page=0&offset=0"

func handler(w http.ResponseWriter, r *http.Request) {
	rss, err := feed()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", rss)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":9090", nil)
}

type story struct {
	Url         string `json:"url"`
	Title       string `json:"title"`
	Abstract    string `json:"abstract"`
	PublishDate string `json:"publishDate"`
}

type page struct {
	Story story `json:"page"`
}

func stories2RSS(stories []story) (string, error) {
	feed := &feeds.Feed{
		Title:       "National Geographic",
		Link:        &feeds.Link{Href: "https://github.com/chmllr/ng2rss"},
		Description: "National Geographic",
		Author:      &feeds.Author{Name: "Christian Müller", Email: "@drmllr"},
		Created:     time.Now(),
	}

	feed.Items = make([]*feeds.Item, len(stories))
	for i, story := range stories {
		date, err := time.Parse(time.UnixDate, story.PublishDate)
		if err != nil {
			log.Println("couldn't parse story publishing date", story.PublishDate)
			date = time.Date(2017, time.December, 21, 21, 10, 0, 0, time.UTC)
		}
		feed.Items[i] = &feeds.Item{
			Title:       story.Title,
			Link:        &feeds.Link{Href: story.Url},
			Description: story.Abstract,
			Author:      &feeds.Author{Name: "National Geographic"},
			Created:     date,
		}
	}

	return feed.ToRss()
}

func feed() (string, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("couldn't fetch page: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't read response: %v", err)
	}
	pages := []page{}
	if err := json.Unmarshal(body, &pages); err != nil {
		return "", fmt.Errorf("couldn't unmarshal response: %v", err)
	}
	res := []story{}
	for _, page := range pages {
		res = append(res, page.Story)
	}
	defer log.Println("story request and feed assembling took", time.Since(start))
	return stories2RSS(res)
}
