// Copyright 2012 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	// Current version.
	Version = "1.0.beta1"

	// defaultUrl to be used as base for Elasticsearch requests.
	defaultUrl = "http://localhost:9200"

	// pingDuration is the time to periodically check the Elasticsearch URLs.
	pingDuration = 60 * time.Second
)

var (
	// ErrNoClient is raised when no active Elasticsearch client is available.
	ErrNoClient = errors.New("no active client")
)

// Client is an ElasticSearch client. Create one by calling NewClient.
type Client struct {
	urls []string // urls is a list of all clients for ElasticSearch queries

	c *http.Client // c is the net/http Client to use for requests

	mu        sync.Mutex // mutex for the next two fields
	activeUrl string     // currently active connection url
	hasActive bool       // true if we have an active connection
}

// Creates a new client to work with ElasticSearch.
func NewClient(client *http.Client, urls ...string) (*Client, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	c := &Client{c: client}
	switch len(urls) {
	case 0:
		c.urls = make([]string, 1)
		c.urls[0] = defaultUrl
	case 1:
		c.urls = make([]string, 1)
		c.urls[0] = urls[0]
	default:
		c.urls = urls
	}
	c.pingUrls()
	go c.pinger() // start goroutine periodically ping all clients
	return c, nil
}

// NewRequest creates a new request with the given method and prepends
// the base URL to the path. If no active connection to Elasticsearch
// is available, ErrNoClient is returned.
func (c *Client) NewRequest(method, path string) (*Request, error) {
	if !c.hasActive {
		return nil, ErrNoClient
	}
	return NewRequest(method, c.activeUrl+path)
}

// pinger periodically runs pingUrls.
func (c *Client) pinger() {
	ticker := time.NewTicker(pingDuration)
	for {
		select {
		case <-ticker.C:
			c.pingUrls()
		}
	}
}

// pingUrls iterates through all client URLs. It checks if the client
// is available. It takes the first one available and saves its URL
// in activeUrl. If no client is available, hasActive is set to false
// and NewRequest will fail.
func (c *Client) pingUrls() {
	for _, url_ := range c.urls {
		params := make(url.Values)
		params.Set("timeout", "1")
		req, err := NewRequest("HEAD", url_+"/?"+params.Encode())
		if err == nil {
			res, err := c.c.Do((*http.Request)(req))
			if err == nil {
				defer res.Body.Close()
				if res.StatusCode == http.StatusOK {
					// Everything okay: Update activeUrl and set hasActive to true.
					c.mu.Lock()
					defer c.mu.Unlock()
					if c.activeUrl != "" && c.activeUrl != url_ {
						log.Printf("elastic: switched connection from %s to %s", c.activeUrl, url_)
					}
					c.activeUrl = url_
					c.hasActive = true
					return
				}
			} else {
				log.Printf("elastic: %v", err)
			}
		} else {
			log.Printf("elastic: %v", err)
		}
	}

	// No client available
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hasActive = false
}

// CreateIndex starts the service to create a new index.
func (c *Client) CreateIndex(name string) *CreateIndexService {
	builder := NewCreateIndexService(c)
	builder.Index(name)
	return builder
}

// DeleteIndex starts the service to delete an index.
func (c *Client) DeleteIndex(name string) *DeleteIndexService {
	builder := NewDeleteIndexService(c)
	builder.Index(name)
	return builder
}

// IndexExists allows to check if an index exists.
func (c *Client) IndexExists(name string) *IndexExistsService {
	builder := NewIndexExistsService(c)
	builder.Index(name)
	return builder
}

// Index a document.
func (c *Client) Index() *IndexService {
	builder := NewIndexService(c)
	return builder
}

// Delete a document.
func (c *Client) Delete() *DeleteService {
	builder := NewDeleteService(c)
	return builder
}

// DeleteByQuery deletes documents as found by a query.
func (c *Client) DeleteByQuery() *DeleteByQueryService {
	builder := NewDeleteByQueryService(c)
	return builder
}

// Get a document.
func (c *Client) Get() *GetService {
	builder := NewGetService(c)
	return builder
}

// Exists checks if a document exists.
func (c *Client) Exists() *ExistsService {
	builder := NewExistsService(c)
	return builder
}

// Count documents.
func (c *Client) Count(indices ...string) *CountService {
	builder := NewCountService(c)
	builder.Indices(indices...)
	return builder
}

// Search is the entry point for searches.
func (c *Client) Search(indices ...string) *SearchService {
	builder := NewSearchService(c)
	builder.Indices(indices...)
	return builder
}

// Suggest returns an interface to return suggestions.
func (c *Client) Suggest(indices ...string) *SuggestService {
	builder := NewSuggestService(c)
	builder.Indices(indices...)
	return builder
}

// Scan through documents.
func (c *Client) Scan(indices ...string) *ScanService {
	builder := NewScanService(c)
	builder.Indices(indices...)
	return builder
}

// Flush asks Elasticsearch to free memory from the index and
// flush data to disk.
func (c *Client) Flush() *FlushService {
	builder := NewFlushService(c)
	return builder
}

// Bulk is the entry point to mass insert/update/delete documents.
func (c *Client) Bulk() *BulkService {
	builder := NewBulkService(c)
	return builder
}

// Alias enables the caller to add and/or remove aliases.
func (c *Client) Alias() *AliasService {
	builder := NewAliasService(c)
	return builder
}

// Aliases returns aliases by index name(s).
func (c *Client) Aliases() *AliasesService {
	builder := NewAliasesService(c)
	return builder
}
