package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents a client session
type Session struct {
	ID          string
	ClientName  string
	CreatedAt   time.Time
	Preferences SessionPreferences
}

// SessionPreferences holds session-specific preferences
type SessionPreferences struct {
	ReceiveMouseEvents  bool
	ReceiveKeyEvents    bool
	ReceiveResizeEvents bool
}

// SessionRegistry manages active sessions
type SessionRegistry struct {
	sessions   map[string]*Session
	maxSession int
	mu         sync.RWMutex
}

// NewSessionRegistry creates a new session registry
func NewSessionRegistry(maxSessions int) *SessionRegistry {
	return &SessionRegistry{
		sessions:   make(map[string]*Session),
		maxSession: maxSessions,
	}
}

// Create creates a new session
func (r *SessionRegistry) Create(clientName string, prefs SessionPreferences) (*Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.sessions) >= r.maxSession {
		return nil, fmt.Errorf("maximum sessions reached (%d)", r.maxSession)
	}

	session := &Session{
		ID:          uuid.New().String(),
		ClientName:  clientName,
		CreatedAt:   time.Now(),
		Preferences: prefs,
	}

	r.sessions[session.ID] = session
	return session, nil
}

// Get retrieves a session by ID
func (r *SessionRegistry) Get(id string) (*Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[id]
	return session, ok
}


// Exists checks if a session exists
func (r *SessionRegistry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.sessions[id]
	return ok
}


// Delete removes a session
func (r *SessionRegistry) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[id]; ok {
		delete(r.sessions, id)
		return true
	}
	return false
}

// Count returns the number of active sessions
func (r *SessionRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions)
}

// List returns all active session IDs
func (r *SessionRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		ids = append(ids, id)
	}
	return ids
}

// ListAll returns all active sessions
func (r *SessionRegistry) ListAll() []*Session {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sessions := make([]*Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

