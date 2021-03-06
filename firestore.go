package certcache

// Implementation of the autocert.Cache interface as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Firestore is a Google Firestore implementation of autocert.Cache
type Firestore struct {
	collectionName string // firestore has "collections" with "documents"
	ctxt           context.Context
	client         *firestore.Client
}

const (
	defaultFirestoreCertCacheCollectionName = "certcache"
)

// NewFirestore is the default constructor for a Firestore CertCache
func NewFirestore(credsPath, projectID string) *Firestore {
	return NewFirestoreWithCollection(credsPath, projectID, defaultFirestoreCertCacheCollectionName)
}

// NewFirestoreWithCollection is a constructor for a FirestoreCertCache
// with a custom Firestore Collection name
func NewFirestoreWithCollection(credsPath, projectID, certsCollectionName string) *Firestore {
	cntxt := context.Background()
	creds := option.WithCredentialsFile(credsPath)
	cl, err := firestore.NewClient(cntxt, projectID, creds)
	if err != nil {
		log.Fatalf("[FIRESTORE] failed to initialize firestore client: %s", err)
	}
	return &Firestore{
		ctxt:           cntxt,
		collectionName: certsCollectionName,
		client:         cl,
	}
}

type format struct {
	Data string `firestore:"data"`
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (fcc *Firestore) Get(ctx context.Context, key string) ([]byte, error) {
	log.Println(fmt.Sprintf("[firestore-certcache] fetching %s from firestore", key))
	docSnapshot, err := fcc.client.Collection(fcc.collectionName).Doc(key).Get(fcc.ctxt)
	if err != nil {
		log.Println(fmt.Sprintf("[firestore-certcache] error fetching %s from firestore: %s", key, err))
		if grpc.Code(err) == codes.NotFound {
			return nil, autocert.ErrCacheMiss
		}
		return nil, err
	}
	var doc format
	if err := docSnapshot.DataTo(&doc); err != nil {
		log.Println(fmt.Sprintf("[firestore-certcache] error reading document snapshot for %s: %s", key, err))
		return nil, err
	}
	log.Println(fmt.Sprintf("[firestore-certcache] fetched %s from firestore", key))
	return []byte(doc.Data), nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (fcc *Firestore) Put(ctx context.Context, key string, data []byte) error {
	log.Println(fmt.Sprintf("[firestore-certcache] storing %s in firestore", key))
	newDocRef := fcc.client.Collection(fcc.collectionName).Doc(key)
	if _, err := newDocRef.Set(fcc.ctxt, format{Data: string(data)}); err != nil {
		log.Println(fmt.Sprintf("[firestore-certcache] failed to store %s in firestore", key))
		return err
	}
	log.Println(fmt.Sprintf("[firestore-certcache] successfully stored %s in firestore", key))
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (fcc *Firestore) Delete(ctx context.Context, key string) error {
	_, err := fcc.client.Collection(fcc.collectionName).Doc(key).Delete(fcc.ctxt)
	return err
}
