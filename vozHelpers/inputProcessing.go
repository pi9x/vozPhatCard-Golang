package vozHelpers

import s "strings"

const Delim = "|"

func CreateCardList(msg string) []Card {
	var cards []Card
	tmp := s.Split(msg, "\n")
	for _, st := range tmp {
		if s.Contains(st, Delim) {
			var info = s.SplitN(st, Delim, 2)
			if len(info) == 2 {
				cards = append(cards, Card{info[0], info[0] + ": " + info[1]})
			}
		}
	}
	return cards
}
