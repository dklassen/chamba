package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"golang.org/x/crypto/bcrypt"
)

var requiredSignupFields = []string{"firstname", "lastname", "email", "password"}

// AuthenticatedRequest represents an a request that has a verified user associated
// in the db. Here we take a request with basic auth and generate a request with the
// user retrieved from the database
type AuthenticatedRequest struct {
	*http.Request
	User User
}

// AuthToken Represents the auth token we use for authorization
type AuthToken struct {
	ID        uint   `gorm:"primary_key"`
	UserID    int    `sql:"index"`
	Token     string `sql:"not null"`
	Expiry    time.Time
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User represents a chamba user
type User struct {
	ID           uint   `gorm:"primary_key"`
	FirstName    string `sql:"not null"`
	LastName     string `sql:"not null"`
	UserName     string `sql:"not null"`
	PrimaryEmail string `sql:"not null;unique"`
	Password     string `sql:"not null;unique"`
	AuthToken    AuthToken
	DeletedAt    *time.Time
	CreatedAt    time.Time `sql:"not null;unique"`
	UpdatedAt    time.Time
}

func (user *User) toJSON() (js []byte, err error) {
	return json.Marshal(user)
}

// Exists checks that a currect user struct can be saved to the database
// at the moment the only restriction is the PrimaryEmail field which must be unique
// Per user.
func (user User) Exists(db *gorm.DB) (exists bool) {
	checkUser := User{}
	empty := User{}
	db.Where(&User{PrimaryEmail: user.PrimaryEmail}).First(&checkUser)
	return checkUser != empty
}

// Save the user to the database after a few checks to make sure we can do so
func (user User) Save(db *gorm.DB) (err error) {
	if user.Exists(db) {
		err = UserExistsError{fmt.Sprintf("Unable to save user with PrimaryEmail %s already exists in database", user.PrimaryEmail)}
		return
	}
	db.Save(&user)
	return
}

// AppContext contains state that is passed between requests
type AppContext struct {
	DB     *gorm.DB
	Apikey string
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

// UserExistsError user struct already exists in the database
type UserExistsError struct {
	message string
}

// Handler function takes state of the application and response reader/writers
// for processing a request
type Handler func(env *AppContext,
	w http.ResponseWriter,
	r *http.Request)

// Authenticated Handler has a request containing the retrieved user
// information after a auth request
type authenticatedHandler func(env *AppContext,
	w http.ResponseWriter,
	r *AuthenticatedRequest)

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
		log.Println(errorMessage)
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
		http.Error(w, "ServerError", http.StatusInternalServerError)
		return
	}

	err = user.Save(env.DB)
	if err != nil {
		log.Printf(err.Error())
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// authenticateUser compare the posted email and password against
// user in the database
func authenticateUser(db *gorm.DB, email, password string) (user User, err error) {
	db.Where(User{PrimaryEmail: email}).First(&user)
	if user.PrimaryEmail == email && comparePassword(password, user.Password) == nil {
		return
	}
	return user, AuthenticationError{"Authorization denied"}
}

func authenticateTokenUser(db *gorm.DB, token string) (user User, err error) {
	db.Preload("AuthToken", "token = ?", token).First(&user)
	empty := User{}
	if user == empty {
		err = AuthenticationError{"Authorization denied"}
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

func Signin(env *AppContext, w http.ResponseWriter, r *AuthenticatedRequest) {
	user := r.User // Get the authenticated user
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
		log.Println(err)
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

func clearToken(env *AppContext, w http.ResponseWriter, r *AuthenticatedRequest) {
	user := r.User
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

func authenticateAuthToken(h authenticatedHandler) Handler {
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
		basicAuthResponse := AuthenticatedRequest{r, user}
		h(env, w, &basicAuthResponse)
	}
}

// BasicAuth middleware for using basic auth headers to secure endpoints
func BasicAuth(h authenticatedHandler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		email, password, err := parseBasicAuthHeader(r)
		if err != nil {
			log.Println(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		user, err := authenticateUser(env.DB, email, password)
		if err != nil {
			log.Println(err)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		basicAuthResponse := AuthenticatedRequest{r, user}
		h(env, w, &basicAuthResponse)
	}
}

// PostOnly middleware for filtering non post requests
func PostOnly(h Handler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			h(env, w, r)
			return
		}
		http.Error(w, "POST requests only", http.StatusMethodNotAllowed)
	}
}

// GetOnly middleware for filtering non get requests
func GetOnly(h Handler) Handler {
	return func(env *AppContext, w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			h(env, w, r)
			return
		}
		http.Error(w, "GET requests only", http.StatusMethodNotAllowed)
	}
}

func (h AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandlerFunc(h.AppContext, w, r)
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
