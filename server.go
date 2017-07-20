package main

import (  
	"net/http"
	// "encoding/json"
	"github.com/gorilla/mux"
	"fmt"
	"log"
)

func main() {
    handleRequests()
}

func handleRequests() {
    c := NewTournamentController(Database())

	myRouter := mux.NewRouter().StrictSlash(true)
	// myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/fund", c.addFundsToPlayer).Methods("GET").Queries("playerId", "{playerId}", "points", "{points}")
	myRouter.HandleFunc("/take", c.takeFundsFromPlayer).Methods("GET").Queries("playerId", "{playerId}", "points", "{points}")
	myRouter.HandleFunc("/announceTournament", c.announceTournament).Methods("GET").Queries("tournamentId", "{tournamentId}", "deposit", "{deposit}")
	myRouter.HandleFunc("/joinTournament", c.joinTournament).Methods("GET")
	myRouter.HandleFunc("/resultTournament", c.resultTournament).Methods("POST")
	myRouter.HandleFunc("/balance", c.getBalance).Methods("GET").Queries("playerId", "{playerId}")
	myRouter.HandleFunc("/reset", c.reset).Methods("GET")
	myRouter.HandleFunc("/players", c.GetAllPlayers).Methods("GET")
	myRouter.HandleFunc("/tournaments", c.GetAllTournaments).Methods("GET")
	http.Handle("/", myRouter)

	fmt.Println("Starting up on port 3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

// func homePage(w http.ResponseWriter, r *http.Request){
//     fmt.Fprintf(w, "Welcome to the HomePage!")
// 	fmt.Println("Endpoint Hit: homePage")
// }

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    // response, _ := json.Marshal(payload)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    // w.Write(response)
}