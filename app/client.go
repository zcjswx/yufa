package app

import (
	"net/http"
)

/*
	One user uses one MyClient to run all requests.
*/

var Client *MyClient

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MyClient struct {
	*http.Client
	*http.Header
}

func GetClient() *MyClient {
	if Client != nil {
		return Client
	}
	return NewClient()
}

func NewClient() *MyClient {
	c := &MyClient{
		&http.Client{},
		NewHeader(),
	}
	Client = c
	return Client
}

func NewHeader() *http.Header {
	header := &http.Header{}
	header.Set("User-Agent", userAgent)
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Connection", "keep-alive")
	header.Set("Cache-Control", "no-cache")
	header.Set("Referer", baseURI)
	header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	header.Set("Accept", "*/*")
	return header
}

func (c *MyClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) > 0 {
		logger.Debug("Set-Cookie Header Found:", cookies)
		c.Header.Set("Cookie", getCookieBody(extractRelevantCookie(resp.Header.Get("Set-Cookie"))))
	}

	return resp, err
}
