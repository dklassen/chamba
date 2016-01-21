package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dklassen/chamba/api"
)

var (
	server        *httptest.Server
	reader        io.Reader //Ignore this for now
	signupURL     string
	signinURL     string
	getTokenURL   string
	clearTokenURL string
)

func init() {
	os.Setenv("GOENV", "test")
	server = httptest.NewServer(api.Handlers())              //Creating new server with the user handlers
	signupURL = fmt.Sprintf("%s/signup", server.URL)         //Grab the address for the API endpoint
	signinURL = fmt.Sprintf("%s/signin", server.URL)         //Grab the address for the API endpoint
	getTokenURL = fmt.Sprintf("%s/getToken", server.URL)     //Grab the address for the API endpoint
	clearTokenURL = fmt.Sprintf("%s/clearToken", server.URL) //Grab the address for the API endpoint
}

func tearDown() {
	api.GetDB().Exec("DELETE FROM users;")
}

func setupUser() (expected api.User) {
	expected = api.User{
		FirstName:    "Mark",
		LastName:     "Twain",
		PrimaryEmail: "mark@twain.com",
		Password:     "Huckelberry",
	}

	data := url.Values{}
	data.Add("firstname", expected.FirstName)
	data.Add("lastname", expected.LastName)
	data.Add("email", expected.PrimaryEmail)
	data.Add("password", expected.Password)

	request, _ := http.NewRequest("POST", signupURL, strings.NewReader(data.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err)
	}
	return
}

func checkResponseBody(response *http.Response) string {
	body, _ := ioutil.ReadAll(response.Body)
	return string(body)
}

type HandleTester func(
	method string,
	params url.Values,
) *httptest.ResponseRecorder

func GenerateHandleTester(t *testing.T, handleFunc api.Handler) HandleTester {
	return func(method string, params url.Values) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", strings.NewReader(params.Encode()))
		if err != nil {
			t.Errorf("%v", err)
		}
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		w := httptest.NewRecorder()
		context := &api.AppContext{DB: api.GetDB(), Apikey: "15c035e1cd738ee91910a3d19f93cb92"}
		appHandle := api.AppHandler{AppContext: context, HandlerFunc: handleFunc}
		appHandle.ServeHTTP(w, req)
		return w
	}
}

func GenerateBasicAuthHandleTester(t *testing.T, handleFunc api.Handler, username, password string) HandleTester {
	return func(method string, params url.Values) *httptest.ResponseRecorder {
		req, err := http.NewRequest(method, "", strings.NewReader(params.Encode()))
		if err != nil {
			t.Errorf("%v", err)
		}
		req.Header.Set(
			"Content-Type",
			"application/x-www-form-urlencoded; param=value",
		)
		req.SetBasicAuth(username, password)
		w := httptest.NewRecorder()
		context := &api.AppContext{DB: api.GetDB(), Apikey: "15c035e1cd738ee91910a3d19f93cb92"}
		appHandle := api.AppHandler{AppContext: context, HandlerFunc: handleFunc}
		appHandle.ServeHTTP(w, req)
		return w
	}
}

func TestFilteringRouteHandlers(t *testing.T) {
	var testCases = []struct {
		handler            func(api.Handler) api.Handler
		method             string
		expectedStatusCode int
	}{
		{handler: api.GetOnly, method: "GET", expectedStatusCode: http.StatusOK},
		{handler: api.GetOnly, method: "POST", expectedStatusCode: http.StatusMethodNotAllowed},
		{handler: api.PostOnly, method: "POST", expectedStatusCode: http.StatusOK},
		{handler: api.PostOnly, method: "GET", expectedStatusCode: http.StatusMethodNotAllowed},
	}

	for _, testCase := range testCases {
		getHandler := testCase.handler(func(env *api.AppContext, w http.ResponseWriter, r *http.Request) {
			return
		})
		test := GenerateHandleTester(t, getHandler)
		w := test(testCase.method, url.Values{})
		if w.Code != testCase.expectedStatusCode {
			t.Errorf("Expected %d but got %d when method %s", testCase.expectedStatusCode, w.Code, testCase.method)
		}
	}
}

func TestAuthTokenIsNotExpiredWhenStillInFuture(t *testing.T) {
	now := time.Now()
	oneSecondInTheFuture := now.Add(time.Duration(1) * time.Second)
	expected := api.AuthToken{
		Token:  "aRandomString",
		Expiry: oneSecondInTheFuture,
	}

	if expected.IsExpired() {
		t.Error("Expected auth token to be valid")
	}
}

func TestAuthTokenIsExpiredWhenInPast(t *testing.T) {
	now := time.Now()
	expected := api.AuthToken{
		Token:  "aRandomString",
		Expiry: now,
	}

	if expected.IsExpired() != true {
		t.Error("Expected auth token to be expired")
	}
}

