package tonic

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"github.com/G-Node/tonic/tonic/db"
)

const port = uint16(7357)
const cookieName = "tonic-test"

func newTestServiceWithRoutes(t *testing.T) *Tonic {
	srv, err := NewService(make([]Element, 1), noopAction, Config{Port: port})
	if err != nil {
		t.Fatalf("Failed to initialise test service: %s", err.Error())
	}

	if err := srv.setupWebRoutes(); err != nil {
		t.Fatalf("Failed to set up web routes: %s", err.Error())
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start test service: %s", err.Error())
	}
	return srv
}

// Create an http.Client that doesn't follow redirects so we can test if the
// server is responding with a redirect.
func httpClientNoRedirect() *http.Client {
	client := new(http.Client)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}

// Create an http.Client that doesn't follow redirects (for testing) and
// contains the specified cookie value.
func httpClientWithCookie(cookieValue string) *http.Client {
	client := httpClientNoRedirect()
	cookie := http.Cookie{
		Name:   cookieName,
		Value:  cookieValue,
		Domain: "localhost",
	}
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse(fmt.Sprintf("http://localhost:%d", port))
	jar.SetCookies(u, []*http.Cookie{&cookie})
	client.Jar = jar
	return client
}

func checkReqStatus(t *testing.T, client *http.Client, method string, path string, expStatus int) {
	address := fmt.Sprintf("http://localhost:%d%s", port, path)
	if resp, err := client.Get(address); err != nil {
		t.Fatalf("Get %q failed: %s", path, err.Error())
	} else {
		if resp.StatusCode != expStatus {
			t.Fatalf("Get %q status %d (expected %d)", path, resp.StatusCode, expStatus)
		}
	}
}

func TestReqLoginRedirects(t *testing.T) {
	srv := newTestServiceWithRoutes(t)
	defer srv.Stop()

	client := httpClientNoRedirect()

	// Routes that require login should redirect with 302
	checkReqStatus(t, client, http.MethodGet, "/", http.StatusFound)
	checkReqStatus(t, client, http.MethodPost, "/", http.StatusFound)
	checkReqStatus(t, client, http.MethodGet, "/log", http.StatusFound)
	checkReqStatus(t, client, http.MethodGet, "/log/42", http.StatusFound)

	client = httpClientWithCookie("bad cookie")
	checkReqStatus(t, client, http.MethodGet, "/", http.StatusFound)
	checkReqStatus(t, client, http.MethodPost, "/", http.StatusFound)
	checkReqStatus(t, client, http.MethodGet, "/log", http.StatusFound)
	checkReqStatus(t, client, http.MethodGet, "/log/42", http.StatusFound)
}

func TestAuthedRoutes(t *testing.T) {
	mockGIN()
	srv := newTestServiceWithRoutes(t)
	defer srv.Stop()

	sess := db.NewSession("fake token")
	job := new(db.Job)
	job.ValueMap = map[string]string{
		"key1":       "value1",
		"key2":       "value2",
		"anotherkey": "anothervalue",
		"onemore":    "lastvalue",
	}
	job.UserID = 42

	srv.db.InsertSession(sess)
	client := httpClientWithCookie(sess.ID)
	checkReqStatus(t, client, http.MethodGet, "/", http.StatusOK)
	checkReqStatus(t, client, http.MethodPost, "/", http.StatusOK)
	// checkReqStatus(t, client, http.MethodGet, "/log", http.Status)
	checkReqStatus(t, client, http.MethodGet, "/log/42", http.StatusNotFound)

}
