package modrinth

import (
	"time"

	"orchestrator/internal/persistent"
)

const cacheTTL = 72 * time.Hour

const (
	bucketSearch   = "modrinth/search"
	bucketProject  = "modrinth/project"
	bucketVersion  = "modrinth/version"
	bucketVersions = "modrinth/versions"
)

// Cache provides a persistent bbolt-backed TTL cache for Modrinth DTOs.
type Cache struct {
	db *persistent.DB
}

// NewCache creates a Cache backed by the given persistent DB.
func NewCache(db *persistent.DB) *Cache {
	return &Cache{db: db}
}

// GetSearch returns a cached search response.
func (c *Cache) GetSearch(key string) (SearchResponse, bool) {
	var v SearchResponse
	ok, err := c.db.Get(bucketSearch, key, &v)
	if err != nil || !ok {
		return SearchResponse{}, false
	}
	return v, true
}

// SetSearch stores a search response in the cache.
func (c *Cache) SetSearch(key string, v SearchResponse) {
	_ = c.db.Set(bucketSearch, key, v, cacheTTL)
}

// GetProject returns a cached project.
func (c *Cache) GetProject(id string) (Project, bool) {
	var v Project
	ok, err := c.db.Get(bucketProject, id, &v)
	if err != nil || !ok {
		return Project{}, false
	}
	return v, true
}

// SetProject stores a project in the cache.
func (c *Cache) SetProject(id string, v Project) {
	_ = c.db.Set(bucketProject, id, v, cacheTTL)
}

// GetVersion returns a cached version.
func (c *Cache) GetVersion(id string) (Version, bool) {
	var v Version
	ok, err := c.db.Get(bucketVersion, id, &v)
	if err != nil || !ok {
		return Version{}, false
	}
	return v, true
}

// SetVersion stores a version in the cache.
func (c *Cache) SetVersion(id string, v Version) {
	_ = c.db.Set(bucketVersion, id, v, cacheTTL)
}

// GetVersions returns a cached version list.
func (c *Cache) GetVersions(key string) ([]Version, bool) {
	var v []Version
	ok, err := c.db.Get(bucketVersions, key, &v)
	if err != nil || !ok {
		return nil, false
	}
	return v, true
}

// SetVersions stores a version list in the cache.
func (c *Cache) SetVersions(key string, v []Version) {
	_ = c.db.Set(bucketVersions, key, v, cacheTTL)
}
