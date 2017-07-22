package person

import "gopkg.in/mgo.v2/bson"

// Person ...
type Person struct {
	ID    bson.ObjectId `bson:"_id"`
	Email string
}
