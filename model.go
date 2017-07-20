package main

import (  
    // "gopkg.in/mgo.v2/bson"
)

type Product struct {  
    Id          string      `json:"id"`
    Title       string      `json:"title"`
    Description string      `json:"description"`
    Price       float64     `json:"price"`
}

type Player struct {  
    PlayerId    string      `json:"playerId"`
    Balance     float64     `json:"balance"`
}

type Tournament struct {  
    Id          int         `json:"id"`
    Deposit     float64     `json:"deposit"`
    Status      string      `json:"status"`
}