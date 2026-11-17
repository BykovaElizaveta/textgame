package main

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type sessionStats struct {
	moves   int
	start   time.Time
	unknown int
}

func (s *sessionStats) reset() {
	s.moves = 0
	s.start = time.Now()
	s.unknown = 0
}

type cliSession struct {
	scanner *bufio.Scanner
	r       *renderer
	stats   *sessionStats
}

// runGame — игровой цикл. Принимает reader/writer, чтобы сценарии можно было
// прогонять в тестах без реальной консоли.
func runGame(in io.Reader, out io.Writer) {
	s := &cliSession{
		scanner: bufio.NewScanner(in),
		r:       newRenderer(out),
		stats:   &sessionStats{},
	}
	s.scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	s.startGame()
	s.loop()
}

func normalizeCommand(line string) string {
	return strings.Join(strings.Fields(strings.ToLower(line)), " ")
}

func (s *cliSession) startGame() {
	initGame()
	s.stats.reset()

	s.r.line("")
	s.r.line(s.r.paint(colorBold+";"+colorCyan, "  ╔══════════════════════════════════╗"))
	s.r.line(s.r.paint(colorBold+";"+colorCyan, "  ║          TEXT QUEST              ║"))
	s.r.line(s.r.paint(colorBold+";"+colorCyan, "  ║     Соберись и иди в универ      ║"))
	s.r.line(s.r.paint(colorBold+";"+colorCyan, "  ╚══════════════════════════════════╝"))
	s.r.line(s.r.paint(colorGray, "  Введи «помощь», чтобы увидеть команды."))
	s.r.line("")
	s.r.typeLine(s.r.paint(colorCyan, handleCommand("осмотреться")))
}

func (s *cliSession) loop() {
	for {
		s.prompt()
		if !s.scanner.Scan() {
			s.sayGoodbye()
			return
		}

		command := normalizeCommand(s.scanner.Text())
		if command == "" {
			continue
		}
		if command == "выход" || command == "quit" || command == "exit" {
			s.sayGoodbye()
			return
		}
		if s.metaCommand(command) {
			continue
		}

		previousRoom := currentPlayer.currentRoom
		answer := handleCommand(command)
		s.stats.moves++
		s.r.typeLine(s.colorAnswer(answer))

		if isFailure(answer) {
			s.stats.unknown++
		} else {
			s.stats.unknown = 0
		}
		s.maybeHint()

		if strings.HasPrefix(command, "идти ") &&
			currentPlayer.currentRoom != previousRoom &&
			currentPlayer.currentRoom.name == "улица" {
			s.victory()
			if !s.askReplay() {
				return
			}
			s.startGame()
		}
	}
}

func (s *cliSession) prompt() {
	s.r.write(s.r.paint(colorCyan, "["+currentPlayer.currentRoom.name+"]") + "> ")
	s.r.flush()
}

func (s *cliSession) sayGoodbye() {
	s.r.line("")
	s.r.line(s.r.paint(colorGray, "До встречи!"))
}

func (s *cliSession) metaCommand(command string) bool {
	switch command {
	case "помощь", "?", "help":
		s.printHelp()
		return true
	case "инвентарь", "inventory":
		s.printInventory()
		return true
	case "карта", "map":
		s.printMap()
		return true
	case "заново", "restart":
		s.r.line(s.r.paint(colorGray, "Новая игра!"))
		s.startGame()
		return true
	}
	return false
}

func (s *cliSession) printHelp() {
	s.r.line(s.r.paint(colorCyan, "Команды:"))
	rows := [][2]string{
		{"осмотреться", "посмотреть вокруг"},
		{"идти <место>", "перейти в комнату"},
		{"взять <предмет>", "положить в рюкзак"},
		{"надеть <предмет>", "надеть рюкзак"},
		{"применить <предмет> <цель>", "например: применить ключи дверь"},
		{"инвентарь", "что в рюкзаке"},
		{"карта", "куда можно пройти"},
		{"помощь", "показать этот список"},
		{"заново", "начать игру заново"},
		{"выход", "выйти из игры"},
	}
	for _, row := range rows {
		s.r.line("  " + s.r.paint(colorYellow, row[0]) + " — " + row[1])
	}
	s.r.line(s.r.paint(colorGray, "Пример: надеть рюкзак → взять ключи → идти коридор → применить ключи дверь → идти улица"))
}

