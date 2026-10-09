package main

type Room struct {
	name       string
	desribtion string
	ways       map[string]*Room
	items      map[string]bool
}

type Human struct {
	currentRoom *Room
	backpack    bool
	inventory   []string
}

func main() {

	initGame()

	_ = handleCommand("some command")
}

func initGame() {

}

func handleCommand(command string) string {

	return "not implemented"
}
