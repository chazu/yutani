package server

import (
	"sync"

	"github.com/google/uuid"
	"github.com/rivo/tview"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// WidgetInfo stores metadata about a widget
type WidgetInfo struct {
	ID         string
	SessionID  string
	Type       pb.WidgetType
	Primitive  tview.Primitive
	Properties *pb.WidgetProperties
	ParentID   string
	ChildIDs   []string
}

// WidgetRegistry manages widgets and their ownership
type WidgetRegistry struct {
	widgets      map[string]*WidgetInfo // widgetId -> info
	sessionRoots map[string]string      // sessionId -> root widgetId
	mu           sync.RWMutex
}

// NewWidgetRegistry creates a new widget registry
func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{
		widgets:      make(map[string]*WidgetInfo),
		sessionRoots: make(map[string]string),
	}
}

// Register registers a widget for a session and returns its ID
func (r *WidgetRegistry) Register(sessionID string, widgetType pb.WidgetType, widget tview.Primitive, properties *pb.WidgetProperties) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()
	r.widgets[id] = &WidgetInfo{
		ID:         id,
		SessionID:  sessionID,
		Type:       widgetType,
		Primitive:  widget,
		Properties: properties,
		ChildIDs:   make([]string, 0),
	}
	return id
}

// Get retrieves a widget primitive by ID
func (r *WidgetRegistry) Get(id string) (tview.Primitive, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.widgets[id]
	if !ok {
		return nil, false
	}
	return info.Primitive, true
}

// GetInfo retrieves full widget info by ID
func (r *WidgetRegistry) GetInfo(id string) (*WidgetInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.widgets[id]
	return info, ok
}

// GetOwner returns the session ID that owns a widget
func (r *WidgetRegistry) GetOwner(widgetID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.widgets[widgetID]
	if !ok {
		return "", false
	}
	return info.SessionID, true
}

// UpdateProperties updates widget properties
func (r *WidgetRegistry) UpdateProperties(widgetID string, properties *pb.WidgetProperties) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, ok := r.widgets[widgetID]
	if !ok {
		return false
	}
	info.Properties = properties
	return true
}

// AddChild adds a child widget to a parent
func (r *WidgetRegistry) AddChild(parentID, childID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	parent, ok := r.widgets[parentID]
	if !ok {
		return false
	}
	child, ok := r.widgets[childID]
	if !ok {
		return false
	}

	// Remove from old parent if exists
	if child.ParentID != "" {
		oldParent, ok := r.widgets[child.ParentID]
		if ok {
			for i, id := range oldParent.ChildIDs {
				if id == childID {
					oldParent.ChildIDs = append(oldParent.ChildIDs[:i], oldParent.ChildIDs[i+1:]...)
					break
				}
			}
		}
	}

	// Add to new parent
	parent.ChildIDs = append(parent.ChildIDs, childID)
	child.ParentID = parentID
	return true
}

// RemoveChild removes a child widget from a parent
func (r *WidgetRegistry) RemoveChild(parentID, childID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	parent, ok := r.widgets[parentID]
	if !ok {
		return false
	}
	child, ok := r.widgets[childID]
	if !ok {
		return false
	}

	// Remove from parent's children
	for i, id := range parent.ChildIDs {
		if id == childID {
			parent.ChildIDs = append(parent.ChildIDs[:i], parent.ChildIDs[i+1:]...)
			child.ParentID = ""
			return true
		}
	}
	return false
}

// SetRoot sets the root widget for a session
func (r *WidgetRegistry) SetRoot(sessionID, widgetID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.widgets[widgetID]; !ok {
		return false
	}
	r.sessionRoots[sessionID] = widgetID
	return true
}

// GetRoot returns the root widget ID for a session
func (r *WidgetRegistry) GetRoot(sessionID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	widgetID, ok := r.sessionRoots[sessionID]
	return widgetID, ok
}

// Delete removes a widget and its children recursively
func (r *WidgetRegistry) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.deleteRecursive(id)
}

// deleteRecursive removes a widget and all its children (must be called with lock held)
func (r *WidgetRegistry) deleteRecursive(id string) bool {
	info, ok := r.widgets[id]
	if !ok {
		return false
	}

	// Delete all children first
	for _, childID := range info.ChildIDs {
		r.deleteRecursive(childID)
	}

	// Remove from parent's children list
	if info.ParentID != "" {
		parent, ok := r.widgets[info.ParentID]
		if ok {
			for i, childID := range parent.ChildIDs {
				if childID == id {
					parent.ChildIDs = append(parent.ChildIDs[:i], parent.ChildIDs[i+1:]...)
					break
				}
			}
		}
	}

	// Remove from session root if it is one
	for sessionID, rootID := range r.sessionRoots {
		if rootID == id {
			delete(r.sessionRoots, sessionID)
		}
	}

	delete(r.widgets, id)
	return true
}

// DeleteBySession removes all widgets owned by a session
func (r *WidgetRegistry) DeleteBySession(sessionID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	// Collect IDs to delete (to avoid modifying map while iterating)
	toDelete := make([]string, 0)
	for id, info := range r.widgets {
		if info.SessionID == sessionID {
			toDelete = append(toDelete, id)
		}
	}

	// Delete each widget
	for _, id := range toDelete {
		if r.deleteRecursive(id) {
			count++
		}
	}

	// Clean up session root
	delete(r.sessionRoots, sessionID)
	return count
}

// ListBySession returns all widget infos owned by a session
func (r *WidgetRegistry) ListBySession(sessionID string) []*WidgetInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]*WidgetInfo, 0)
	for _, info := range r.widgets {
		if info.SessionID == sessionID {
			infos = append(infos, info)
		}
	}
	return infos
}

// Count returns the total number of widgets
func (r *WidgetRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.widgets)
}

