package strictrequest_test

import (
	"fmt"
	"github.com/kozmos/strict-request"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func CreateTestServer(handlerFunc func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handlerFunc))
}

func TestSimpleRequest(t *testing.T) {
	wait := make(chan bool, 1)

	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")

		assert.Equal(t, "GET", r.Method)

		wait <- false
	})

	defer testServer.Close()

	resp, err := strictrequest.StrictRequest("GET", testServer.URL, strictrequest.Options{})
	assert.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, string(body), "ok\n")

	<-wait
}

func TestPostingBody(t *testing.T) {
	wait := make(chan bool, 1)

	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")

		assert.Equal(t, "POST", r.Method)

		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)

		assert.Equal(t, string(body), "hello world")

		wait <- false
	})

	defer testServer.Close()

	strictrequest.StrictRequest("POST", testServer.URL, strictrequest.Options{
		BodyBytes: []byte("hello world"),
	})

	<-wait
}

func TestPassingTimeout(t *testing.T) {
	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(75 * time.Millisecond)
		fmt.Fprintln(w, "ok")
	})

	defer testServer.Close()

	resp, err := strictrequest.StrictRequest("GET", testServer.URL, strictrequest.Options{
		TimeoutMs: 100,
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestFailingTimeout(t *testing.T) {
	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(30 * time.Second)
		fmt.Fprintln(w, "ok")
	})

	start := time.Now().UnixNano()

	resp, err := strictrequest.StrictRequest("GET", testServer.URL, strictrequest.Options{
		TimeoutMs: 100,
	})

	diff := time.Now().UnixNano() - start

	assert.True(t, diff < 101000000)
	assert.Error(t, err)
	assert.Nil(t, resp)

	testServer.CloseClientConnections()
}

func TestNotFollowingRedirections(t *testing.T) {
	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://wikipedia.org", 301)
	})

	defer testServer.Close()

	resp, err := strictrequest.StrictRequest("GET", testServer.URL, strictrequest.Options{})

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 301)
	assert.Equal(t, resp.Header.Get("Location"), "http://wikipedia.org")
}

func TestAllowingRedirections(t *testing.T) {
	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	testProxyServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, testServer.URL, 301)
	})

	defer testServer.Close()
	defer testProxyServer.Close()

	resp, err := strictrequest.StrictRequest("GET", testProxyServer.URL, strictrequest.Options{
		AllowRedirects: true,
	})

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)

	body, err := ioutil.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, "ok\n", string(body))
}

func TestMaxSize(t *testing.T) {
	testServer := CreateTestServer(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})

	defer testServer.Close()

	resp, err := strictrequest.StrictRequest("GET", "http://azer.bike", strictrequest.Options{
		MaxSizeMb: 0.000014,
	})

	assert.Nil(t, err)

	body, err := ioutil.ReadAll(resp.Body)

	assert.Nil(t, err)
	assert.Equal(t, string(body), "<!DOCTYPE html>")
}

func TestIsSameURLDifferentScheme(t *testing.T) {
	assert.True(t, strictrequest.IsSameURLDifferentScheme("http://wikipedia.org", "https://wikipedia.org"))
	assert.True(t, strictrequest.IsSameURLDifferentScheme("https://wikipedia.org", "http://wikipedia.org"))
	assert.False(t, strictrequest.IsSameURLDifferentScheme("http://wikipedia..org", "http://wikipedia.org"))
}
