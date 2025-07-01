package guild

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type GuildData struct {
	GuildID        string                   `firestore:"guildID"`
	Guesses        map[string]int           `firestore:"guesses"`
	SpunNumbers    []SpunNumber             `firestore:"rolledNumbers"`
	LastNumberSpun SpunNumber               `firestore:"latsNumberSpun"`
	CollectionRef  *firestore.CollectionRef `firestore:"-"`
}

type SpunNumber struct {
	Number   int
	DateSpun time.Time
}

func (g *GuildData) SetGuess(ctx context.Context, userID string, guess int) error {
	if g.Guesses == nil {
		g.Guesses = make(map[string]int)
	}

	g.Guesses[userID] = guess

	// Update only the guesses field in Firestore
	_, err := g.CollectionRef.Doc(g.GuildID).Update(ctx, []firestore.Update{
		{
			Path:  "guesses",
			Value: g.Guesses,
		},
	})
	return err
}

func (g *GuildData) AddSpunNumber(ctx context.Context, number int) error {
	rolled := SpunNumber{
		Number:   number,
		DateSpun: time.Now(),
	}
	g.SpunNumbers = append(g.SpunNumbers, rolled)
	g.LastNumberSpun = rolled

	_, err := g.CollectionRef.Doc(g.GuildID).Update(ctx, []firestore.Update{
		{
			Path:  "rolledNumbers",
			Value: g.SpunNumbers,
		},
		{
			Path:  "latsNumberSpun",
			Value: g.LastNumberSpun,
		},
	})
	return err
}

func (g *GuildData) BulkAddSpunNumbers(ctx context.Context, numbers []int) error {
	uniqueMap := make(map[int]struct{})
	uniqueNumbers := make([]int, 0, len(numbers))
	for _, n := range numbers {
		if _, exists := uniqueMap[n]; !exists && n >= 1 && n <= 1000 {
			uniqueMap[n] = struct{}{}
			uniqueNumbers = append(uniqueNumbers, n)
		}
	}

	rolledNumbers := make([]SpunNumber, len(uniqueNumbers))
	now := time.Now()
	for i, n := range uniqueNumbers {
		rolledNumbers[i] = SpunNumber{
			Number:   n,
			DateSpun: now,
		}
	}
	g.SpunNumbers = rolledNumbers
	if len(rolledNumbers) > 0 {
		g.LastNumberSpun = rolledNumbers[len(rolledNumbers)-1]
	} else {
		g.LastNumberSpun = SpunNumber{}
	}

	_, err := g.CollectionRef.Doc(g.GuildID).Update(ctx, []firestore.Update{
		{
			Path:  "rolledNumbers",
			Value: g.SpunNumbers,
		},
		{
			Path:  "latsNumberSpun",
			Value: g.LastNumberSpun,
		},
	})
	return err
}
