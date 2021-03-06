package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"

	"github.com/dustywilson/dandy/person"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Default values
var (
	mgoHosts = "127.0.0.1"
	addr     = "0.0.0.0:8080"
)

// errors
var (
	ErrBadRequest          = errors.New("bad argument or request")
	ErrInternalServerError = errors.New("internal server error")
)

func init() {
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}
	flag.StringVar(&addr, "addr", addr, "HTTP service listens on this address")

	if v := os.Getenv("MGO"); v != "" {
		mgoHosts = v
	}
	flag.StringVar(&mgoHosts, "mgo", mgoHosts, "MongoDB hosts, comma separated")

	flag.Parse()
}

func main() {
	mgoSession, err := mgo.Dial(mgoHosts)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	defer mgoSession.Close()

	personSvc, err := person.NewService(mgoSession.DB("person"))
	if err != nil {
		println(err)
		os.Exit(1)
	}

	r := mux.NewRouter()

	r.Methods("POST").Path("/person/").Handler(httptransport.NewServer(
		person.CreateEndpoint(personSvc),
		decodeCreatePersonRequest,
		encodeIDResponse,
	))

	r.Methods("GET").Path("/person/email:{email}").Handler(httptransport.NewServer(
		person.FindByEmailEndpoint(personSvc),
		decodeGetPersonByEmailRequest,
		encodeJSONResponse,
	))

	r.Methods("GET").Path("/person/{id:[0-9A-Za-z]{12}}").Handler(httptransport.NewServer(
		person.FindByIDEndpoint(personSvc),
		decodeSimpleIDRequest,
		encodeJSONResponse,
	))

	r.Methods("PUT").Path("/person/{id:[0-9A-Za-z]{12}}").Handler(httptransport.NewServer(
		person.UpdateEndpoint(personSvc),
		decodeUpdatePersonRequest,
		encodeIDResponse,
	))

	r.Methods("DELETE").Path("/person/{id:[0-9A-Za-z]{12}}").Handler(httptransport.NewServer(
		person.DeleteEndpoint(personSvc),
		decodeSimpleIDRequest,
		encodeStringResponse,
	))

	r.Methods("GET").Path("/person").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/person/", http.StatusSeeOther)
	})

	r.Methods("GET").PathPrefix("/person/").Handler(http.StripPrefix("/person/", http.FileServer(http.Dir("../../person/www"))))

	println(http.ListenAndServe(addr, r))
	os.Exit(1)
}

func decodeSimpleIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	idstr, ok := vars["id"]
	if !ok || !bson.IsObjectIdHex(idstr) {
		return nil, ErrBadRequest
	}
	id := bson.ObjectIdHex(idstr)
	if !id.Valid() {
		return nil, ErrBadRequest
	}

	return id, nil
}

func decodeCreatePersonRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	var request person.Person
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return nil, err
	}
	request.ID = ""
	return request, nil
}

func decodeGetPersonByEmailRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	email, ok := vars["email"]
	if !ok {
		return nil, ErrBadRequest
	}

	return email, nil
}

func decodeUpdatePersonRequest(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()

	vars := mux.Vars(r)
	idstr, ok := vars["id"]
	if !ok || !bson.IsObjectIdHex(idstr) {
		return nil, ErrBadRequest
	}
	id := bson.ObjectIdHex(idstr)
	if !id.Valid() {
		return nil, ErrBadRequest
	}

	var request person.Person
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return nil, err
	}
	request.ID = id
	return request, nil
}

func encodeIDResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	id, ok := response.(bson.ObjectId)
	if !ok {
		return ErrInternalServerError
	}

	w.Write([]byte(id.Hex()))
	return nil
}

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func encodeStringResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	str, ok := response.(string)
	if !ok {
		return ErrInternalServerError
	}

	w.Write([]byte(str))
	return nil
}
