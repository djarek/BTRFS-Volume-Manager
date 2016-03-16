package main

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const dbName = "btrfs"
const usersCollectionName = "users"

var (
	connected = false
	session   *mgo.Session
	db        *mgo.Database
	usersRepo UsersRepository
)

// UsersRepository is a collection of users
type UsersRepository struct {
	coll *mgo.Collection
}

// FindUserByUsername provides searching for user by username.
// It returns one User and error.
func (repo UsersRepository) FindUserByUsername(username string) (User, error) {
	result := User{}
	err := repo.coll.Find(bson.M{"username": username}).One(&result)
	return result, err
}

// FindUsersByFirstName provides searching for users by first name.
// It returns array of Users and error.
func (repo UsersRepository) FindUsersByFirstName(
	firstName string) ([]User, error) {
	var result []User
	err := repo.coll.Find(bson.M{"firstName": firstName}).All(&result)
	return result, err
}

// FindUsersByLastName provides searching for users by last name.
// It returns array of Users and error.
func (repo UsersRepository) FindUsersByLastName(
	lastName string) ([]User, error) {
	var result []User
	err := repo.coll.Find(bson.M{"lastName": lastName}).All(&result)
	return result, err
}

// FindUsersByRegistrationDate provides sarching for users by registration date.
// It returns array of Users and error.
func (repo UsersRepository) FindUsersByRegistrationDate(
	registrationDate time.Time) ([]User, error) {
	var result []User
	err := repo.coll.Find(bson.M{
		"registrationDate": registrationDate}).All(&result)
	return result, err
}

func startDB() {
	log.Println("Connecting to Database")
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	db = session.DB(dbName)
	connected = true
	session.SetMode(mgo.Monotonic, true)
	usersRepo.coll = session.DB(dbName).C(usersCollectionName)

	// Unique index
	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = usersRepo.coll.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// Initialize data base if it is empty
	var results []User
	err = usersRepo.coll.Find(nil).All(&results)
	if len(results) == 0 {
		initializeDB()
	}
}

// Function that closes database connection
func stopDB() {
	log.Println("Closing databse connection")
	session.Close()
	connected = false
}

// Function that adds an admin user
func initializeDB() {
	id := bson.NewObjectId()
	password := []byte("admin")
	hashedPassword, err := bcrypt.GenerateFromPassword(
		password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	err = usersRepo.coll.Insert(
		&User{
			ID:               id,
			Username:         "admin",
			HashedPassword:   string(hashedPassword),
			FirstName:        "Jo",
			LastName:         "Doe",
			RegistrationDate: time.Now()})
	if err != nil {
		panic(err)
	}
	log.Println("Initialized database. Added admin")
}

// Funtion that drops entire DB
func dropDB() {
	err := session.DB(dbName).DropDatabase()
	if err != nil {
		panic(err)
	}
	log.Println("Droped database")
}
