# go-tournament-app
A small REST API application in GoLang with MongoDB and Gorilla MUX setup with Docker.

**SETUP**

You would need to have GoLang and Docker Toolbox installed to run this application.

In this case, clone the repository to your GOPATH/src folder and just run the following commands inside the repo folder:

* `docker-compose build`
* `docker-compose up`

As the server is configured to run on port `3000`, just use `localhost` or Docker machine's specific IP which it tells you (depending on your operating system) and start the web application:

* `http://localhost:3000`

or for example:
* `http://192.168.99.100:3000`

---

**AVAILABLE ENDPOINTS**
* `GET - /fund?playerId={playerId}&points={points}` - Adds funds to a player's balance. Creates the player if doesn't exist
* `GET - /take?playerId={playerId}&points={points}` - Takes funds from a player's balance.
* `GET - /announceTournament?tournamentId={tournamentId}&deposit={deposit}` - Initializes a tournament with its required deposit amount
* `GET - /joinTournament?tournamentId={tournamentId}&playerId={playerId}&backerId={backerId}` - Adds a player (and its possible backers) to the tournament. `tournamentId` and `playerId` are mandatory but `backerId`-s are optional and those can be added more than one
* `POST - /resultTournament` - Finishes the tournament and declares the winner(s). Payload must be in this JSON format: `{"tournamentId": 1, "winners": [{"playerId": "P1", "prize": 2000}]}
`
* `GET - /balance?playerId={playerId}` - Shows player's current balance
* `GET - /reset` - Resets the whole database.
* `GET - /players` - Returns all players.
* `Get - /tournaments` - Returns all tournaments.
