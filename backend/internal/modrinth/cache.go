package modrinth

import "time"

const cacheTTL = 5 * time.Minute

type cacheEntry[T any] struct {
	value     T
	timestamp time.Time
}

func (e cacheEntry[T]) valid() bool {
	return time.Since(e.timestamp) < cacheTTL
}

// Cache provides a generic in-memory TTL cache.
type Cache struct {
	search  map[string]cacheEntry[SearchResponse]
	project map[string]cacheEntry[Project]
	version map[string]cacheEntry[Version]
	versions map[string]cacheEntry[[]Version]
}

// NewCache creates a new Cache.
func NewCache() *Cache {
	return &Cache{
		search:   make(map[string]cacheEntry[SearchResponse]),
		project:  make(map[string]cacheEntry[Project]),
		version:  make(map[string]cacheEntry[Version]),
		versions: make(map[string]cacheEntry[[]Version]),
	}
}

// GetSearch returns a cached search response.
func (c *Cache) GetSearch(key string) (SearchResponse, bool) {
	if e, ok := c.search[key]; ok && e.valid() {
		return e.value, true
	}
	return SearchResponse{}, false
}

// SetSearch stores a search response in the cache.
func (c *Cache) SetSearch(key string, v SearchResponse) {
	c.search[key] = cacheEntry[SearchResponse]{value: v, timestamp: time.Now()}
}

// GetProject returns a cached project.
func (c *Cache) GetProject(id string) (Project, bool) {
	if e, ok := c.project[id]; ok && e.valid() {
		return e.value, true
	}
	return Project{}, false
}

// SetProject stores a project in the cache.
func (c *Cache) SetProject(id string, v Project) {
	c.project[id] = cacheEntry[Project]{value: v, timestamp: time.Now()}
}

// GetVersion returns a cached version.
func (c *Cache) GetVersion(id string) (Version, bool) {
	if e, ok := c.version[id]; ok && e.valid() {
		return e.value, true
	}
	return Version{}, false
}

// SetVersion stores a version in the cache.
func (c *Cache) SetVersion(id string, v Version) {
	c.version[id] = cacheEntry[Version]{value: v, timestamp: time.Now()}
}

// GetVersions returns a cached version list.
func (c *Cache) GetVersions(key string) ([]Version, bool) {
	if e, ok := c.versions[key]; ok && e.valid() {
		return e.value, true
	}
	return nil, false
}

// SetVersions stores a version list in the cache.
func (c *Cache) SetVersions(key string, v []Version) {
	c.versions[key] = cacheEntry[[]Version]{value: v, timestamp: time.Now()}
}
