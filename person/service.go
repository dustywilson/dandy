package person

import (
	"context"
	"errors"
	"log"

	"github.com/go-kit/kit/endpoint"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Errors
var (
	ErrBadRequest = errors.New("bad argument or request")
)

// Service provides the interface to Person-related methods
type Service interface {
	Create(ctx context.Context, person Person) (id bson.ObjectId, err error)
	FindByID(ctx context.Context, id bson.ObjectId) (person Person, err error)
	FindByEmail(ctx context.Context, email string) (person Person, err error)
	Update(ctx context.Context, person Person) (id bson.ObjectId, err error)
	Delete(ctx context.Context, id bson.ObjectId) (err error)
}

type service struct {
	people *mgo.Collection
}

// NewService returns a Service
func NewService(db *mgo.Database) Service {
	s := &service{
		people: db.C("people"),
	}
	if err := s.people.EnsureIndex(mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	}); err != nil {
		log.Fatal(err)
	}
	return s
}

func (s *service) Create(_ context.Context, person Person) (id bson.ObjectId, err error) {
	person.ID = bson.NewObjectId()
	err = s.people.Insert(person)
	if err == nil {
		id = person.ID
	}
	return
}

func CreateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(Person)
		if !ok {
			return nil, ErrBadRequest
		}
		req.ID = ""
		id, err := svc.Create(ctx, req)
		return id, err
	}
}

func (s *service) FindByID(_ context.Context, id bson.ObjectId) (person Person, err error) {
	if !id.Valid() {
		err = mgo.ErrNotFound
		return
	}
	err = s.people.FindId(id).One(&person)
	return
}

func FindByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(bson.ObjectId)
		if !ok {
			return nil, ErrBadRequest
		}
		person, err := svc.FindByID(ctx, req)
		return person, err
	}
}

func (s *service) FindByEmail(_ context.Context, email string) (person Person, err error) {
	err = s.people.Find(bson.M{"email": email}).One(&person)
	log.Printf("FindByEmail: %+v", person)
	return
}

func FindByEmailEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(string)
		if !ok {
			return nil, ErrBadRequest
		}
		person, err := svc.FindByEmail(ctx, req)
		return person, err
	}
}

func (s *service) Update(_ context.Context, person Person) (id bson.ObjectId, err error) {
	if !person.ID.Valid() {
		err = mgo.ErrNotFound
		return
	}
	err = s.people.UpdateId(person.ID, person)
	if err == nil {
		id = person.ID
	}
	return
}

func UpdateEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(Person)
		if !ok {
			return nil, ErrBadRequest
		}
		person, err := svc.Update(ctx, req)
		return person, err
	}
}

func (s *service) Delete(_ context.Context, id bson.ObjectId) (err error) {
	if !id.Valid() {
		err = mgo.ErrNotFound
		return
	}
	err = s.people.RemoveId(id)
	return
}

func DeleteEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(bson.ObjectId)
		if !ok {
			return nil, ErrBadRequest
		}
		err := svc.Delete(ctx, req)
		if err != nil {
			return nil, err
		}

		return "OK", nil
	}
}
