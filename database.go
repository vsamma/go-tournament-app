package main

import (  
	"gopkg.in/mgo.v2"
    "os"
)

func Database() *mgo.Session {  
    session, err := mgo.Dial("gotournament_database_1")

    if err != nil {
        panic(err)
    }

    session.SetMode(mgo.Monotonic, true)

    session.DB(os.Getenv("DB_NAME"))

    return session
}