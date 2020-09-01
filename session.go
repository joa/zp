package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// login to the switch.
//
// There are two http calls here. The first one performs the actual login and establishes a server-side session.
// The second call requests the frameset which contains JavaScript that'll set the session cookie.
func login() (token string, err error) {
	// (1) perform login
	loginURL := fmt.Sprintf(
		"%s://%s/cgi-bin/dispatcher.cgi?login=1&username=%s&password=%s&dummy=%d",
		*protocol,
		*addr,
		url.QueryEscape(*username),
		url.QueryEscape(encode(*password)),
		time.Now().Unix()*1000)

	if *debug {
		fmt.Println("get", loginURL)
	}

	resp, err := httpGet(loginURL, "")

	if err != nil {
		return
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	if *debug {
		fmt.Println("status", resp.Status)
	}

	if status := resp.StatusCode; status < http.StatusOK || status >= http.StatusBadRequest {
		err = fmt.Errorf("unexpected status: %d", resp.StatusCode)
		return
	}

	// (2) get session
	sessionURL := fmt.Sprintf("%s://%s/cgi-bin/dispatcher.cgi?cmd=1", *protocol, *addr)

	if *debug {
		fmt.Println("get", sessionURL)
	}

	resp, err = httpGet(sessionURL, "")

	if err != nil {
		return
	}

	if *debug {
		fmt.Println("status", resp.Status)
	}

	if status := resp.StatusCode; status < http.StatusOK || status >= http.StatusBadRequest {
		err = fmt.Errorf("unexpected status: %d", resp.StatusCode)
		return
	}

	var body []byte

	body, err = ioutil.ReadAll(resp.Body)

	_ = resp.Body.Close()

	if err != nil {
		return
	}

	if *debug {
		fmt.Println("received html:")
		fmt.Println(string(body))
	}

	token = parseSession(body)

	if *debug {
		fmt.Println("session:", token)
	}

	return
}

// logout from the switch
func logout(token string) (err error) {
	var resp *http.Response

	u := fmt.Sprintf("%s://%s/cgi-bin/dispatcher.cgi?cmd=4", *protocol, *addr)

	if *debug {
		fmt.Println("get", u)
	}

	resp, err = httpGet(u, token)

	if err != nil {
		return
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	if *debug {
		fmt.Println("status", resp.Status)
	}

	if status := resp.StatusCode; status < http.StatusOK || status >= http.StatusBadRequest {
		err = fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return
}

// parseSession inside html data.
func parseSession(html []byte) string {
	res := regexp.MustCompile(fmt.Sprintf(`(?s)setCookie\("%s", "(.+?)"\);`, SessionCookieName)).FindSubmatch(html)

	if len(res) < 2 {
		return ""
	}

	return string(res[1])
}

// httpGet the URL with a given session.
func httpGet(u string, sess string) (resp *http.Response, err error) {
	var req *http.Request

	req, err = http.NewRequest("GET", u, nil)

	if err != nil {
		return
	}

	if sess != "" {
		req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sess})
	}

	resp, err = httpClient.Do(req)

	return
}
