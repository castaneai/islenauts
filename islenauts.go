package islenauts

import (
	"context"
	"net/http"
	"time"

	"fmt"

	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/castaneai/asaka"
	"google.golang.org/appengine/log"
)

const (
	baseURL = "https://www.lashinbang.com"
)

type Client struct {
	hc        *http.Client
	sessionID string
}

func NewClient(hc *http.Client) (*Client, error) {
	return &Client{
		hc: hc,
	}, nil
}

type Item struct {
	Title       string    `datastore:"title,noindex"`
	Description string    `datastore:"description,noindex"`
	ImageURL    string    `datastore:"imageUrl,noindex"`
	LinkURL     string    `datastore:"linkUrl,noindex"`
	Tags        []string  `datastore:"tags"`
	StoreNames  []string  `datastore:"storeNames,noindex"`
	CreatedAt   time.Time `datastore:"createdAt"`
}

func createCookie(name, value string) *http.Cookie {
	expires := time.Now().AddDate(1, 0, 0)
	return &http.Cookie{Name: name, Value: value, Expires: expires, HttpOnly: false}
}

func createSearchURL(search string) string {
	v := url.Values{}
	v.Set("search", search)
	return fmt.Sprintf("%s/kaitai/product/?%s", baseURL, v.Encode())
}

func (c *Client) GetItems(ctx context.Context, search string) ([]*Item, error) {
	opts := &asaka.ClientOption{
		Cookies: map[string]http.Cookie{"r18": *createCookie("r18", "is-r18")},
	}
	ac, err := asaka.NewClient(c.hc, opts)
	if err != nil {
		return nil, err
	}

	doc, err := ac.GetDoc(ctx, createSearchURL(search))
	if err != nil {
		return nil, err
	}
	var items []*Item
	doc.Find(".thumblist li").Each(func(i int, s *goquery.Selection) {
		item, err := parseItem(s)
		if err != nil {
			log.Infof(ctx, "[parse error] %+v", err)
			return
		}
		items = append(items, item)
	})
	return items, nil
}

func parseItem(s *goquery.Selection) (*Item, error) {
	createdAt, err := time.Parse("2006.01.02", s.Find(".fc_gray").Text())
	if err != nil {
		return nil, err
	}
	return &Item{
		Title:       s.Find("h3").Text(),
		Description: s.Find("p").Text(),
		ImageURL:    s.Find(".thumb img").AttrOr("src", ""),
		LinkURL:     baseURL + "/" + s.Find("h3 a").AttrOr("href", ""),
		Tags: s.Find(".label_tag span").Map(func(i int, ts *goquery.Selection) string {
			return ts.Text()
		}),
		StoreNames: s.Find(".label_store span").Map(func(i int, ts *goquery.Selection) string {
			return ts.Text()
		}),
		CreatedAt: createdAt,
	}, nil
}
