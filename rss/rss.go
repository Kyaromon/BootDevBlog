package rss


import (
    "context"
    "encoding/xml"
    "html"
    "io"
    "net/http"
)

type RSSFeed struct {
    XMLName xml.Name `xml:"rss"`
    Channel struct {
        Title       string    `xml:"title"`
        Description string    `xml:"description"`
        Items       []RSSItem `xml:"item"`
    } `xml:"channel"`
}

type RSSItem struct {
    Title       string `xml:"title"`
    Description string `xml:"description"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("User-Agent", "gator")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var feed RSSFeed
    err = xml.Unmarshal(body, &feed)
    if err != nil {
        return nil, err
    }

    feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
    feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
    for i := range feed.Channel.Items {
        feed.Channel.Items[i].Title = html.UnescapeString(feed.Channel.Items[i].Title)
        feed.Channel.Items[i].Description = html.UnescapeString(feed.Channel.Items[i].Description)
    }

    return &feed, nil
}
