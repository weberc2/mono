package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type ResultVisitor interface {
	Visit(*Result)
}

type Crawler struct {
	Client   *http.Client
	Host     string
	Callback func(*Result)
	seen     map[string]struct{}
}

func NewCrawler(host string) *Crawler {
	return &Crawler{
		Client:   http.DefaultClient,
		Host:     host,
		Callback: func(*Result) {},
		seen:     map[string]struct{}{},
	}
}

func (crawler *Crawler) SetVisitor(
	printer ResultVisitor,
) *Crawler {
	crawler.Callback = printer.Visit
	return crawler
}

func (crawler *Crawler) Reset() *Crawler {
	crawler.seen = map[string]struct{}{}
	return crawler
}

func (crawler *Crawler) SetHost(host string) *Crawler {
	crawler.Host = host
	return crawler
}

func (crawler *Crawler) SetClient(client *http.Client) *Crawler {
	crawler.Client = client
	return crawler
}

func (crawler *Crawler) SetCallback(callback func(*Result)) *Crawler {
	crawler.Callback = callback
	return crawler
}

func (crawler *Crawler) Seen(url *url.URL) bool {
	_, seen := crawler.seen[url.String()]
	return seen
}

func (crawler *Crawler) Crawl(base *url.URL) error {
	if base.Scheme != "http" && base.Scheme != "https" {
		// ignore non-http/https links (e.g., mailto)
		return nil
	}

	base.Fragment = "" // strip fragments
	if crawler.Seen(base) {
		return nil
	}
	crawler.seen[base.String()] = struct{}{}

	rsp, err := crawler.Client.Get(base.String())
	if err != nil {
		return fmt.Errorf("checking links for url `%s`: %w", base, err)
	}

	if rsp.StatusCode != http.StatusOK {
		return ErrNotOk(rsp.StatusCode)
	}

	// If the URL's host is not part of the site we're checking, then return
	// early
	if base.Host != crawler.Host {
		return nil
	}

	// Otherwise recurse into the links in the response body
	defer func() {
		if err := rsp.Body.Close(); err != nil {
			log.Printf(
				"ERROR closing response body for url `%s`: %v",
				base,
				err,
			)
		}
	}()

	doc, err := goquery.NewDocumentFromReader(rsp.Body)
	if err != nil {
		return fmt.Errorf(
			"checking links for url `%s`: parsing HTML document: %w",
			base,
			err,
		)
	}
	doc.Find("a[href]").Each(func(i int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists {
			html, err := a.Html()
			if err != nil {
				log.Printf("error rendering `%v` as HTML: %v", a, err)
			}
			log.Fatalf(
				"program error: goquery selector should query only <a> tags "+
					"with `href` attributes, but this tag is missing an "+
					"href attribute: %s",
				html,
			)
		}

		// correctly parses hrefs relative to `base`:
		// * absolute urls: https://foo.com
		// * relative urls: /baz ./bar ../../qux
		target, err := base.Parse(href)
		if err != nil {
			crawler.Callback(&Result{
				BaseURL:       base.String(),
				TargetText:    a.Text(),
				TargetURL:     href,
				URLParseError: err,
			})
			return
		}

		if err := crawler.Crawl(target); err != nil {
			if statusCode, ok := err.(ErrNotOk); ok {
				crawler.Callback(&Result{
					BaseURL:    base.String(),
					TargetText: a.Text(),
					TargetURL:  href,
					StatusCode: int(statusCode),
				})
				return
			}
			crawler.Callback(&Result{
				BaseURL:      base.String(),
				TargetText:   a.Text(),
				TargetURL:    href,
				NetworkError: err,
			})
			return
		}

		crawler.Callback(&Result{
			BaseURL:    base.String(),
			TargetText: a.Text(),
			TargetURL:  href,
			StatusCode: 200,
		})
	})

	return nil
}

type ErrNotOk int

func (e ErrNotOk) Error() string {
	return fmt.Sprintf("non-200 status code: %d", e)
}
