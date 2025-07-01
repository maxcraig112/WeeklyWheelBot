package firestore

import (
	"cloud.google.com/go/firestore"
)

// FirestoreClientInterface defines methods for Firestore operations.
type FirestoreClientInterface interface {
	GetCollection(collectionID string) *firestore.CollectionRef
	// CreateGuildDocument(ctx context.Context, collection *firestore.CollectionRef, guildID string) error
	// CreateOrGetGuildDocument(ctx context.Context, collection *firestore.CollectionRef, guildID string) (error, *firestore.DocumentRef)
	Close() error
}

// FirestoreClient wraps the Firestore client and implements FirestoreClientInterface.
type FirestoreClient struct {
	client *firestore.Client
	dbID   string
}
