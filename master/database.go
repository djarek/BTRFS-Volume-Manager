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
	connected bool = false
	session   *mgo.Session
	collUsers *mgo.Collection
	db        *mgo.Database
)

func findByUsername(username string) (User, error) {
	result := User{}
	err := collUsers.Find(bson.M{"username": username}).One(&result)
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
	collUsers = session.DB(dbName).C(usersCollectionName)

	// Unique index
	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = collUsers.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// Initialize data base if it is empty
	var results []User
	err = collUsers.Find(nil).All(&results)
	if len(results) == 0 {
		initializeDB()
	}
}

func stopDB() {
	log.Println("Closing Databse connection")
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
	err = collUsers.Insert(
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
