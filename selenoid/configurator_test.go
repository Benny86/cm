package selenoid

import (
	"fmt"
	. "github.com/aandryashin/matchers"
	"github.com/aandryashin/selenoid/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var (
	mock *httptest.Server
)

func init() {
	mock = httptest.NewServer(mux())
	os.Setenv("DOCKER_HOST", "tcp://"+hostPort(mock.URL))
}

func mux() http.Handler {
	mux := http.NewServeMux()

	//Docker Registry API mock
	mux.HandleFunc("/v2/", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))
	mux.HandleFunc("/v2/selenoid/firefox/tags/list", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprintln(w, `{"name":"firefox", "tags": ["46.0", "45.0", "47.0", "latest"]}`)
		},
	))

	mux.HandleFunc("/v2/selenoid/phantomjs/tags/list", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprintln(w, `{"name":"phantomjs", "tags": ["2.1.1", "latest"]}`)
		},
	))

	//Docker API mock
	mux.HandleFunc("/v1.29/images/create", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	))
	return mux
}

func hostPort(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		panic(err)
	}
	return u.Host
}

func TestImageWithTag(t *testing.T) {
	AssertThat(t, imageWithTag("selenoid/firefox", "tag"), EqualTo{"selenoid/firefox:tag"})
}

func TestFetchImageTags(t *testing.T) {
	c := Configurator{
		RegistryUrl: mock.URL,
		Verbose:     true,
	}
	err := c.Init()
	defer c.Close()
	AssertThat(t, err, Is{nil})
	tags := c.fetchImageTags("selenoid/firefox")
	AssertThat(t, len(tags), EqualTo{3})
	AssertThat(t, tags[0], EqualTo{"47.0"})
	AssertThat(t, tags[1], EqualTo{"46.0"})
	AssertThat(t, tags[2], EqualTo{"45.0"})
}

func TestPullImages(t *testing.T) {
	c := Configurator{
		RegistryUrl: mock.URL,
		Verbose:     true,
	}
	err := c.Init()
	defer c.Close()
	AssertThat(t, err, Is{nil})
	tags := c.pullImages("selenoid/firefox", []string{"46.0", "45.0"})
	AssertThat(t, len(tags), EqualTo{2})
	AssertThat(t, tags[0], EqualTo{"46.0"})
	AssertThat(t, tags[1], EqualTo{"45.0"})
}

func TestCreateConfig(t *testing.T) {
	testCreateConfig(t, true)
}

func TestLimitNoPull(t *testing.T) {
	testCreateConfig(t, false)
}

func testCreateConfig(t *testing.T, pull bool) {
	c := Configurator{
		RegistryUrl: mock.URL,
		Limit:       2,
		Pull:        pull,
		Verbose:     true,
	}
	err := c.Init()
	defer c.Close()
	AssertThat(t, err, Is{nil})
	cfg := c.createConfig()
	AssertThat(t, len(cfg), EqualTo{2})

	firefoxVersions, hasFirefoxKey := cfg["firefox"]
	AssertThat(t, hasFirefoxKey, Is{true})
	AssertThat(t, firefoxVersions, Is{Not{nil}})

	correctFFBrowsers := make(map[string]*config.Browser)
	correctFFBrowsers["47.0"] = &config.Browser{
		Image: "selenoid/firefox:47.0",
		Port:  "4444",
		Path:  "/wd/hub",
	}
	correctFFBrowsers["46.0"] = &config.Browser{
		Image: "selenoid/firefox:46.0",
		Port:  "4444",
		Path:  "/wd/hub",
	}
	AssertThat(t, firefoxVersions, EqualTo{config.Versions{
		Default:  "47.0",
		Versions: correctFFBrowsers,
	}})

	phantomjsVersions, hasPhantomjsKey := cfg["phantomjs"]
	AssertThat(t, hasPhantomjsKey, Is{true})
	AssertThat(t, phantomjsVersions, Is{Not{nil}})
	AssertThat(t, phantomjsVersions.Default, EqualTo{"2.1.1"})

	correctPhantomjsBrowsers := make(map[string]*config.Browser)
	correctPhantomjsBrowsers["2.1.1"] = &config.Browser{
		Image: "selenoid/phantomjs:2.1.1",
		Port:  "4444",
	}
}