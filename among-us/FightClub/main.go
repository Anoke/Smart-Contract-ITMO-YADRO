package FightClub

import (
	"github.com/nspcc-dev/neo-go/pkg/interop/native/oracle"
	"github.com/nspcc-dev/neo-go/pkg/interop/native/std"
	"github.com/nspcc-dev/neo-go/pkg/interop/runtime"
	"github.com/nspcc-dev/neo-go/pkg/interop/storage"
	"math"
)

type Player struct {
	name     string
	speed    int
	stamina  int
	strength int
	balance  int
	items    []int
	points   int
}

const storeURLKey = "storeURL"
const playersListKey = "playersList"

func _deploy(data interface{}, isUpdate bool) {
	if isUpdate {
		return
	}
	ctx := storage.GetContext()
	storeURL := "https://git.frostfs.info/Web3N3/among-us/data.json"
	storage.Put(ctx, storeURLKey, storeURL)
	runtime.Log("Smart contract deployed successfully.")
}

func NewPlayer(playerName string) {
	ctx := storage.GetContext()
	existingPlayer := storage.Get(ctx, playerName)
	if existingPlayer != nil {
		panic("player already exists")
	}

	items := make([]int, 0)

	player := Player{
		balance:  20,
		items:    items,
		name:     playerName,
		points:   5,
		speed:    0,
		stamina:  0,
		strength: 0,
	}
	storage.Put(ctx, playerName, std.Serialize(player))
	UpdatePlayersList(playerName)
	runtime.Log("New player created: " + playerName)
}

func UpdatePlayersList(playerName string) {
	ctx := storage.GetContext()
	playersListBytes := storage.Get(ctx, playersListKey)
	var playersList []string
	if playersListBytes != nil {
		playersList = std.Deserialize(playersListBytes.([]byte)).([]string)
	}
	playersList = append(playersList, playerName)

	storage.Put(ctx, playersListKey, std.Serialize(playersList))
}

func getPlayer(playerName string) Player {
	ctx := storage.GetReadOnlyContext()
	data := storage.Get(ctx, playerName)
	if data == nil {
		panic("player not found")
	}
	return std.Deserialize(data.([]byte)).(Player)
}

func Balance(playerName string) int {
	p := getPlayer(playerName)
	return p.balance
}

func Items(playerName string) []int {
	p := getPlayer(playerName)
	return p.items
}

func BuyItem(playerName string, itemID int) {
	player := getPlayer(playerName)
	for i := range player.items {
		if itemID == player.items[i] {
			panic("item has already been purchased")
		}
	}

	if contains(player.items, itemID) {
		panic("item has already been purchased")
	}

	ctx := storage.GetContext()
	storeURL := storage.Get(ctx, storeURLKey).(string)
	filter := []byte("$.store.item[" + std.Itoa10(itemID) + "]")
	oracle.Request(storeURL, filter, "cbBuyItem", playerName, 2*oracle.MinimumResponseGas)
}

func contains(arr []int, item int) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}
	return false
}

func CbBuyItem(url string, userData any, code int, result []byte) {
	callingHash := runtime.GetCallingScriptHash()
	if !callingHash.Equals(oracle.Hash) {
		panic("not called from the oracle contract")
	}
	if code != oracle.Success {
		panic("request failed for " + url + " with code " + std.Itoa(code, 10))
	}
	runtime.Log("Result for " + url + " is: " + string(result))

	resultLen := len(result)
	data := std.JSONDeserialize(result[1 : resultLen-1]).(map[string]any)
	price := data["price"].(int)
	gearID := data["id"].(int)
	playerName := userData.(string)

	player := getPlayer(playerName)
	if player.balance < price {
		panic("insufficient balance")
	}
	player.balance -= price

	player.items = append(player.items, gearID)

	ctx := storage.GetContext()
	storage.Put(ctx, playerName, std.Serialize(player))
	runtime.Log("Item purchased successfully by " + playerName)
}

func DistributePoints(playerName string, pointsSpeed, pointsStamina, pointsStrength int) {
	player := getPlayer(playerName)
	if pointsSpeed+pointsStamina+pointsStrength > player.points {
		panic("not enough points")
	}
	player.speed += pointsSpeed
	player.stamina += pointsStamina
	player.strength += pointsStrength
	player.points -= pointsSpeed + pointsStamina + pointsStrength
	ctx := storage.GetContext()
	storage.Put(ctx, playerName, std.Serialize(player))
	runtime.Log("Points distributed to " + playerName)
}

