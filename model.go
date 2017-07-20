package main

type Player struct {  
    PlayerId        string          `json:"playerId"`
    Balance         float64         `json:"balance"`
}

type Tournament struct {  
    Id              int             `json:"id"`
    Deposit         float64         `json:"deposit"`
    Status          string          `json:"status"`
    Participants    []Participant   `json:"participants"`
}

type Participant struct {
    PlayerId        string          `json:"playerId"`
    Deposit         float64         `json:"deposit"`
    Backers         []string        `json:"backers"`
}

type TournamentResult struct {
    TournamentId    int             `json:"tournamentId"`
    Winners         []Winner        `json:"winners"`
}

type Winner struct {
    PlayerId        string          `json:"playerId"`
    Prize           float64         `json:"prize"`
}