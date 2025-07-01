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
