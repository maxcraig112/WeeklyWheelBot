package firestore

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/joho/godotenv"
)

// NewFirestoreClient initializes and returns a FirestoreClient using a specific database ID.
func NewFirestoreClient(ctx context.Context) (*FirestoreClient, error) {
	_ = godotenv.Load()

	projectID := os.Getenv("GCP_PROJECT_ID")
	databaseID := os.Getenv("DATABASE_ID")

	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if err != nil {
		return nil, err
	}

	return &FirestoreClient{
		client: client,
		dbID:   databaseID,
	}, nil
}

func (fc *FirestoreClient) GetCollection(collectionID string) *firestore.CollectionRef {
	return fc.client.Collection(collectionID)
}

// Close closes the Firestore client connection.
func (fc *FirestoreClient) Close() error {
	return fc.client.Close()
}
