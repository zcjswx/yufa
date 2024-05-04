package app

import "strings"

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
