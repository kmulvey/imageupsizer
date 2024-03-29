package imageupsizer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var urlRegex = regexp.MustCompile(`(http|ftp|https):\/\/([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:\/~+#-]*[\w@?^=%&\/~+#-])`)
var OtherSizesNotAvailableError = errors.New("No other sizes of this image found.")
var NoMatchesError = errors.New("Looks like there aren’t any matches for your search")

func scrape(url string, linkFn findUrlFunc) (*url.URL, error) {
	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var html string
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.InnerHTML(`html`, &html),
	)
	if err != nil {
		return nil, err
	}

	return linkFn(html)
}

type findUrlFunc func(string) (*url.URL, error)

func findLargestImageLinkInHtml(html string) (*url.URL, error) {
	// cast a wide net around the link so we make sure we get it
	var firstDataIDRegex = regexp.MustCompile(`Image Results.*data-id="[a-zA-Z0-9_-]*"`)
	var firstDataID = firstDataIDRegex.FindString(html)
	var dataIDRegex = regexp.MustCompile(`data-id="[a-zA-Z0-9_-]*"`)
	var dataID = dataIDRegex.FindString(firstDataID)
	dataID = strings.ReplaceAll(dataID, "data-id=", "")
	dataID = strings.ReplaceAll(dataID, `"`, "")
	var jsBlockRegex = regexp.MustCompile(dataID + ".*" + dataID)
	var jsBlock = jsBlockRegex.FindAllString(html, 2)
	if len(jsBlock) < 2 {
		return nil, errors.New("did not find enough urls in the js block")
	}

	var urls = urlRegex.FindAllString(jsBlock[1], 2)
	if len(urls) < 2 {
		if regexp.MustCompile("Looks like there aren’t any matches for your search").MatchString(html) {
			return nil, NoMatchesError
		}

		var err = os.WriteFile("largest_image.html", []byte(html), os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("could not dump to largest_image.html, err: %w", err)
		}

		return nil, errors.New("did not find enough urls")
	}

	return url.Parse(urls[1])
}
func findAllSizesLinkInHtml(html string) (*url.URL, error) {
	// cast a wide net around the link so we make sure we get it
	var wideLinkRegex = regexp.MustCompile(`\/search\?tbs=simg:.*>All sizes`)
	var wideLink = wideLinkRegex.FindString(html)
	var index = strings.Index(wideLink, `"`)
	if index == -1 {
		if regexp.MustCompile(`No other sizes of this image found.`).MatchString(html) {
			return nil, OtherSizesNotAvailableError
		}

		var err = os.WriteFile("all_sizes.html", []byte(html), os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("could not dump to all_sizes.html, err: %w", err)
		}

		return nil, errors.New("wide link not found in html, html dumped to all_sizes.html")
	}
	var link = wideLink[:index]
	link = strings.ReplaceAll(link, "&amp;", "&")

	return url.Parse("https://www.google.com" + link)
}

func findImageSourceLinkInHtml(html string) (*url.URL, error) {
	var linkRegex = regexp.MustCompile(`https:\/\/www\.google\.com\/search\?tbs=sbi:[a-zA-Z0-9_-]*`)
	var link = linkRegex.FindString(html)

	return url.Parse(link)
}

func findImageInFacebookHtml(html string) (*url.URL, error) {
	html = strings.TrimSpace(html)
	var wideURLRegex = regexp.MustCompile(`additional_profile_has_taggable_products.*accessibility_caption`)
	var js = wideURLRegex.FindString(html)

	var begin = strings.Index(js, `{"uri":"`) + 8
	var end = strings.Index(js, `","width"`)

	return url.Parse(strings.ReplaceAll(js[begin:end], "\\/", "/"))
}
