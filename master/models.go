package main

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User model prototype without hashing algorithms implemented yet
type User struct {
	ID               bson.ObjectId `bson:"_id,omitempty"`
	Username         string        `bson:"username,omitempty"`
	HashedPassword   string        `bson:"hashedPassword,omitempty"`
	FirstName        string        `bson:"firstName"`
	LastName         string        `bson:"lastName"`
	RegistrationDate time.Time     `bson:"registrationDate"`
}

//StorageServer represents a Network Attached Storage device
type StorageServer struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `bson:"name"`
}

//BlockDevice represents a block device retrieved by blkid probe
type BlockDevice struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	VolID bson.ObjectId `bson:"volID"` //can be empty
	Path  string        `bson:"path,omitempty"`
	UUID  string        `bson:"uuid,omitempty"`
	Type  string        `bson:"type,omitempty"`
}

//BtrfsVolume represents a filesystem volume which can potentially span over
//multiple devices
type BtrfsVolume struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	ServID bson.ObjectId `bson:"servID"` // can be empty
	Label  string        `bson:"label"`
}

// Simply type for authentication process
type LoginAndPassword struct {
	Username string
	Password string
}

func findByUsername(username string) (User, error) {
	result := User{}
	err := collUsers.Find(bson.M{"username": username}).One(&result)

	return result, err
}

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
}

func dropDB(database *mgo.Database) {
	err := database.DropDatabase()
	if err != nil {
		panic(err)
	}
}
