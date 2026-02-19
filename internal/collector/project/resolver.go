package project

import (
	"context"
	"log/slog"
	"sync"
	"time"

	keystonedb "github.com/vexxhost/openstack_database_exporter/internal/db/keystone"
)

const defaultTTL = 5 * time.Minute

// Info holds resolved project name and domain ID.
type Info struct {
	Name     string
	DomainID string
}

// Resolver resolves project IDs to names and domain IDs via keystone DB.
// It caches the mapping and refreshes it periodically based on a TTL.
type Resolver struct {
	logger     *slog.Logger
	keystoneDB *keystonedb.Queries
	ttl        time.Duration

	mu       sync.RWMutex
	projects map[string]Info
	lastLoad time.Time
}

// NewResolver creates a resolver that fetches projects from keystone and
// caches them for the given TTL. If keystoneDB is nil, the resolver returns
// project IDs as-is. A zero TTL uses the default (5 minutes).
func NewResolver(logger *slog.Logger, keystoneDB *keystonedb.Queries, ttl time.Duration) *Resolver {
	if ttl == 0 {
		ttl = defaultTTL
	}

	r := &Resolver{
		logger:     logger,
		keystoneDB: keystoneDB,
		ttl:        ttl,
		projects:   make(map[string]Info),
	}

	if keystoneDB == nil {
		logger.Warn("Keystone database not available, tenant labels will use project IDs")
		return r
	}

	r.refresh()
	return r
}

// refresh reloads the project mapping from keystone.
func (r *Resolver) refresh() {
	if r.keystoneDB == nil {
		return
	}

	projects, err := r.keystoneDB.GetProjectMetrics(context.Background())
	if err != nil {
		r.logger.Error("Failed to load projects from keystone", "error", err)
		return
	}

	newMap := make(map[string]Info, len(projects))
	for _, p := range projects {
		newMap[p.ID] = Info{
			Name:     p.Name,
			DomainID: p.DomainID,
		}
	}

	r.mu.Lock()
	r.projects = newMap
	r.lastLoad = time.Now()
	r.mu.Unlock()

	r.logger.Info("Loaded project mappings from keystone", "count", len(newMap))
}

// ensureFresh triggers a refresh if the cached data is older than the TTL.
func (r *Resolver) ensureFresh() {
	r.mu.RLock()
	stale := time.Since(r.lastLoad) > r.ttl
	r.mu.RUnlock()

	if stale {
		r.refresh()
	}
}

// Resolve returns the project name and domain_id for a given project ID.
// Falls back to the project ID itself if not found.
func (r *Resolver) Resolve(projectID string) (name, domainID string) {
	r.ensureFresh()

	r.mu.RLock()
	defer r.mu.RUnlock()

	if info, ok := r.projects[projectID]; ok {
		return info.Name, info.DomainID
	}
	return projectID, ""
}

// AllProjects returns a snapshot of all cached project IDs and their info.
func (r *Resolver) AllProjects() map[string]Info {
	r.ensureFresh()

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid data races
	result := make(map[string]Info, len(r.projects))
	for k, v := range r.projects {
		result[k] = v
	}
	return result
}
