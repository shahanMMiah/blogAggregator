package rss

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)

	if err != nil {
		return RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return RSSFeed{}, err
	}

	feed := RSSFeed{}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return RSSFeed{}, err
	}

	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return RSSFeed{}, err
	}

	feed.Channel.Title = html.UnescapeString(
		feed.Channel.Title)

	feed.Channel.Description = html.UnescapeString(
		feed.Channel.Description)

	for num := range feed.Channel.Item {

		feed.Channel.Item[num].Title = html.UnescapeString(
			feed.Channel.Item[num].Title)

		feed.Channel.Item[num].Description = html.UnescapeString(
			feed.Channel.Item[num].Description)
	}

	return feed, nil

}