func TestSignInWithValidUser(t *testing.T) {
	expected := setupUser()

	// sign in as the new user getting the auth token we need to make
	// api calls going forward
	request, _ := http.NewRequest("POST", signinURL, nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(expected.PrimaryEmail, expected.Password)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 200 {
		t.Error("Expected status code 200 but got: ", response.StatusCode)
	}

	tearDown()
}

func TestSignUpHandler(t *testing.T) {
	var testCases = []struct {
		FirstName          string
		LastName           string
		Email              string
		Password           string
		ExpectedStatusCode int
		Reason             string
	}{
		{"valid first name", "valid last name", "valid@email.com", "valid password", http.StatusOK, "Should be success"},
		{"valid first name", "valid last name", "valid@email.com", "valid password", http.StatusBadRequest, "Should fail because user exists"},
		{"", "valid last name", "valid@email.com", "valid password", http.StatusBadRequest, "Missing first name"},
		{"valid first name", "", "valid@email.com", "valid password", http.StatusBadRequest, "Missing last name "},
		{"valid first name", "valid last name", "", "valid password", http.StatusBadRequest, "Missing email"},
		{"valid first name", "valid last name", "valid@email.com", "", http.StatusBadRequest, "Missing password"},
		{"", "", "", "", http.StatusBadRequest, "Missing all required fields"},
	}

	for _, testCase := range testCases {
		data := url.Values{}
		data.Add("firstname", testCase.FirstName)
		data.Add("lastname", testCase.LastName)
		data.Add("email", testCase.Email)
		data.Add("password", testCase.Password)

		test := GenerateHandleTester(t, api.Signup)
		w := test("POST", data) // In the full route we filter out GET requests
		if w.Code != testCase.ExpectedStatusCode {
			t.Errorf("Expected %d but got %d reason %s", testCase.ExpectedStatusCode, w.Code, w.Body)
		}
	}
	tearDown()
}

func TestSignupSavesUserToDatabaseAsExpected(t *testing.T) {
	expected := api.User{
		FirstName:    "Mark",
		LastName:     "Twain",
		PrimaryEmail: "mark@twain.com",
		Password:     "Huckelberry",
	}

	data := url.Values{}
	data.Add("email", expected.PrimaryEmail)
	data.Add("firstname", expected.FirstName)
	data.Add("lastname", expected.LastName)
	data.Add("password", expected.Password)

	request, _ := http.NewRequest("POST", signupURL, strings.NewReader(data.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}

	result := api.User{}
	api.GetDB().Where(api.User{PrimaryEmail: expected.PrimaryEmail}).First(&result)

	if expected.FirstName != result.FirstName && expected.PrimaryEmail != expected.PrimaryEmail {
		t.Error("Unable to find saved user expected:", expected, "got:", result)
	}

	tearDown()
}

func TestEnteredEmailAndPasswordsInSignInRoute(t *testing.T) {
	expectedEmail := "bloop@bloop.com"
	expectedPassword := "some password"

	var testingTable = []struct {
		EnteredEmail       string
		EnteredPassword    string
		ExpectedStatusCode int
		Reson              string
	}{
		{expectedEmail, expectedPassword, 200, "Should have all required information"},
		{expectedEmail, "different password", 401, "Should fail because of different password"},
		{"unknownemail@bloop.com", expectedPassword, 401, "Should fail because of unknown email"},
		{"unknownemail@bloop.com", "nope", 401, "Should fail because of unknown email and incorrect password"},
		{"", "", 401, "Should fail because of no email or password"},
	}

	// Create single user in the database
	data := url.Values{}
	data.Add("firstname", "test")
	data.Add("lastname", "test")
	data.Add("email", expectedEmail)
	data.Add("password", expectedPassword)

	request, _ := http.NewRequest("POST", signupURL, strings.NewReader(data.Encode()))
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}

	for _, testCase := range testingTable {
		test := GenerateBasicAuthHandleTester(t, api.BasicAuth(api.Signin), testCase.EnteredEmail, testCase.EnteredPassword)
		w := test("POST", data) // In the full route we filter out GET requests
		if w.Code != testCase.ExpectedStatusCode {
			t.Errorf("Expected %d but got %d reason %s", testCase.ExpectedStatusCode, w.Code, w.Body)
		}
	}
	tearDown()
}

func TestClearTokenRouteReturnsSuccessWhenValidTokenSent(t *testing.T) {
	expected := setupUser()

	// Signin and grab the auth token
	request, _ := http.NewRequest("POST", signinURL, nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(expected.PrimaryEmail, expected.Password)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}
	var jsonResponse map[string]interface{}
	body, err := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(body, &jsonResponse); err != nil {
		t.Error(err)
	}

	json.Unmarshal(body, jsonResponse)
	token := jsonResponse["Token"]

	// Use that token to issue clear request
	request, _ = http.NewRequest("POST", clearTokenURL, nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+token.(string))
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 200 {
		t.Error("Expected status code 200 but got: ", response.StatusCode)
	}

	tearDown()
}

func Test401IsReturnedWhenInvalidTokenIsSent(t *testing.T) {
	request, _ := http.NewRequest("POST", clearTokenURL, nil)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", "Bearer "+"A MADE UP TOKEN")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Error(err)
	}

	if response.StatusCode != 401 {
		t.Error("Expected status code 401 but got: ", response.StatusCode)
	}

	tearDown()
}
