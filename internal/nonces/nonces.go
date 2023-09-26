// wraith - Copyright (c) 2023 Michael D Henderson. All rights reserved.

package nonces

import (
	"github.com/google/uuid"
	"log"

	"sync"
	"time"
)

func NewFactory(ttl time.Duration) *Factory {
	return &Factory{
		data: make(map[string]time.Time),
		ttl:  ttl,
	}
}

// Factory implements a factory with caching, ttl, and cache-cleaning.
type Factory struct {
	sync.Mutex
	data          map[string]time.Time
	ttl           time.Duration // time-to-live for each nonce
	checkExpiryAt time.Time     // next time to check for expired data
}

// Create creates a new nonce.
// Should maybe panic since errors aren't recoverable.
func (f *Factory) Create() (string, error) {
	f.Lock()
	defer f.Unlock()

	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	nonce := id.String()

	f.data[nonce] = time.Now().Add(f.ttl)

	log.Printf("[nonce] created %q %v\n", nonce, f.data[nonce])
	return nonce, nil
}

// Lookup returns true if it finds the nonce in the cache, and it hasn't expired.
// Side effect: clears the cache every so often.
func (f *Factory) Lookup(nonce string) bool {
	f.Lock()
	defer f.Unlock()

	log.Printf("[nonce] lookup %q %v\n", nonce, time.Now())
	for nonce, exp := range f.data {
		log.Printf("[nonce] cached %q %v\n", nonce, exp)
	}

	now := time.Now()
	exp, ok := f.data[nonce]
	if ok {
		ok = now.Before(exp)
		delete(f.data, nonce)
	}

	// do we need to check for expired data?
	if now.After(f.checkExpiryAt) {
		for nonce, exp := range f.data {
			if now.After(exp) {
				delete(f.data, nonce)
				continue
			}
		}
		// check again in 15 minutes
		f.checkExpiryAt = now.Add(15 * time.Minute)
	}

	return ok
}
