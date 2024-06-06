package app

import (
	"golang.org/x/net/html"
	"io"
	"strings"
)

func extractCSRFToken(html string) string {
	csrfTokenIndex := strings.Index(html, "csrf-token")
	if csrfTokenIndex == -1 {
		return ""
	}

	startIndex := strings.Index(html[csrfTokenIndex:], "content=") + 9
	token := html[startIndex+csrfTokenIndex:]
	endIndex := strings.Index(token, "\"")
	if endIndex == -1 {
		return ""
	}

	return token[:endIndex]
}

func extractRelevantCookie(cookieHeader string) string {
	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		if strings.Contains(cookie, "_yatri_session") {
			parts := strings.Split(cookie, "=")
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return ""
}

func getCookieBody(cookie string) string {
	return "_yatri_session=" + cookie
}

func getElementById(id string, n *html.Node) (element *html.Node, ok bool) {
	for _, a := range n.Attr {
		if a.Key == "name" && a.Val == id {
			return n, true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if element, ok = getElementById(id, c); ok {
			return
		}
	}
	return
}

func getAuthenticityToken(body io.Reader) string {
	root, _ := html.Parse(body)
	node, ok := getElementById("authenticity_token", root)
	if ok {
		for _, v := range node.Attr {
			if v.Key == "value" {
				return v.Val
			}
		}
	} else {
		logger.Error("authenticity_token not found")
	}
	return ""
}
