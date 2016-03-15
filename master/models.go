package main

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

// User model
type User struct {
	ID               bson.ObjectId `bson:"_id,omitempty"`
	Username         string        `bson:"username,omitempty"`
	HashedPassword   string        `bson:"hashedPassword,omitempty"`
	FirstName        string        `bson:"firstName"`
	LastName         string        `bson:"lastName"`
	RegistrationDate time.Time     `bson:"registrationDate"`
}

// StorageServer represents a Network Attached Storage device
type StorageServer struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `bson:"name"`
}

// BlockDevice represents a block device retrieved by blkid probe
type BlockDevice struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	VolID bson.ObjectId `bson:"volID"` //can be empty
	Path  string        `bson:"path,omitempty"`
	UUID  string        `bson:"uuid,omitempty"`
	Type  string        `bson:"type,omitempty"`
}

// BtrfsVolume represents a filesystem volume which can potentially span over
// multiple devices
type BtrfsVolume struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	ServID bson.ObjectId `bson:"servID"` // can be empty
	Label  string        `bson:"label"`
}
