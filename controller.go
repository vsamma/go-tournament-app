package main

import (  
    //"github.com/EwanValentine/project/api/models"
    // "github.com/martini-contrib/render"
    //"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
	"os"
	"net/http"
    "encoding/json"
    "fmt"
    // "net/url"
    //"github.com/gorilla/mux"
    "strconv"
)

type (  
    TournamentController struct {
        session *mgo.Session
    }
)

func NewTournamentController(s *mgo.Session) *TournamentController {  
    return &TournamentController{s}
}

// func (c *TournamentController) GetAllProducts(r render.Render) {  
func (c *TournamentController) GetAllProducts(w http.ResponseWriter, r *http.Request) {  	
    products := []Product{}
    session := c.session.DB(os.Getenv("DB_NAME")).C("products")
    err := session.Find(nil).Limit(100).All(&products)

    if err != nil {
        panic(err)
    }

	// r.JSON(200, products)
    // json.NewEncoder(w).Encode(products)
    respondWithJSON(w, http.StatusCreated, products)
}

// func (c *TournamentController) PostProduct(product Product, r render.Render) {  
func (c *TournamentController) PostProduct(w http.ResponseWriter, r *http.Request) {  
    session := c.session.DB(os.Getenv("DB_NAME")).C("products")
    fmt.Println("Endpoint Hit: POST /products")
    // product.Id = bson.NewObjectId()
    // product.Title = product.Title
    // product.Description = product.Description
    // product.Price = product.Price
    // session.Insert(product)

    var p Product
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&p); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    defer r.Body.Close()

    // if err := p.createProduct(a.DB); err != nil {
    if err := session.Insert(p); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    fmt.Println("New product added")

    // r.JSON(201, product)
    // json.NewEncoder(w).Encode(products)
    respondWithJSON(w, http.StatusCreated, p)
}

func (c *TournamentController) reset(w http.ResponseWriter, r *http.Request){
    // fmt.Fprintf(w, "Welcome to the HomePage!")
    fmt.Println("Endpoint Hit: reset")
    if err := c.session.DB(os.Getenv("DB_NAME")).DropDatabase(); err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
	
}

func (c *TournamentController) GetAllPlayers(w http.ResponseWriter, r *http.Request) {  	
    players := []Player{}
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")
    err := session.Find(nil).Limit(100).All(&players)

    if err != nil {
        panic(err)
    }

	// r.JSON(200, products)
    // json.NewEncoder(w).Encode(products)
    respondWithJSON(w, http.StatusCreated, players)
}

func (c *TournamentController) GetAllTournaments(w http.ResponseWriter, r *http.Request) {  	
    tournaments := []Tournament{}
    session := c.session.DB(os.Getenv("DB_NAME")).C("tournaments")
    err := session.Find(nil).Limit(100).All(&tournaments)

    if err != nil {
        panic(err)
    }

	// r.JSON(200, products)
    // json.NewEncoder(w).Encode(products)
    respondWithJSON(w, http.StatusCreated, tournaments)
}

func (c *TournamentController) addFundsToPlayer(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: addFundsToPlayer")
    //Get and parse URL parameters
    params := r.URL.Query()

    playerId := params.Get("playerId")
    fmt.Println("PlayerId: " + playerId)

    points, err := strconv.ParseFloat(params.Get("points"), 64)

    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")

    player := Player{}
	err = session.Find(bson.M{"playerid": playerId}).One(&player)
	if err != nil {
        fmt.Println("Player was not found and is being added: " + playerId)
        err = session.Insert(&Player{PlayerId: playerId, Balance: points})

        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
        fmt.Println("Player and points are added successfully: " + playerId)
	} else {
        err = addFunds(session, player, points)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
    }


    fmt.Println(fmt.Sprintf("Added '%f' points to player '%s'", points, playerId))

    respondWithJSON(w, http.StatusOK, nil)
}

func (c *TournamentController) takeFundsFromPlayer(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: takeFundsFromPlayer")
    //Get and parse URL parameters
    params := r.URL.Query()

    playerId := params.Get("playerId")
    fmt.Println("PlayerId: " + playerId)

    points, err := strconv.ParseFloat(params.Get("points"), 64)

    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")
    
    player := Player{}
	err = session.Find(bson.M{"playerid": playerId}).One(&player)
	if err != nil {
        respondWithError(w, http.StatusNotFound, err.Error())
        return
    }

    err = takeFunds(session, player, points)
	if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, nil)
}

func (c *TournamentController) announceTournament(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: announceTournament")
    //Get and parse URL parameters
    params := r.URL.Query()

    tournamentId, err := strconv.Atoi(params.Get("tournamentId"))
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }
    fmt.Println(fmt.Sprintf("tournamentId: %d",tournamentId))

    deposit, err := strconv.ParseFloat(params.Get("deposit"), 64)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("tournaments")

    tournament := Tournament{}
	err = session.Find(bson.M{"id": tournamentId}).One(&tournament)
	if err != nil {
        fmt.Println(fmt.Sprintf("Adding a new tournament: %d with deposit amount %f", tournamentId, deposit))
        err = session.Insert(&Tournament{Id: tournamentId, Deposit: deposit, Status: "OPEN"})
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
	} else {
        respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Tournament with id '%d' already exists!", tournamentId))
        return
    }

    respondWithJSON(w, http.StatusOK, nil)
}

func  (c *TournamentController) joinTournament(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: joinTournament")
    params := r.URL.Query()
    fmt.Println(params)

    if len(params) == 0 {
        respondWithError(w, http.StatusInternalServerError, "No GET Parameters were found. 'tournamentId' and 'playerId' are mandatory!")
        return
	}

    tournamentId := params.Get("tournamentId")
    if tournamentId == "" {
        respondWithError(w, http.StatusInternalServerError, "GET Parameter 'tournamentId' missing!")
        return
    }

    playerId := params.Get("playerId")
    if playerId == "" {
        respondWithError(w, http.StatusInternalServerError, "GET Parameter 'playerId' missing!")
        return
    }

    backers := params["backerId"]
    if len(backers) > 0 {
        for _, element := range backers {
            fmt.Println("Backer: " + element)
        }
	}

    respondWithJSON(w, http.StatusOK, nil)
}

func (c *TournamentController) resultTournament(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: resultTournament")

    respondWithJSON(w, http.StatusOK, nil)
}

func (c *TournamentController) getBalance(w http.ResponseWriter, r *http.Request){
    fmt.Println("Endpoint Hit: getBalance")
    //Get and parse URL parameters
    params := r.URL.Query()

    playerId := params.Get("playerId")
    fmt.Println("PlayerId: " + playerId)

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")
    
    player := Player{}
    err := session.Find(bson.M{"playerid": playerId}).One(&player)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    respondWithJSON(w, http.StatusOK, player)
}

func takeFunds(session *mgo.Collection, player Player, points float64) error {
    return addFunds(session, player, -points)
}

func addFunds(session *mgo.Collection, player Player, points float64) error {
    fmt.Println("Player was found and points are being added: " + player.PlayerId)
    fmt.Println(player)

    player.Balance += points

    fmt.Println(fmt.Sprintf("Player new balance now: %f",player.Balance))

    colQuerier := bson.M{"playerid": player.PlayerId}
    change := bson.M{"$set": bson.M{"balance": player.Balance}}
    return session.Update(colQuerier, change)
}