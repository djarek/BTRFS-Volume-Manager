package db

import (
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/djarek/btrfs-volume-manager/master/models"
)

const dbName = "btrfs"
const usersCollectionName = "users"

var (
	connected = false
	session   *mgo.Session
	db        *mgo.Database
	UsersRepo UsersRepository
)

// UsersRepository is a collection of users
type UsersRepository struct {
	coll *mgo.Collection
}

// FindUserByUsername provides searching for user by username.
// It returns one User and error.
func (repo UsersRepository) FindUserByUsername(username string) (models.User, error) {
	result := models.User{}
	err := repo.coll.Find(bson.M{"username": username}).One(&result)
	return result, err
}

// Function that connects database and basically all necessary initialization
// processes.
func StartDB() {
	log.Println("Connecting to Database")
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	db = session.DB(dbName)
	connected = true
	session.SetMode(mgo.Monotonic, true)
	UsersRepo.coll = session.DB(dbName).C(usersCollectionName)

	// Unique index
	index := mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = UsersRepo.coll.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	// Initialize data base if it is empty
	var results []models.User
	err = UsersRepo.coll.Find(nil).All(&results)
	if len(results) == 0 {
		initializeDB()
	}
}

// Function that closes database connection.
func StopDB() {
	log.Println("Closing databse connection")
	session.Close()
	connected = false
}

// Function that adds an admin user.
func initializeDB() {
	id := bson.NewObjectId()
	password := []byte("admin")
	hashedPassword, err := bcrypt.GenerateFromPassword(
		password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	err = UsersRepo.coll.Insert(
		&models.User{
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

// Funtion that drops entire database.
func DropDB() {
	err := session.DB(dbName).DropDatabase()
	if err != nil {
		panic(err)
	}
	log.Println("Dropped database")
}
