package publisher

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ask-23/go-wx/pkg/config"
	"github.com/ask-23/go-wx/pkg/database"
)

// Publisher defines the interface for a service that publishes weather data
type Publisher interface {
	Start() error
	Stop() error
	Name() string
}

// publishers is a registry of available publisher implementations
var publishers = map[string]func(config.PublisherConfig, *database.Database) (Publisher, error){
	"wunderground": NewWundergroundPublisher,
	"custom":       NewCustomPublisher,
}

// InitializePublishers creates publishers based on configuration
func InitializePublishers(configs []config.PublisherConfig, db *database.Database) ([]Publisher, error) {
	var pubs []Publisher

	for _, cfg := range configs {
		// Skip disabled publishers
		if !cfg.Enabled {
			log.Printf("Publisher %s is disabled, skipping", cfg.Name)
			continue
		}

		// Look up publisher constructor by name
		constructor, ok := publishers[cfg.Name]
		if !ok {
			log.Printf("Unknown publisher type: %s, skipping", cfg.Name)
			continue
		}

		// Create the publisher
		pub, err := constructor(cfg, db)
		if err != nil {
			return nil, fmt.Errorf("failed to create publisher %s: %w", cfg.Name, err)
		}

		pubs = append(pubs, pub)
		log.Printf("Publisher %s initialized", cfg.Name)
	}

	return pubs, nil
}

// BasePublisher provides common functionality for publishers
type BasePublisher struct {
	config       config.PublisherConfig
	db           *database.Database
	ticker       *time.Ticker
	done         chan struct{}
	wg           sync.WaitGroup
	running      bool
	publisherMux sync.Mutex
}

// Start begins the publishing process
func (b *BasePublisher) Start() error {
	b.publisherMux.Lock()
	defer b.publisherMux.Unlock()

	if b.running {
		return fmt.Errorf("publisher is already running")
	}

	// Set up ticker for periodic publishing
	interval := time.Duration(b.config.Interval) * time.Second
	b.ticker = time.NewTicker(interval)
	b.done = make(chan struct{})
	b.running = true

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.ticker.C:
				if err := b.publish(); err != nil {
					log.Printf("Error publishing to %s: %v", b.config.Name, err)
				}
			case <-b.done:
				return
			}
		}
	}()

	log.Printf("Started publisher: %s with interval %d seconds", b.config.Name, b.config.Interval)
	return nil
}

// Stop stops the publishing process
func (b *BasePublisher) Stop() error {
	b.publisherMux.Lock()
	defer b.publisherMux.Unlock()

	if !b.running {
		return nil
	}

	b.ticker.Stop()
	close(b.done)
	b.wg.Wait()
	b.running = false

	log.Printf("Stopped publisher: %s", b.config.Name)
	return nil
}

// Name returns the name of this publisher
func (b *BasePublisher) Name() string {
	return b.config.Name
}

// publish is a placeholder for the actual publishing mechanism
// It should be overridden by specific publisher implementations
func (b *BasePublisher) publish() error {
	return fmt.Errorf("publish not implemented")
}
