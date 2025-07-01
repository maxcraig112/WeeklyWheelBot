package guild

import (
	"context"

	fs "main/gcp/firestore"

	"cloud.google.com/go/firestore"
)

const (
	GUILD_COLLECTION_ID = "guilds"
)

type GuildStore struct {
	firestoreClient fs.FirestoreClientInterface
	guilds          *firestore.CollectionRef
}

func NewGuildStore(client fs.FirestoreClientInterface) *GuildStore {
	return &GuildStore{
		firestoreClient: client,
		guilds:          client.GetCollection(GUILD_COLLECTION_ID),
	}
}

func (fc *GuildStore) CreateGuildDocument(ctx context.Context, guildID string) (*GuildData, error) {
	guildData := &GuildData{
		GuildID:          guildID,
		Guesses:          map[string]int{},
		RolledNumbers:    []RolledNumber{},
		LastNumberRolled: RolledNumber{},
	}
	_, err := fc.guilds.Doc(guildID).Set(ctx, guildData)
	if err != nil {
		return nil, err
	}
	guildData.CollectionRef = fc.guilds
	return guildData, nil
}

func (fc *GuildStore) CreateOrGetGuildDocument(ctx context.Context, guildID string) (*GuildData, error) {
	docRef := fc.guilds.Doc(guildID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		// If not found, create a new document
		return fc.CreateGuildDocument(ctx, guildID)
	}

	var guildData GuildData
	if err := docSnap.DataTo(&guildData); err != nil {
		return nil, err
	}
	guildData.CollectionRef = fc.guilds
	return &guildData, nil
}
