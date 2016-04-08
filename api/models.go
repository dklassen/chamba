package api

// AuthToken Represents the auth token we use for authorization
import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type AuthToken struct {
	gorm.Model
	UserID int    `sql:"index"`
	Token  string `sql:"not null"`
	Expiry time.Time
}

// User represents a chamba user
type User struct {
	gorm.Model
	FirstName    string `sql:"not null"`
	LastName     string `sql:"not null"`
	UserName     string `sql:"not null"`
	PrimaryEmail string `sql:"not null;unique"`
	Password     string `sql:"not null;unique"`
	Type         string
	FarmID       uint
	Address      Address
	Category     string
	AuthToken    AuthToken
}

// Address is a physical location on the earth
type Address struct {
	gorm.Model
	FarmID          uint
	UserID          uint
	Latitude        int
	Longitude       int
	City            string
	PostalOrZipCode string
	ProvinceOrState string
}

// Review table holds reviews about a farm or work experience
type Review struct {
	gorm.Model
	Stars   int
	Comment string
}

// Task is a unit of work completed
type Task struct {
	Status bool
}

type Crop struct {
	gorm.Model
	FarmID uint
	Name   string
}

// Farm represents a chamba farm where users can work
type Farm struct {
	gorm.Model
	Owner       User // the chamba user associated with the farm
	Name        string
	Description string
	Crops       []Crop
	Address     Address
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

// UserExistsError user struct already exists in the database
type UserExistsError struct {
	message string
}

// Save the user to the database after a few checks to make sure we can do so
func (user User) Save(db *gorm.DB) (err error) {
	if user.Exists(db) {
		return UserExistsError{fmt.Sprintf("Unable to save user with PrimaryEmail %s already exists in database", user.PrimaryEmail)}
	}
	return db.Save(&user).Error
}
