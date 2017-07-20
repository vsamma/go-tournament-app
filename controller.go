package main

import (  
	"gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
	"os"
	"net/http"
    "encoding/json"
    "fmt"
    "strconv"
    "sync"
    "errors"
)

type (  
    TournamentController struct {
        session *mgo.Session
    }
)

func NewTournamentController(s *mgo.Session) *TournamentController {  
    return &TournamentController{s}
}

func (c *TournamentController) reset(w http.ResponseWriter, r *http.Request){
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

    respondWithJSON(w, http.StatusCreated, players)
}

func (c *TournamentController) GetAllTournaments(w http.ResponseWriter, r *http.Request) {  	
    tournaments := []Tournament{}
    session := c.session.DB(os.Getenv("DB_NAME")).C("tournaments")
    err := session.Find(nil).Limit(100).All(&tournaments)
    if err != nil {
        panic(err)
    }

    respondWithJSON(w, http.StatusCreated, tournaments)
}

func (c *TournamentController) addFundsToPlayer(w http.ResponseWriter, r *http.Request){
    //Get and parse URL parameters
    params := r.URL.Query()

    playerId := params.Get("playerId")
    points, err := strconv.ParseFloat(params.Get("points"), 64)

    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")

    player, err := getPlayerById(session, playerId)
	if err != nil {
        err = session.Insert(&Player{PlayerId: playerId, Balance: points})

        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
	} else {
        err = addFunds(session, player, points)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
    }
    respondWithJSON(w, http.StatusOK, nil)
}

func (c *TournamentController) takeFundsFromPlayer(w http.ResponseWriter, r *http.Request){
    //Get and parse URL parameters
    params := r.URL.Query()

    playerId := params.Get("playerId")

    points, err := strconv.ParseFloat(params.Get("points"), 64)

    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")
    
    player, err := getPlayerById(session, playerId)
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
    //Get and parse URL parameters
    params := r.URL.Query()

    tournamentId, err := strconv.Atoi(params.Get("tournamentId"))
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    deposit, err := strconv.ParseFloat(params.Get("deposit"), 64)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("tournaments")

    //Check if tournament exists already
    _, err = getTournamentById(session, tournamentId)
	if err != nil {
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
    params := r.URL.Query()
    fmt.Println(params)

    //Params existance check. TournamentId and PlayerId are mandatory for joining a tournament.
    if len(params) == 0 {
        respondWithError(w, http.StatusInternalServerError, "No GET Parameters were found. 'tournamentId' and 'playerId' are mandatory!")
        return
	}

    tournamentId, err := strconv.Atoi(params.Get("tournamentId"))
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    }

    playerId := params.Get("playerId")
    if playerId == "" {
        respondWithError(w, http.StatusInternalServerError, "GET Parameter 'playerId' missing!")
        return
    }

    //Tournament existance check.
    tournamentSession := c.session.DB(os.Getenv("DB_NAME")).C("tournaments")

    tournament, err := getTournamentById(tournamentSession, tournamentId)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    } else if tournament.Status != "OPEN" {
        respondWithError(w, http.StatusConflict, fmt.Sprintf("Tournament with ID '%d' has already finished.", tournamentId))
        return
    }

    playerSession := c.session.DB(os.Getenv("DB_NAME")).C("players")
    player, err := getPlayerById(playerSession, playerId)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Player '%s' was not found!", playerId))
        return
    }

    //TODO - how to handle backers? 
    //Should the first check be if player has enough points and if not, then deal with backers?
    //Or if backers exist, always divide the deposit sum between player + backers even if player has enough points on his own?
    //DECISION - according to the task, it seems that player's balance must be checked first.
    //SO if backers are added as params, they are ignored if player himself has enough funds.
    if player.Balance < tournament.Deposit {
        //Player's balance is lower than needed for the tournament. Check for backers.
        backers := params["backerId"]
        if len(backers) > 0 {
            partDeposit := tournament.Deposit / float64(len(backers) + 1)

            backersList := []Player{} 
            err = playerSession.Find(bson.M{"playerid": bson.M{"$in": backers}}).All(&backersList)
            if err != nil {
                respondWithError(w, http.StatusInternalServerError, "Backers query failed!: " + err.Error())
                return
            }

            //Add player to backers list for correct deposit calculation
            backersList = append(backersList, player)

            var wg sync.WaitGroup
            wg.Add(len(backersList))
            for _, p := range backersList {
                go func(p Player) {
                    defer wg.Done()

                    if p.Balance >= partDeposit {
                        err = takeFunds(playerSession, p, partDeposit)
                        if err != nil {
                            respondWithError(w, http.StatusInternalServerError, err.Error())
                            return
                        }
                    } else {
                        respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Player '%s' does not have enough funds!",p.PlayerId))
                        return
                    }
                }(p)
            }
            
            var participant Participant
            participant.PlayerId = player.PlayerId
            participant.Deposit = partDeposit
            participant.Backers = backers
            err = addParticipant(tournamentSession, tournament, participant)
            if err != nil {
                respondWithError(w, http.StatusInternalServerError, err.Error())
                return
            }

            wg.Wait()
        } else {
            respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Player '%s' does not have enough funds to join Tournament '%d' and no backers as well.", playerId, tournamentId))
            return
        }
    } else {
        //Player has enough points so I will deliberately ignore backers because task description had no information about this
        err = takeFunds(playerSession, player, tournament.Deposit)

        var participant Participant
        participant.PlayerId = player.PlayerId
        participant.Deposit = tournament.Deposit
        err = addParticipant(tournamentSession, tournament, participant)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
    }

    respondWithJSON(w, http.StatusOK, nil)
}

