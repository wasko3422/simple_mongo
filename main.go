package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Name    string        `json:"name" bson:"name"`
	Surname string        `json:"surname" bson:"surname"`
	ID      bson.ObjectId `json:"id" bson:"_id"`
}

var session *mgo.Session

func init() {
	session = mongo()
}

func main() {
	r := httprouter.New()
	r.GET("/user/:id", getUser)
	r.POST("/user", postUser)
	r.DELETE("/user/:id", deleteUser)
	r.PUT("/user/:id", updateUser)

	http.ListenAndServe("localhost:8080", r)
}

func mongo() *mgo.Session {
	s, err := mgo.Dial("mongodb://localhost")
	if err != nil {
		panic(err)
	}
	return s
}

func getUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bid := bson.ObjectIdHex(id)
	u := User{}
	if err := session.DB("web").C("users").FindId(bid).One(&u); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	uj, err := json.Marshal(u)
	if err != nil {
		log.Fatalln(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", uj)
}

func postUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	u := User{}
	json.NewDecoder(r.Body).Decode(&u)

	u.ID = bson.NewObjectId()
	session.DB("web").C("users").Insert(u)

	uj, err := json.Marshal(u)
	if err != nil {
		log.Fatalln(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s\n", uj)
}

func deleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bid := bson.ObjectIdHex(id)
	if err := session.DB("web").C("users").RemoveId(bid); err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s was deleted", bid)
}

func updateUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")
	u := User{}
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	bid := bson.ObjectIdHex(id)
	json.NewDecoder(r.Body).Decode(&u)

	if err := session.DB("web").C("users").Update(bson.M{"_id": bid}, bson.M{"$set": bson.M{"name": u.Name, "surname": u.Surname}}); err != nil { //ugly as fuck
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s was modified", bid)
}
