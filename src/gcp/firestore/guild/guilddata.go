package guild

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type GuildData struct {
	GuildID          string                   `firestore:"guildID"`
	Guesses          map[string]int           `firestore:"guesses"`
	RolledNumbers    []RolledNumber           `firestore:"rolledNumbers"`
	LastNumberRolled RolledNumber             `firestore:"latsNumberRolled"`
	CollectionRef    *firestore.CollectionRef `firestore:"-"`
}

type RolledNumber struct {
	Number     int
	DateRolled time.Time
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

func (g *GuildData) AddRolledNumber(ctx context.Context, number int) error {
	rolled := RolledNumber{
		Number:     number,
		DateRolled: time.Now(),
	}
	g.RolledNumbers = append(g.RolledNumbers, rolled)
	g.LastNumberRolled = rolled

	_, err := g.CollectionRef.Doc(g.GuildID).Update(ctx, []firestore.Update{
		{
			Path:  "rolledNumbers",
			Value: g.RolledNumbers,
		},
		{
			Path:  "latsNumberRolled",
			Value: g.LastNumberRolled,
		},
	})
	return err
}

func (g *GuildData) BulkAddRolledNumbers(ctx context.Context, numbers []int) error {
	uniqueMap := make(map[int]struct{})
	uniqueNumbers := make([]int, 0, len(numbers))
	for _, n := range numbers {
		if _, exists := uniqueMap[n]; !exists && n >= 1 && n <= 1000 {
			uniqueMap[n] = struct{}{}
			uniqueNumbers = append(uniqueNumbers, n)
		}
	}

	rolledNumbers := make([]RolledNumber, len(uniqueNumbers))
	now := time.Now()
	for i, n := range uniqueNumbers {
		rolledNumbers[i] = RolledNumber{
			Number:     n,
			DateRolled: now,
		}
	}
	g.RolledNumbers = rolledNumbers
	if len(rolledNumbers) > 0 {
		g.LastNumberRolled = rolledNumbers[len(rolledNumbers)-1]
	} else {
		g.LastNumberRolled = RolledNumber{}
	}

	_, err := g.CollectionRef.Doc(g.GuildID).Update(ctx, []firestore.Update{
		{
			Path:  "rolledNumbers",
			Value: g.RolledNumbers,
		},
		{
			Path:  "latsNumberRolled",
			Value: g.LastNumberRolled,
		},
	})
	return err
}
