package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

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
	base.Fragment = "" // strip fragments
	if crawler.Seen(base) {
		return nil
	}
	crawler.seen[base.String()] = struct{}{}

	var body io.ReadCloser
	if base.Scheme == "file" {
		var err error
		if body, err = os.Open(base.Path); err != nil {
			return fmt.Errorf("opening file:// url `%s`: %w", base, err)
		}
	} else if base.Scheme == "http" || base.Scheme == "https" {
		req, err := http.NewRequest("GET", base.String(), nil)
		if err != nil {
			return fmt.Errorf(
				"checking links for url `%s`: preparing HTTP request: %w",
				base,
				err,
			)
		}

		req.Header.Set("User-Agent", "linkcheck/1.0")

		rsp, err := crawler.Client.Do(req)
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
		body = rsp.Body
	} else {
		// ignore links that are not of scheme file, http, or https (e.g.,
		// ignore `mailto` links).
		return nil
	}

	// closes body (so we don't keep open file handles while crawling interior
	// links)
	doc, err := readDoc(body)
	if err != nil {
		return fmt.Errorf("checking links for url `%s`: %w", base, err)
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

func readDoc(body io.ReadCloser) (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf(
			"parsing HTML document: %w",
			errors.Join(err, body.Close()),
		)
	}
	if err := body.Close(); err != nil {
		return nil, fmt.Errorf("parsing HTML document: %w", err)
	}

	return doc, nil
}

type ErrNotOk int

func (e ErrNotOk) Error() string {
	return fmt.Sprintf("non-200 status code: %d", e)
}