func (s *cliSession) printInventory() {
	if currentPlayer.backpack {
		s.r.line(s.r.paint(colorGreen, "рюкзак надет"))
	} else {
		s.r.line(s.r.paint(colorGray, "рюкзак ещё не надет"))
	}
	if len(currentPlayer.inventory) == 0 {
		s.r.line("в рюкзаке пока пусто")
		return
	}
	s.r.line("в рюкзаке: " + s.r.paint(colorYellow, strings.Join(currentPlayer.inventory, ", ")))
}

func (s *cliSession) printMap() {
	room := currentPlayer.currentRoom
	s.r.line(s.r.paint(colorCyan, "Вы здесь: "+room.name))

	keys := make([]string, 0, len(room.ways))
	for key := range room.ways {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		s.r.line("никуда не пройти")
		return
	}
	for _, key := range keys {
		line := "  → " + key
		if lockMsg, locked := room.locks[key]; locked {
			line += " " + s.r.paint(colorRed, "🔒 "+lockMsg)
		}
		s.r.line(line)
	}
}

func isFailure(answer string) bool {
	return answer == "неизвестная команда" ||
		strings.HasPrefix(answer, "не") ||
		strings.HasPrefix(answer, "нет ") ||
		strings.Contains(answer, "закрыта")
}

func (s *cliSession) colorAnswer(answer string) string {
	switch {
	case isFailure(answer):
		return s.r.paint(colorRed, answer)
	case strings.Contains(answer, "добавлен"),
		strings.Contains(answer, "надели"),
		strings.Contains(answer, "открыта"):
		return s.r.paint(colorGreen, answer)
	default:
		return s.r.paint(colorCyan, answer)
	}
}

func hasItem(item string) bool {
	for _, invItem := range currentPlayer.inventory {
		if invItem == item {
			return true
		}
	}
	return false
}

func (s *cliSession) maybeHint() {
	if s.stats.unknown == 0 || s.stats.unknown%3 != 0 {
		return
	}

	var hint string
	switch {
	case !currentPlayer.backpack:
		hint = "Подсказка: сначала надень рюкзак — «надеть рюкзак»."
	case !hasItem("ключи"):
		hint = "Подсказка: ключи в комнате — зайди туда и возьми их."
	default:
		if _, locked := currentPlayer.currentRoom.locks["улица"]; locked {
			hint = "Подсказка: попробуй «применить ключи дверь»."
		} else {
			hint = "Подсказка: «осмотреться» покажет, что вокруг."
		}
	}
	s.r.line(s.r.paint(colorGray, hint))
}

func (s *cliSession) victory() {
	s.r.line("")
	s.r.line(s.r.paint(colorBold+";"+colorGreen, "  ╔══════════════════════════════════╗"))
	s.r.line(s.r.paint(colorBold+";"+colorGreen, "  ║      ТЫ УСПЕЛ В УНИВЕР! 🎓       ║"))
	s.r.line(s.r.paint(colorBold+";"+colorGreen, "  ╚══════════════════════════════════╝"))
	s.r.line(fmt.Sprintf("  Ходов: %d. Время: %s.", s.stats.moves, time.Since(s.stats.start).Round(time.Second)))
	s.r.line("")
}

func (s *cliSession) askReplay() bool {
	s.r.write("Сыграть ещё? (д/н): ")
	s.r.flush()
	if !s.scanner.Scan() {
		s.sayGoodbye()
		return false
	}

	switch normalizeCommand(s.scanner.Text()) {
	case "д", "да", "y", "yes":
		return true
	default:
		s.sayGoodbye()
		return false
	}
}
