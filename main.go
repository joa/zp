package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Command int

const (
	CommandUp   Command = 1
	CommandDown Command = 0
)

func (c Command) ToState() string { return strconv.Itoa(int(c)) }

const SessionCookieName = "XSSID"

var stringToCmd = map[string]Command{
	"up":   CommandUp,
	"down": CommandDown,
	"on":   CommandUp,
	"off":  CommandDown,
}

var (
	debug      = flag.Bool("debug", false, "whether or not to enable debug logging")
	protocol   = flag.String("protocol", "http", "http or https")
	addr       = flag.String("address", "", "address of the zyxel managed switch")
	username   = flag.String("username", "admin", "username for the web gui")
	password   = flag.String("password", "admin", "password for the web gui")
	port       = flag.Int("port", -1, "the port to power up or down")
	cmd        Command
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func main() {
	if !parseCmdline() {
		return
	}

	sess, err := login()

	if err != nil {
		fmt.Printf("couldn't request session: %s", err)
		os.Exit(1)
		return
	}

	if sess == "" {
		fmt.Println("invalid username/password")
		os.Exit(1)
		return
	}

	defer func() {
		if err := logout(sess); err != nil && *debug {
			fmt.Printf("logout failed: %s\n", err)
		}
	}()

	if err := updatePort(sess, *port, cmd); err != nil {
		fmt.Printf("update failed: %s\n", err)
		os.Exit(1)
	}
}

func parseCmdline() bool {
	var ok bool

	flag.Parse()

	cmd, ok = stringToCmd[strings.ToLower(flag.Arg(0))]

	if !ok {
		if arg := flag.Arg(0); arg == "" {
			fmt.Println("missing 'up' or 'down'")
		} else {
			fmt.Printf("'%s' should be 'up' or 'down'\n", arg)
		}

		os.Exit(2)

		return false
	}

	if *addr == "" {
		fmt.Println("missing -address")

		os.Exit(2)

		return false
	}

	return flag.Parsed()
}

func updatePort(token string, p int, cmd Command) (err error) {
	u := fmt.Sprintf("%s://%s/cgi-bin/dispatcher.cgi", *protocol, *addr)
	data := url.Values{}

	data.Add(SessionCookieName, token)
	data.Add("portlist", strconv.Itoa(p))
	data.Add("state", cmd.ToState())
	data.Add("portPriority", "3")
	data.Add("portPowerMode", "3")
	data.Add("portLimitMode", "0")
	data.Add("poeTimeRange", "20")
	data.Add("cmd", "775")
	data.Add("sysSubmit", "Apply")

	var req *http.Request
	var resp *http.Response

	if *debug {
		fmt.Println("data:", data.Encode())
	}

	req, err = http.NewRequest("POST", u, strings.NewReader(data.Encode()))

	if err != nil {
		return
	}

	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if *debug {
		fmt.Println("post", u)
	}

	resp, err = httpClient.Do(req)

	if err != nil && err.Error() != fmt.Sprintf("Post %s: net/http: HTTP/1.x transport connection broken: malformed MIME header line: <html>", u) {
		// server responds with broken header. go is not able to recover like other clients (for instance: curl)
		return
	} else {
		err = nil
	}

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	if resp != nil {
		if *debug {
			fmt.Println("status", resp.Status)
		}

		if status := resp.StatusCode; status < http.StatusOK || status >= http.StatusBadRequest {
			err = fmt.Errorf("unexpected status: %d", resp.StatusCode)
		}
	}

	return
}