func (c *TournamentController) resultTournament(w http.ResponseWriter, r *http.Request){
    var result TournamentResult
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&result); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }
    defer r.Body.Close()

    //Get tournament info
    tournamentSession := c.session.DB(os.Getenv("DB_NAME")).C("tournaments")
    playerSession := c.session.DB(os.Getenv("DB_NAME")).C("players")

    tournament, err := getTournamentById(tournamentSession, result.TournamentId)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, err.Error())
        return
    } else if tournament.Status != "OPEN" {
        respondWithError(w, http.StatusConflict, fmt.Sprintf("Tournament with ID '%d' has already finished.", result.TournamentId))
        return
    }

    //Transfer winnings to all winners
    for _, winner := range result.Winners {
        err = payWinnings(tournamentSession, playerSession, tournament, winner)
        if err != nil {
            respondWithError(w, http.StatusInternalServerError, err.Error())
            return
        }
    }

    //Close tournament
    changeTournamentStatus(tournamentSession, tournament, "CLOSED")

    respondWithJSON(w, http.StatusOK, nil)
}

func payWinnings(tournamentSession *mgo.Collection, playerSession *mgo.Collection, tournament Tournament, winner Winner) error {
    var winningParticipant Participant
    for _, participant := range tournament.Participants {
        if winner.PlayerId == participant.PlayerId{
            winningParticipant = participant
            break
        }
    }

    //Get player
    player, err := getPlayerById(playerSession, winningParticipant.PlayerId)
    if err != nil {
        return err
    }

    //Pay out winnings
    if len(winningParticipant.Backers) == 0 {
        //No backers, so all winnings go to player. 
        if tournament.Deposit != winningParticipant.Deposit {
            return errors.New("player's deposit doesn't match with tournament's required deposit")
        } 

        addFunds(playerSession, player, winner.Prize)
    } else {
        //Winner participant has backers, so divide the winnings.
        partWinnings := winner.Prize / float64(len(winningParticipant.Backers)+1)

        //Get backers
        backersList := []Player{} 
        err = playerSession.Find(bson.M{"playerid": bson.M{"$in": winningParticipant.Backers}}).All(&backersList)
        if err != nil {
            return err
        }

        //Add player to the same list as his backers to pay winnings to everybody
        backersList = append(backersList, player)

        //Pay out all the winnings to player and his backers
        for _, p := range backersList {
            err = addFunds(playerSession, p, partWinnings)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (c *TournamentController) getBalance(w http.ResponseWriter, r *http.Request){
    //Get and parse URL parameters
    params := r.URL.Query()
    playerId := params.Get("playerId")

    //Get players collection and add player if doesn't exist and then add funds
    session := c.session.DB(os.Getenv("DB_NAME")).C("players")
    
    player, err := getPlayerById(session, playerId)
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
    player.Balance += points
    colQuerier := bson.M{"playerid": player.PlayerId}
    change := bson.M{"$set": bson.M{"balance": player.Balance}}
    return session.Update(colQuerier, change)
}

func getPlayerById(session *mgo.Collection, playerId string) (Player, error) {
    player := Player{}
    err := session.Find(bson.M{"playerid": playerId}).One(&player)
    return player, err
}

func getTournamentById(session *mgo.Collection, tournamentId int) (Tournament, error) {
    tournament := Tournament{}
    err := session.Find(bson.M{"id": tournamentId}).One(&tournament)
    return tournament, err
}

func addParticipant(session *mgo.Collection, tournament Tournament, participant Participant) error {
    tournament.Participants = append(tournament.Participants, participant)
    colQuerier := bson.M{"id": tournament.Id}
    change := bson.M{"$set": bson.M{"participants": tournament.Participants}}
    return session.Update(colQuerier, change)
}

func changeTournamentStatus(session *mgo.Collection, tournament Tournament, newStatus string) error {
    colQuerier := bson.M{"id": tournament.Id}
    change := bson.M{"$set": bson.M{"status": newStatus}}
    return session.Update(colQuerier, change)
}
