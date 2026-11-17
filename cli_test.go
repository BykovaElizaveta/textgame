package main

import (
	"bytes"
	"strings"
	"testing"
)

func playScript(t *testing.T, script string) string {
	t.Helper()
	var out bytes.Buffer
	runGame(strings.NewReader(script), &out)
	return out.String()
}

func TestNormalizeCommand(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"осмотреться", "осмотреться"},
		{"  ИДТИ   Коридор  ", "идти коридор"},
		{"   ", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := normalizeCommand(tc.in); got != tc.want {
			t.Errorf("normalizeCommand(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCLIWelcomeAndExit(t *testing.T) {
	for name, script := range map[string]string{
		"command": "выход\n",
		"eof":     "",
	} {
		t.Run(name, func(t *testing.T) {
			out := playScript(t, script)
			for _, want := range []string{
				"TEXT QUEST",
				"ты находишься на кухне",
				"[кухня]> ",
				"До встречи!",
			} {
				if !strings.Contains(out, want) {
					t.Errorf("вывод не содержит %q:\n%s", want, out)
				}
			}
			if strings.Contains(out, "\x1b[") {
				t.Errorf("в не-терминал не должны печататься ANSI-коды:\n%s", out)
			}
		})
	}
}

func TestCLIUppercaseAndSpaces(t *testing.T) {
	out := playScript(t, "  ИДТИ   КОРИДОР  \nвыход\n")
	want := "ничего интересного. можно пройти - кухня, комната, улица"
	if !strings.Contains(out, want) {
		t.Errorf("команда с регистром и лишними пробелами не сработала:\n%s", out)
	}
}

func TestCLIHelp(t *testing.T) {
	out := playScript(t, "помощь\nвыход\n")
	for _, want := range []string{"Команды:", "применить <предмет> <цель>", "инвентарь", "заново"} {
		if !strings.Contains(out, want) {
			t.Errorf("справка не содержит %q:\n%s", want, out)
		}
	}
}

func TestCLIInventoryAndMap(t *testing.T) {
	script := "инвентарь\nидти коридор\nкарта\nидти комната\nнадеть рюкзак\nвзять ключи\nинвентарь\nвыход\n"
	out := playScript(t, script)
	for _, want := range []string{
		"рюкзак ещё не надет",
		"в рюкзаке пока пусто",
		"Вы здесь: коридор",
		"🔒 дверь закрыта",
		"рюкзак надет",
		"в рюкзаке: ключи",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("вывод не содержит %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "TEXT QUEST") != 1 {
		t.Errorf("случайный перезапуск игры:\n%s", out)
	}
}

func TestCLIHintAfterFailures(t *testing.T) {
	out := playScript(t, "бла\nбла\nбла\nвыход\n")
	if !strings.Contains(out, "Подсказка:") {
		t.Errorf("после трёх промахов должна появиться подсказка:\n%s", out)
	}
}

func TestCLIFullPlaythroughAndVictory(t *testing.T) {
	script := strings.Join([]string{
		"идти коридор",
		"идти комната",
		"надеть рюкзак",
		"взять ключи",
		"взять конспекты",
		"идти коридор",
		"применить ключи дверь",
		"идти улица",
		"н",
	}, "\n") + "\n"

	out := playScript(t, script)
	for _, want := range []string{
		"дверь открыта",
		"на улице весна. можно пройти - домой",
		"ТЫ УСПЕЛ В УНИВЕР",
		"Ходов:",
		"Сыграть ещё? (д/н):",
		"До встречи!",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("прохождение не содержит %q:\n%s", want, out)
		}
	}
}

func TestCLIRestart(t *testing.T) {
	out := playScript(t, "заново\nосмотреться\nвыход\n")
	if strings.Count(out, "TEXT QUEST") != 2 {
		t.Errorf("после «заново» игра должна начаться сначала:\n%s", out)
	}
	if strings.Count(out, "ты находишься на кухне") < 2 {
		t.Errorf("после перезапуска должно быть снова описание кухни:\n%s", out)
	}
}
