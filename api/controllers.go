package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/jinzhu/gorm"

	"golang.org/x/crypto/bcrypt"
)

var requiredSignupFields = []string{"firstname", "lastname", "email", "password"}

// AppContext contains state that is passed between requests
type AppContext struct {
	DB     *gorm.DB
	Apikey string
	User   User
}

// AppHandler contains global state for processing the request
type AppHandler struct {
	*AppContext
	HandlerFunc func(*AppContext, http.ResponseWriter, *http.Request)
}

// AuthenticationError an error thrown during authorization
type AuthenticationError struct {
	message string
}

// Handler function takes state of the application and response reader/writers
// for processing a request
type Handler func(env *AppContext,
	w http.ResponseWriter,
	r *http.Request)

func (e AuthenticationError) Error() string {
	return e.message
}

func (e UserExistsError) Error() string {
	return e.message
}

func (token *AuthToken) isExpired() bool {
	return token.Expiry.Before(time.Now())
}

func oneDayFromNow() time.Time {
	return time.Now().Add(time.Duration(86400) * time.Second)
}

func saltPassword(inputPassword string) (string, error) {
	password := []byte(inputPassword)
	saltedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	return string(saltedPassword), err
}

func comparePassword(password, hashedPassword string) (err error) {
	a := []byte(password)
	b := []byte(hashedPassword)
	return bcrypt.CompareHashAndPassword(b, a)
}

func randomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Signup route for a user. Takes and validates the sign in information
// returns user information json blob
func Signup(env *AppContext, w http.ResponseWriter, r *http.Request) {
	missingFields := []string{}
	signupFields := map[string]string{}

	for _, field := range requiredSignupFields {
		value := r.PostFormValue(field)
		if value == "" {
			missingFields = append(missingFields, field)
		} else {
			signupFields[field] = value
		}
	}

	if len(missingFields) != 0 {
		errorMessage := fmt.Sprintf("Signup was missing required fields %q", missingFields)
		log.Error(errorMessage)
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	// ** To make life easy we are going forgo encryption and shadowing
	// decryptedPassword, err := encrypt.AESDecrypt(apiKey, encryptedPassword)
	// if err != nil {
	// 	log.Println(err)
	// 	http.Error(w, "ServerError", http.StatusInternalServerError)
	// 	return
	// }
	//
	saltedPassword, err := saltPassword(signupFields["password"])
	if err != nil {
		log.Error(err)
		http.Error(w, "ServerError", http.StatusInternalServerError)
		return
	}

	user := User{FirstName: signupFields["firstname"],
		LastName:     signupFields["lastname"],
		Password:     saltedPassword,
		PrimaryEmail: signupFields["email"],
	}

	js, err := user.toJSON()
	if err != nil {
		log.Error(err)
		http.Error(w, "ServerError", http.StatusInternalServerError)
		return
	}

	err = user.Save(env.DB)
	if err != nil {
		log.Error(err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func authenticateUser(db *gorm.DB, email, password string) (user User, err error) {
	db.Where(User{PrimaryEmail: email}).First(&user)
	if user.PrimaryEmail == email && comparePassword(password, user.Password) == nil {
		return
	}

	log.Errorf("Authorization denied for: %s", email)
	return user, AuthenticationError{"Authorization denied"}
}

func authenticateTokenUser(db *gorm.DB, token string) (user User, err error) {
	db.Preload("AuthToken", "token = ?", token).First(&user)
	empty := User{}
	if user == empty {
		err = AuthenticationError{"No user found for token"}
		log.Error(err)
	}
	return
}

func parseAuthTokenFromRequest(r *http.Request) (token string, err error) {
	if r.Header["Authorization"] == nil {
		return "", errors.New("Authorization Header was not found")
	}
	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)
	if len(auth) != 2 || strings.EqualFold(auth[0], "Bearer") == false {
		return "", errors.New("No bearer token found")
	}
	return auth[1], nil
}

func parseBasicAuthHeader(r *http.Request) (username, password string, err error) {
	if r.Header["Authorization"] == nil {
		return "", "", errors.New("Authorization Header was not found")
	}
	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 || strings.EqualFold(auth[0], "Basic") == false {
		return "", "", errors.New("Bad syntax for basic auth header")
	}

	payload, _ := base64.StdEncoding.DecodeString(auth[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		log.Println("Header payload was unable to be parsed")
		return "", "", errors.New("Authorization failed")
	}
	return pair[0], pair[1], err
}

// Signin checks user to see if valid auth token if yes generates new auth token and expires
// old
func Signin(env *AppContext, w http.ResponseWriter, r *http.Request) {
	user := env.User // Get the authenticated user
	if user.AuthToken.isExpired() {
		newToken := AuthToken{Token: randomString(20),
			Expiry: oneDayFromNow()}
		env.DB.Model(&user).Update("AuthToken", newToken)
	}

	token := struct {
		Token  string
		Expiry time.Time
	}{
		user.AuthToken.Token,
		user.AuthToken.Expiry,
	}
	js, err := json.Marshal(token)
	if err != nil {
		log.WithFields(log.Fields{
			"action": "signin",
		}).Error(err)
		http.Error(w, "ServerError", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func resetPassword(w http.ResponseWriter, r *http.Request) {
	// do nothing atm
	// require username and secret token
	// THings to do:
	//  need a new password
	// validate current password
	// setup new password
	// save
}

func clearToken(env *AppContext, w http.ResponseWriter, r *http.Request) {
	user := env.User
	if user.AuthToken.isExpired() == false {
		db.Where("token = ?", user.AuthToken.Token).Delete(&AuthToken{})

		// expire the token and update the database
		w.Header().Set("Content-Type", "application/text")
		w.Write([]byte("Token Cleared"))
		return
	}
	http.Error(w, "authorization failed", http.StatusUnauthorized)
	return
}

func authenticateAuthToken(h Handler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		token, err := parseAuthTokenFromRequest(r)
		if err != nil {
			log.Println(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		user, err := authenticateTokenUser(env.DB, token)
		if err != nil {
			log.Println(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		// lookup user with token and attach to authenticated request
		env.User = user
		h(env, w, r)
	}
}

// BasicAuth middleware for using basic auth headers to secure actions
func BasicAuth(h Handler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		email, password, err := parseBasicAuthHeader(r)
		if err != nil {
			log.WithFields(log.Fields{"action": "BasicAuth"}).Error(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		user, err := authenticateUser(env.DB, email, password)
		if err != nil {
			log.WithFields(log.Fields{"action": "BasicAuth"}).Error(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		env.User = user
		h(env, w, r)

	}
}

// PostOnly middleware for filtering non post requests
func PostOnly(h Handler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			log.WithFields(log.Fields{
				"url":         r.URL,
				"http_method": r.Method,
				"datetime":    time.Now(),
			}).Error("POST requests only")
			http.Error(w, "POST requests only", http.StatusMethodNotAllowed)
			return
		}
		h(env, w, r)
	}
}

// GetOnly middleware for filtering non get requests
func GetOnly(h Handler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			log.WithFields(log.Fields{
				"url":         r.URL,
				"http_method": r.Method,
				"datetime":    time.Now(),
			}).Error("GET requests only")
			http.Error(w, "GET requests only", http.StatusMethodNotAllowed)
			return
		}
		h(env, w, r)
	}
}

func checkResponseBody(response *http.Request) string {
	body, _ := ioutil.ReadAll(response.Body)
	return string(body)
}

func (h AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.HandlerFunc(h.AppContext, w, r)
	end := time.Now()
	latency := end.Sub(start)
	log.WithFields(log.Fields{
		"datetime":           start,
		"url":                r.URL,
		"ip":                 r.RemoteAddr,
		"latency_nanesecond": latency.Nanoseconds(),
		"http_user_agent":    r.UserAgent(),
		"http_method":        r.Method,
		"response_header":    w.Header(),
		"request_body":       checkResponseBody(r),
	}).Info("Served request")
}

// Handlers register api routes here
func Handlers() *http.ServeMux {
	context := &AppContext{DB: GetDB(), Apikey: "15c035e1cd738ee91910a3d19f93cb92"}
	mux := http.NewServeMux()
	mux.Handle("/signup", AppHandler{context, PostOnly(Signup)})
	mux.Handle("/signin", AppHandler{context, PostOnly(BasicAuth(Signin))})
	mux.Handle("/clearToken", AppHandler{context, PostOnly(authenticateAuthToken(clearToken))})
	return mux
}
