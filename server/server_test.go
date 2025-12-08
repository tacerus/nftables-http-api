package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tacerus/nftables-http-api/core"
)

type appTest struct {
	c *http.Client
	s *http.Server
}

var at *appTest

// performs a GET request against the live server as opposed to mocking a handler
// returns the response object and the decoded body
func realGet(t *testing.T, path string) (*http.Response, []byte) {
	t.Helper()

	r, err := at.c.Get("http://localhost:8000" + path)
	if err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Error(err)
	}

	return r, b
}

// performs a GET request against a mocked handler
// returns the response object and body
func mockGet(t *testing.T, path string) (*httptest.ResponseRecorder, []byte) {
	t.Helper()

	r := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()

	at.s.Handler.ServeHTTP(rr, r)

	return rr, rr.Body.Bytes()
}

// test for expected response code with common message
func assertStatusEqual(t *testing.T, have int, want int) {
	t.Helper()

	assert.Equalf(t, want, have, "Have status %d, but want status %d", have, want)
}

func allowDestructive() bool {
	return os.Getenv(VAR_DESTRUCTIVE) == "yes"
}

func testDestructive(t *testing.T) {
	if !allowDestructive() {
		t.Skip("Skipping destructive test.")
	}
}

func TestMain(m *testing.M) {
	// TODO: work with capabilities instead of root
	//if allowDestructive() && os.Getuid() != 0 {
	//	os.Exit(1)
	//}

	at = new(appTest)

	app := NewApp(core.Config{
		Bind: "[::1]:8000",
	})
	at.s = app.Start()
	defer at.s.Shutdown(context.Background())

	at.c = &http.Client{}

	cmdNft := "nft"
	cmdFlush := []string{
		"flush", "ruleset",
	}

	if allowDestructive() {
		cmdInit := []*exec.Cmd{
			exec.Command(cmdNft, cmdFlush...),
			exec.Command(
				"nft", "add", "table", "inet", "filter",
			),
			exec.Command(
				"nft", "add", "set", "inet", "filter", "testset6", "{ type ipv6_addr ; flags interval ; }",
			),
			exec.Command(
				"nft", "add", "set", "inet", "filter", "testset4", "{ type ipv4_addr ; flags interval ; }",
			),
		}

		for _, cmd := range cmdInit {
			out, err := cmd.CombinedOutput()
			fmt.Printf("=> command: \"%s\", output: \"%s\"\n", cmd.String(), string(out))
			if err != nil {
				panic(err)
			}
		}
	}

	m.Run()

	if allowDestructive() {
		cmd := exec.Command(cmdNft, cmdFlush...)
		out, err := cmd.CombinedOutput()
		fmt.Printf("=> command: \"%s\", output: \"%s\"\n", cmd.String(), string(out))
		if err != nil {
			panic(err)
		}
	}
}

func TestIndex(t *testing.T) {
	r, _ := mockGet(t, "/")
	assertStatusEqual(t, r.Code, http.StatusNotFound)
}

const (
	VAR_DESTRUCTIVE = "NFT_HTTP_API_TEST_DESTRUCTIVE"

	T_OK               = 0
	T_FAMILY_NOT_EXIST = 1
	T_TABLE_NOT_EXIST  = 2
	T_SET_NOT_EXIST    = 3
)

func TestElementGet(t *testing.T) {
	testDestructive(t)

	r := "/element/"
	testCases := []struct {
		path       string
		expectCode int
		expectBody string
	}{
		{r + "/foo/bar/baz", T_FAMILY_NOT_EXIST, `{"message":"Specified family is not valid."}`},
		{r + "/inet/bar/baz", T_TABLE_NOT_EXIST, `{"message":"Table not found"}`},
		{r + "/inet/filter/baz", T_SET_NOT_EXIST, `{"message":"Set not found"}`},
		{r + "/inet/filter/testset4", T_OK, `{"Elements":[],"Flags":["interval"],"Name":"testset4","Type":"ipv4_addr"}`},
		{r + "/inet/filter/testset6", T_OK, `{"Elements":[],"Flags":["interval"],"Name":"testset6","Type":"ipv6_addr"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			r, b := realGet(t, tc.path)
			switch tc.expectCode {
			case T_FAMILY_NOT_EXIST:
				assertStatusEqual(t, r.StatusCode, http.StatusBadRequest)
			case T_TABLE_NOT_EXIST, T_SET_NOT_EXIST:
				assertStatusEqual(t, r.StatusCode, http.StatusNotFound)
			}

			assert.JSONEq(t, tc.expectBody, string(b))
		})
	}
}
