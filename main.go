package main

import "strings"

type Room struct {
	name        string
	description string
	entryText   string
	look        func() string
	ways        map[string]*Room
	items       map[string]bool
	triggers    map[string]func() string
	locks       map[string]string
}

type Human struct {
	currentRoom *Room
	backpack    bool
	inventory   []string
}

var currentPlayer *Human

func (man *Human) lookAround() string {
	return man.currentRoom.look()
}

func (man *Human) goTo(toRoom string) string {
	isGo, ok := man.currentRoom.ways[toRoom]
	if ok == false {
		return "нет пути в " + toRoom
	}
	lockMsg, isLocked := man.currentRoom.locks[toRoom]
	if isLocked {
		return lockMsg
	}
	man.currentRoom = isGo
	return isGo.entryText
}

func (man *Human) putOn(item string) string {
	isAvailable, ok := man.currentRoom.items[item]
	if ok == false || isAvailable == false || item != "рюкзак" {
		return "нет такого"
	}
	if man.backpack == true {
		return "вы уже надели: " + item
	}
	man.currentRoom.items[item] = false
	man.backpack = true
	return "вы надели: " + item
}

func (man *Human) take(item string) string {
	isAvailable, ok := man.currentRoom.items[item]
	if ok == false || isAvailable == false {
		return "нет такого"
	}
	if man.backpack == false {
		return "некуда класть"
	}
	man.currentRoom.items[item] = false
	man.inventory = append(man.inventory, item)
	return "предмет добавлен в инвентарь: " + item
}

func (man *Human) useTo(item, target string) string {
	hasItem := false
	for _, invItem := range man.inventory {
		if invItem == item {
			hasItem = true
			break
		}
	}
	if !hasItem {
		return "нет предмета в инвентаре - " + item
	}
	triggerKey := item + " " + target
	action, exists := man.currentRoom.triggers[triggerKey]
	if !exists {
		return "не к чему применить"
	}
	return action()
}

var playerCommand map[string]func([]string) string = map[string]func([]string) string{
	"осмотреться": func(args []string) string {
		return currentPlayer.lookAround()
	},
	"идти": func(args []string) string {
		if len(args) < 2 {
			return "куда идти?"
		}
		return currentPlayer.goTo(args[1])
	},
	"надеть": func(args []string) string {
		if len(args) < 2 {
			return "что надеть?"
		}
		return currentPlayer.putOn(args[1])
	},
	"взять": func(args []string) string {
		if len(args) < 2 {
			return "что взять?"
		}
		return currentPlayer.take(args[1])
	},
	"применить": func(args []string) string {
		if len(args) < 3 {
			return "что применить?"
		}
		return currentPlayer.useTo(args[1], args[2])
	},
}

func main() {

	initGame()

	_ = handleCommand("some command")
}

func initGame() {
	kitchen := &Room{
		name:      "кухня",
		entryText: "кухня, ничего интересного. можно пройти - коридор",
		ways:      make(map[string]*Room),
		items:     make(map[string]bool),
		triggers:  make(map[string]func() string),
		locks:     make(map[string]string),
	}

	corridor := &Room{
		name:      "коридор",
		entryText: "ничего интересного. можно пройти - кухня, комната, улица",
		ways:      make(map[string]*Room),
		items:     make(map[string]bool),
		triggers:  make(map[string]func() string),
		locks:     make(map[string]string),
	}

	myRoom := &Room{
		name:      "комната",
		entryText: "ты в своей комнате. можно пройти - коридор",
		ways:      make(map[string]*Room),
		items:     make(map[string]bool),
		triggers:  make(map[string]func() string),
		locks:     make(map[string]string),
	}

	street := &Room{
		name:      "улица",
		entryText: "на улице весна. можно пройти - домой",
		ways:      make(map[string]*Room),
		items:     make(map[string]bool),
		triggers:  make(map[string]func() string),
		locks:     make(map[string]string),
	}

	kitchen.ways["коридор"] = corridor

	corridor.ways["кухня"] = kitchen
	corridor.ways["комната"] = myRoom
	corridor.ways["улица"] = street

	myRoom.ways["коридор"] = corridor

	street.ways["домой"] = corridor

	kitchen.items["чай"] = true

	myRoom.items["ключи"] = true
	myRoom.items["конспекты"] = true
	myRoom.items["рюкзак"] = true

	currentPlayer = &Human{
		currentRoom: kitchen,
		backpack:    false,
		inventory:   make([]string, 0), // Инициализируем пустой срез
	}

	kitchen.look = func() string {
		text := "ты находишься на кухне, на столе: чай, надо "
		if currentPlayer.backpack {
			text += "идти в универ"
		} else {
			text += "собрать рюкзак и идти в универ"
		}
		return text + ". можно пройти - коридор"
	}

	corridor.look = func() string {
		return "ничего интересного. можно пройти - кухня, комната, улица"
	}

	myRoom.look = func() string {
		var tableItems []string

		if myRoom.items["ключи"] {
			tableItems = append(tableItems, "ключи")
		}
		if myRoom.items["конспекты"] {
			tableItems = append(tableItems, "конспекты")
		}

		var parts []string
		if len(tableItems) > 0 {
			parts = append(parts, "на столе: "+strings.Join(tableItems, ", "))
		}
		if myRoom.items["рюкзак"] {
			parts = append(parts, "на стуле: рюкзак")
		}

		var result string
		if len(parts) > 0 {
			result = strings.Join(parts, ", ")
		} else {
			result = "пустая комната"
		}

		return result + ". можно пройти - коридор"
	}

	street.look = func() string {
		return "на улице весна. можно пройти - домой"
	}

	corridor.locks["улица"] = "дверь закрыта"

	corridor.triggers["ключи дверь"] = func() string {
		delete(corridor.locks, "улица")
		return "дверь открыта"
	}
}

func handleCommand(command string) string {
	args := strings.Split(command, " ")
	action, ok := playerCommand[args[0]]
	if ok == false {
		return "неизвестная команда"
	}
	return action(args)
}