func getAllPlayers() []string {
	ctx := storage.GetReadOnlyContext()
	playersListBytes := storage.Get(ctx, playersListKey)
	if playersListBytes == nil {
		return []string{}
	}

	playersList := std.Deserialize(playersListBytes.([]byte)).([]string)
	return playersList
}

func getRealOpponent(playerName string) string {
	ctx := storage.GetReadOnlyContext()
	playersListBytes := storage.Get(ctx, playersListKey)
	if playersListBytes == nil {
		panic("no players available")
	}

	playersList := std.Deserialize(playersListBytes.([]byte)).([]string)
	if len(playersList) < 2 {
		panic("not enough players for an opponent")
	}

	var opponents []string

	for _, name := range playersList {
		if name != playerName {
			opponents = append(opponents, name)
		}
	}

	if len(opponents) == 0 {
		panic("no available opponents")
	}
	randomNumber := random(len(opponents))
	return opponents[randomNumber]
}

func random(a int) int {
	return runtime.GetRandom() % a
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func calculateWinningChances(playerData Player, opponentData Player) int {
	chance := 50

	if playerData.speed != opponentData.speed {
		chance += (playerData.speed - opponentData.speed) * 5
		chance = int(math.Max(5.0, math.Min(float64(chance), 95.0)))
	}

	if playerData.stamina != opponentData.stamina {
		chance += (playerData.stamina - opponentData.stamina) * 5
		chance = int(math.Max(5.0, math.Min(float64(chance), 95.0)))
	}
	if playerData.strength != opponentData.strength {
		chance += (playerData.strength - opponentData.strength) * 5
		chance = int(math.Max(5.0, math.Min(float64(chance), 95.0)))
	}

	return chance
}

func battle(playerName string, opponentName string, bet int) string {
	if getPlayer(opponentName).name == "" {
		panic("opponent not found")
	}

	player := getPlayer(playerName)
	opponent := getPlayer(opponentName)

	if player.balance < bet || opponent.balance < bet {
		panic("insufficient balance for the battle")
	}

	chancePlayer := calculateWinningChances(player, opponent)
	winner := chooseWinner(chancePlayer)

	runtime.Log("Battle started: " + playerName + " vs " + opponentName)
	runtime.Log("Chances of " + playerName + ": " + std.Itoa(chancePlayer, 10) + "%")

	if winner == "draw" {
		runtime.Log("The battle ended in a draw!")
	} else {
		runtime.Log("Winner: " + winner)
	}

	distributeWinnings(winner, playerName, opponentName, bet)
	return winner
}

func chooseWinner(chancePlayer1 int) string {
	randomNumber := random(100)
	if randomNumber <= chancePlayer1 {
		return "player1"
	} else {
		return "player2"
	}
}

func distributeWinnings(winner, player1, player2 string, bet int) {
	winnerData := getPlayer(winner)
	player1Data := getPlayer(player1)
	player2Data := getPlayer(player2)

	winnerData.balance += bet

	if winner == player1 {
		player2Data.balance -= bet
	} else {
		player1Data.balance -= bet
	}

	ctx := storage.GetContext()
	storage.Put(ctx, winner, std.Serialize(winnerData))
	storage.Delete(ctx, player1)
	storage.Delete(ctx, player2)

	UpdatePlayersList(winner)
	runtime.Log("Winnings distributed successfully.")
}

func findMatch(playerName string, bet int) string {
	opponent := getRealOpponent(playerName)
	if opponent == "" {
		runtime.Log("No real opponents found for " + playerName + ". Searching for a random opponent.")
		return findRandomMatch(playerName, bet)
	}
	return battle(playerName, opponent, bet)
}

func findRandomMatch(playerName string, bet int) string {
	opponent := createRandomOpponent(playerName)
	if opponent == "" {
		runtime.Log("No random opponents found for " + playerName + ". Waiting for more players.")
		return "draw"
	}
	return battle(playerName, opponent, bet)
}

func createRandomOpponent(playerName string) string {
	opponents := getAllPlayers()
	if len(opponents) < 2 {
		return ""
	}

	var randomOpponentName string
	for {
		randomOpponentName = opponents[random(len(opponents))]
		if randomOpponentName != playerName {
			break
		}
	}

	return randomOpponentName
}
