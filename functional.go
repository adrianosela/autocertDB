package certcache

// Implementation of the autocert.Cache interface as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"
)

// Functional allows the user to use functions to define a cert cache.
// If we have the get function always return an autocert.ErrCacheMiss error,
// we can use this cert cache for testing next cache layer's preconditions,
// or simply logging events
type Functional struct {
	get func(context.Context, string) ([]byte, error)
	put func(context.Context, string) error
	del func(context.Context, string) error
}

// NewFunctional is the constructor for a functional Cert Cache
func NewFunctional(
	get func(context.Context, string) ([]byte, error),
	put func(context.Context, string) error,
	del func(context.Context, string) error,
) *Functional {
	return &Functional{
		get: get,
		put: put,
		del: del,
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (f *Functional) Get(ctx context.Context, key string) ([]byte, error) {
	return f.get(ctx, key)
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (f *Functional) Put(ctx context.Context, key string, data []byte) error {
	return f.put(ctx, key)
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (f *Functional) Delete(ctx context.Context, key string) error {
	return f.del(ctx, key)
}
