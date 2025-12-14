package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tacerus/nftables-http-api/core"
	nftapi "github.com/tacerus/nftables-http-api/nftables"
)

const (
	NFT = "nft"
	URL = "http://localhost:8000"

	VAR_DESTRUCTIVE = "NFT_HTTP_API_TEST_DESTRUCTIVE"

	T_OK               = 0
	T_FAMILY_NOT_EXIST = 1
	T_TABLE_NOT_EXIST  = 2
	T_SET_NOT_EXIST    = 3

	T_BODY_NOT_AUTHORIZED = `{"message":"The supplied token is not authorized to perform this request."}`
	T_BODY_INVALID_TOKEN  = `{"message":"Invalid token."}`
	T_BODY_MISSING_TOKEN  = `{"message":"Missing token."}`
)

var (
	fixtureSets = []struct {
		name  string
		stype string
		flags []string
		// elements shall be listed in the same format as printed by nft
		elements []string
	}{
		{
			"testset4_empty",
			"ipv4_addr",
			[]string{"interval"},
			[]string{},
		},
		{
			"testset6_empty",
			"ipv6_addr",
			[]string{"interval"},
			[]string{},
		},
		{
			"testset4_mixaddrs",
			"ipv4_addr",
			[]string{"interval"},
			[]string{
				"127.0.0.1",
				"192.0.2.16/28",
				"192.168.4.2-192.168.4.50",
			},
		},
		{
			"testset4_open",
			"ipv4_addr",
			[]string{"interval"},
			[]string{
				"128.0.0.0/1",
			},
		},
		{
			"testset6_mixaddrs",
			"ipv6_addr",
			[]string{"interval"},
			[]string{
				"2001:db8:a1:11::/64",
				"2001:db8:a2:11::100",
				"2001:db8:100::/48",
				"2a02:1748:f7df:9c80::/64",
				"2001:db8:200:a::100-2001:db8:200:a::150",
			},
		},
	}

	// TODO: also test fine grained paths
	fixtureTokens = core.ConfigTokens{
		"$2y$05$4j6cgtb28xMeoVdlIF9XVOaTJlvux89oUo5GIEr2LdJNjYPkVz.HK": core.ConfigTokenPaths{ // thisTokenIsAuthorized
			"/set/*": []string{"GET"},
		},
		"$2y$05$ZldSvJppBs5o3E.N2Jf8vO8AtVCNmJtQsjZKj81aFld77KK38SvP.": core.ConfigTokenPaths{ // ICanPut
			"/set/*": []string{"PUT"},
		},
	}
)

type appTest struct {
	c *http.Client
	s *http.Server
}

var at *appTest

// performs a GET request against the live server as opposed to mocking a handler
// returns the response object and the decoded body
func realGet(t *testing.T, path string, token string) (*http.Response, []byte) {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, URL+path, nil)
	if err != nil {
		t.Fatalf("Failed to construct HTTP request for testing: %v", err)
	}

	if token != "" {
		request.Header.Set(TOKEN_HEADER, token)
	}

	response, err := at.c.Do(request)
	if err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}

	return response, b
}

