package main

import (
	"net/http"

	"context"
	"fmt"
	"os"

	"strings"

	"time"

	"encoding/base64"
	"io/ioutil"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/castaneai/islenauts"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	tweetPostDelay = 10 * time.Second
)

func main() {
	http.HandleFunc("/", handle)
	appengine.Main()
}

func handle(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	hc := urlfetch.Client(ctx)
	cli, err := islenauts.NewClient(hc)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		return
	}

	items, err := cli.GetItems(ctx, "tags:抱き枕カバー")
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		return
	}
	for _, item := range items {
		fmt.Fprintf(w, "<div><img src=\"%s\">%s</div>", item.ImageURL, item.Title)
	}

	last, err := getLastNotificationItem(ctx)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		return
	}
	newItems := filterNewNotifications(items, last)
	if len(newItems) > 0 {
		if err := saveNewNotifications(ctx, newItems); err != nil {
			log.Errorf(ctx, "%+v", err)
			return
		}
		if err := postNewItemsToTwitter(ctx, newItems); err != nil {
			log.Errorf(ctx, "%+v", err)
			return
		}
		fmt.Fprintf(w, "<h2>%d notifications saved.<h2>", len(newItems))
	}
}

func postNewItemsToTwitter(ctx context.Context, items []*islenauts.Item) error {
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	tw := anaconda.NewTwitterApiWithCredentials(accessToken, accessTokenSecret, consumerKey, consumerSecret)
	tw.HttpClient.Transport = &urlfetch.Transport{Context: ctx} // for GAE

	for _, item := range items {
		message := fmt.Sprintf("[islenauts-beta] %s\nTags: %s\n%s", item.Title, strings.Join(item.Tags, " "), item.LinkURL)
		if err := postTweet(ctx, tw, message, item.ImageURL); err != nil {
			return err
		}
		time.Sleep(tweetPostDelay)
	}
	return nil
}

func filterNewNotifications(items []*islenauts.Item, lastNotification *islenauts.Item) []*islenauts.Item {
	var fns []*islenauts.Item
	for _, n := range items {
		if isNewNotification(n, lastNotification) {
			fns = append(fns, n)
		}
	}
	return fns
}

func isNewNotification(n *islenauts.Item, last *islenauts.Item) bool {
	if last == nil {
		return true
	}
	// TODO: new notification check
	return !last.CreatedAt.Before(n.CreatedAt)
}

const (
	DatastoreKind = "IslenautsItems"
)

func getLastNotificationItem(ctx context.Context) (*islenauts.Item, error) {
	q := datastore.NewQuery(DatastoreKind).Order("-createdAt").Limit(1)

	var lasts []*islenauts.Item
	if _, err := q.GetAll(ctx, &lasts); err != nil {
		return nil, err
	}
	if len(lasts) < 1 {
		return nil, nil
	}
	return lasts[0], nil
}

func saveNewNotifications(ctx context.Context, items []*islenauts.Item) error {
	var keys []*datastore.Key
	for range items {
		keys = append(keys, datastore.NewIncompleteKey(ctx, DatastoreKind, nil))
	}
	return datastore.RunInTransaction(ctx, func(tc context.Context) error {
		if _, err := datastore.PutMulti(ctx, keys, items); err != nil {
			return err
		}
		return nil
	}, nil)
}

func postTweet(ctx context.Context, tw *anaconda.TwitterApi, message string, imageURL string) error {
	resp, err := tw.HttpClient.Get(imageURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	imageBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	b64Image := base64.StdEncoding.EncodeToString(imageBytes)
	media, err := tw.UploadMedia(b64Image)
	if err != nil {
		return err
	}
	v := url.Values{}
	v.Add("media_ids", media.MediaIDString)
	tres, err := tw.PostTweet(message, v)
	if err != nil {
		return err
	}
	log.Infof(ctx, "[tweet success] %s", tres.Text)
	return nil
}
