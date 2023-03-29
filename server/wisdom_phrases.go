package server

import (
	"math/rand"
	"time"
)

type Quote struct {
	Name, Phrase string
}

func GetRandomQuote() *Quote {
	rand.Seed(time.Now().UnixNano())
	return quotes[rand.Intn(len(quotes))]
}

var quotes = []*Quote{
	{
		Name:   "William Shakespeare",
		Phrase: "The fool doth think he is wise, but the wise man knows himself to be a fool.",
	},
	{
		Name:   "Maurice Switzer",
		Phrase: "It is better to remain silent at the risk of being thought a fool, than to talk and remove all doubt of it.",
	},
	{
		Name:   "Mark Twain",
		Phrase: "Whenever you find yourself on the side of the majority, it is time to reform (or pause and reflect).",
	},
	{
		Name:   "Jess C. Scott",
		Phrase: "When someone loves you, the way they talk about you is different. You feel safe and comfortable.",
	},
	{
		Name:   "Aristotle",
		Phrase: "Knowing yourself is the beginning of all wisdom.",
	},
	{
		Name:   "Socrates",
		Phrase: "The only true wisdom is in knowing you know nothing.",
	},
	{
		Name:   "Isaac Asimov",
		Phrase: "The saddest aspect of life right now is that science gathers knowledge faster than society gathers wisdom.",
	},
	{
		Name:   "John Lennon",
		Phrase: "Count your age by friends, not years. Count your life by smiles, not tears.",
	},
	{
		Name:   "Mark Twain",
		Phrase: "In a good bookroom you feel in some mysterious way that you are absorbing the wisdom contained in all the books through your skin, without even opening them.",
	},
	{
		Name:   "Jonathan Swift",
		Phrase: "May you live every day of your life.",
	},
}
