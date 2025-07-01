package gcp

import (
	"context"
	"main/gcp/firestore"
)

// Clients holds all external service clients.
type Clients struct {
	Firestore firestore.FirestoreClientInterface
}

// InitialiseClients creates and returns all required service clients.
func InitialiseClients(ctx context.Context) (*Clients, error) {
	firestoreClient, err := firestore.NewFirestoreClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Firestore: firestoreClient,
	}, nil
}

func (c *Clients) CloseClients() error {
	if c.Firestore != nil {
		err := c.Firestore.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