// performs a PUT request against the live server as opposed to mocking a handler
// returns the response object and the decoded body
func realPut(t *testing.T, path string, token string, payload interface{}) (*http.Response, []byte) {
	t.Helper()

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(payload)

	request, err := http.NewRequest(http.MethodPut, URL+path, buf)
	if err != nil {
		t.Fatalf("Failed to construct HTTP request for testing: %v", err)
	}

	if token != "" {
		request.Header.Set(TOKEN_HEADER, token)
	}

	response, err := at.c.Do(request)
	if err != nil {
		t.Error(err)
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error(err)
	}

	return response, b
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

type TestSets struct {
	name   string
	family int
	flags  []string
}

func TestMain(m *testing.M) {
	// TODO: work with capabilities instead of root
	//if allowDestructive() && os.Getuid() != 0 {
	//	os.Exit(1)
	//}

	at = new(appTest)

	app := NewApp(core.Config{
		Bind:   "[::1]:8000",
		Tokens: fixtureTokens,
	})
	at.s = app.Start()
	defer at.s.Shutdown(context.Background())

	at.c = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

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
		}

		nftCmdAdd := []string{"add"}
		nftCmdAddElement := slices.Concat(nftCmdAdd, []string{"element"})
		nftCmdAddSet := slices.Concat(nftCmdAdd, []string{"set"})
		nftCmdInetFilter := []string{"inet", "filter"}

		for _, s := range fixtureSets {
			cmdInit = append(cmdInit, exec.Command(
				NFT,
				slices.Concat(nftCmdAddSet, nftCmdInetFilter, []string{
					s.name,
					"{ type " + s.stype + "; flags " + strings.Join(s.flags, ", ") + " ; }",
				})...,
			))

			if len(s.elements) > 0 {
				cmdInit = append(cmdInit, exec.Command(
					NFT,
					slices.Concat(nftCmdAddElement, nftCmdInetFilter, []string{
						s.name,
						"{ " + strings.Join(s.elements, ", ") + " }",
					})...,
				))
			}
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

type testCase struct {
	path       string
	expectCode int
	expectBody string
}

type testCaseSet struct {
	testCase

	payload nftapi.Set
}

func TestSetGet(t *testing.T) {
	testDestructive(t)

	r := "/set/"
	testCases := []testCase{
		{r + "foo/bar/baz", T_FAMILY_NOT_EXIST, `{"message":"Specified family is not valid."}`},
		{r + "inet/bar/baz", T_TABLE_NOT_EXIST, `{"message":"Table not found"}`},
		{r + "inet/filter/baz", T_SET_NOT_EXIST, `{"message":"Set not found"}`},
	}

	for _, s := range fixtureSets {

		// TODO: try to order elements in the same way as returned by nft instead
		slices.Sort(s.elements)

		// generate response bodies like
		//   {"Elements":[],"Flags":["interval"],"Name":"testset4_empty","Type":"ipv4_addr"}
		b, err := json.Marshal(nftapi.Set{
			Elements: s.elements,
			Flags:    s.flags,
			Name:     s.name,
			Type:     s.stype,
		})
		if err != nil {
			t.Fatalf("Failure constructing JSON for testing: %v", err)
		}

		testCases = append(testCases, testCase{
			r + "inet/filter/" + s.name, T_OK, string(b),
		})
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			for _, token := range []string{
				"",
				"thisTokenIsBogus",
				"thisTokenIsAuthorized", // $2y$05$4j6cgtb28xMeoVdlIF9XVOaTJlvux89oUo5GIEr2LdJNjYPkVz.HK
			} {
				r, b := realGet(t, tc.path, token)

				if token != "thisTokenIsAuthorized" {
					assertStatusEqual(t, r.StatusCode, http.StatusUnauthorized)
					if token == "" {
						assert.JSONEq(t, T_BODY_MISSING_TOKEN, string(b))
					} else if token == "thisTokenIsBogus" {
						assert.JSONEq(t, T_BODY_INVALID_TOKEN, string(b))
					} else {
						assert.JSONEq(t, T_BODY_NOT_AUTHORIZED, string(b))
					}
					continue
				}

				switch tc.expectCode {
				case T_FAMILY_NOT_EXIST:
					assertStatusEqual(t, r.StatusCode, http.StatusBadRequest)
				case T_TABLE_NOT_EXIST, T_SET_NOT_EXIST:
					assertStatusEqual(t, r.StatusCode, http.StatusNotFound)
				case T_OK:
					assertStatusEqual(t, r.StatusCode, http.StatusOK)
				}

				assert.JSONEq(t, tc.expectBody, string(b))
			}
		})
	}
}

func TestSetPut(t *testing.T) {
	testDestructive(t)

	r := "/set/"
	testCases := []testCaseSet{
		{testCase: testCase{r + "foo/bar/baz", T_FAMILY_NOT_EXIST, `{"message":"Specified family is not valid."}`}},
		{testCase: testCase{r + "inet/bar/baz", T_TABLE_NOT_EXIST, `{"message":"Table not found"}`}},
	}

	for _, s := range fixtureSets {
		// generate request bodies like
		//   {"Elements":[],"Flags":["interval"],"Name":"testset4_empty","Type":"ipv4_addr"}

		testCases = append(testCases, testCaseSet{
			testCase: testCase{
				path:       r + "inet/filter/" + s.name,
				expectCode: T_OK,
				expectBody: `{"message":"ok"}`,
			},
			payload: nftapi.Set{
				Elements: s.elements,
				Flags:    s.flags,
				Name:     s.name,
				Type:     s.stype,
			},
		})
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			for _, token := range []string{
				"",
				"thisTokenIsBogus",
				"thisTokenIsAuthorized", // $2y$05$4j6cgtb28xMeoVdlIF9XVOaTJlvux89oUo5GIEr2LdJNjYPkVz.HK
				"ICanPut",               // $2y$05$ZldSvJppBs5o3E.N2Jf8vO8AtVCNmJtQsjZKj81aFld77KK38SvP.
			} {
				fmt.Printf("testing token %s\n", token)
				r, b := realPut(t, tc.path, token, tc.payload)

				if token != "ICanPut" {
					assertStatusEqual(t, r.StatusCode, http.StatusUnauthorized)
					if token == "" {
						assert.JSONEq(t, T_BODY_MISSING_TOKEN, string(b))
					} else if token == "thisTokenIsBogus" {
						assert.JSONEq(t, T_BODY_INVALID_TOKEN, string(b))
					} else {
						assert.JSONEq(t, T_BODY_NOT_AUTHORIZED, string(b))
					}
					continue
				}

				switch tc.expectCode {
				case T_FAMILY_NOT_EXIST:
					assertStatusEqual(t, r.StatusCode, http.StatusBadRequest)
				case T_TABLE_NOT_EXIST, T_SET_NOT_EXIST:
					assertStatusEqual(t, r.StatusCode, http.StatusNotFound)
				case T_OK:
					assertStatusEqual(t, r.StatusCode, http.StatusOK)
				}

				assert.JSONEq(t, tc.expectBody, string(b))
			}
		})
	}
}
