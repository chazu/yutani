package services

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	"github.com/rivo/tview"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
)

// TreeService implements the TreeService gRPC service
type TreeService struct {
	pb.UnimplementedTreeServiceServer
	server *server.Server

	// Map widget IDs to their node registries
	nodeRegistries map[string]*treeNodeRegistry
	mu             sync.RWMutex
}

// treeNodeRegistry manages node IDs for a single TreeView widget
type treeNodeRegistry struct {
	nodes map[string]*tview.TreeNode // node ID -> TreeNode
	mu    sync.RWMutex
}

func newTreeNodeRegistry() *treeNodeRegistry {
	return &treeNodeRegistry{
		nodes: make(map[string]*tview.TreeNode),
	}
}

func (r *treeNodeRegistry) register(node *tview.TreeNode) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	nodeID := uuid.New().String()
	r.nodes[nodeID] = node
	// Store the node ID in the node's reference for later retrieval
	node.SetReference(nodeID)
	return nodeID
}

func (r *treeNodeRegistry) get(nodeID string) (*tview.TreeNode, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, ok := r.nodes[nodeID]
	return node, ok
}

func (r *treeNodeRegistry) remove(nodeID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.nodes, nodeID)
}

// NewTreeService creates a new TreeService
func NewTreeService(srv *server.Server) *TreeService {
	return &TreeService{
		server:         srv,
		nodeRegistries: make(map[string]*treeNodeRegistry),
	}
}

func (s *TreeService) getOrCreateRegistry(widgetID string) *treeNodeRegistry {
	s.mu.Lock()
	defer s.mu.Unlock()

	if registry, ok := s.nodeRegistries[widgetID]; ok {
		return registry
	}

	registry := newTreeNodeRegistry()
	s.nodeRegistries[widgetID] = registry
	return registry
}

// createTreeNode creates a tview.TreeNode from a protobuf TreeNode
func createTreeNode(node *pb.TreeNode) *tview.TreeNode {
	treeNode := tview.NewTreeNode(node.Text)

	if node.Color != nil {
		treeNode.SetColor(convertColor(node.Color))
	}

	if node.Selectable != nil {
		treeNode.SetSelectable(*node.Selectable)
	} else {
		treeNode.SetSelectable(true) // Default to selectable
	}

	if node.Expanded != nil {
		treeNode.SetExpanded(*node.Expanded)
	}

	return treeNode
}

// SetRoot sets the root node of the tree
func (s *TreeService) SetRoot(ctx context.Context, req *pb.SetTreeRootRequest) (*pb.SetTreeRootResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.Node == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, and node are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var nodeID string
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		tree, ok := primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Create the root node
		rootNode := createTreeNode(req.Node)

		// Register the node
		registry := s.getOrCreateRegistry(widgetID)
		nodeID = registry.register(rootNode)

		// Set as tree root
		tree.SetRoot(rootNode)

		slog.Info("Tree root set", "widget_id", widgetID, "node_id", nodeID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetTreeRootResponse{
		Success: true,
		NodeId:  &pb.TreeNodeId{Id: nodeID},
	}, nil
}

// AddChild adds a child node to a parent node
func (s *TreeService) AddChild(ctx context.Context, req *pb.AddTreeChildRequest) (*pb.AddTreeChildResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.ParentId == nil || req.Node == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, parent_id, and node are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id
	parentNodeID := req.ParentId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var nodeID string
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		_, ok = primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Get the parent node
		registry := s.getOrCreateRegistry(widgetID)
		parentNode, ok := registry.get(parentNodeID)
		if !ok {
			err = status.Error(codes.NotFound, "parent node not found")
			return
		}

		// Create the child node
		childNode := createTreeNode(req.Node)

		// Register the child node
		nodeID = registry.register(childNode)

		// Add to parent
		parentNode.AddChild(childNode)

		slog.Info("Tree child added", "widget_id", widgetID, "parent_id", parentNodeID, "node_id", nodeID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.AddTreeChildResponse{
		Success: true,
		NodeId:  &pb.TreeNodeId{Id: nodeID},
	}, nil
}

// RemoveNode removes a node from the tree
func (s *TreeService) RemoveNode(ctx context.Context, req *pb.RemoveTreeNodeRequest) (*pb.RemoveTreeNodeResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.NodeId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, and node_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id
	nodeID := req.NodeId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		tree, ok := primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Get the node
		registry := s.getOrCreateRegistry(widgetID)
		node, ok := registry.get(nodeID)
		if !ok {
			err = status.Error(codes.NotFound, "node not found")
			return
		}

		// Remove from parent (tview doesn't have a direct remove method, so we need to find the parent)
		// We'll walk the tree to find and remove the node
		root := tree.GetRoot()
		if root == nil {
			err = status.Error(codes.NotFound, "tree has no root")
			return
		}

		if !removeNodeFromTree(root, node) {
			err = status.Error(codes.NotFound, "node not found in tree")
			return
		}

		// Unregister the node
		registry.remove(nodeID)

		slog.Info("Tree node removed", "widget_id", widgetID, "node_id", nodeID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.RemoveTreeNodeResponse{
		Success: true,
	}, nil
}

// removeNodeFromTree recursively searches for and removes a node from the tree
func removeNodeFromTree(parent, target *tview.TreeNode) bool {
	children := parent.GetChildren()
	for i, child := range children {
		if child == target {
			// Remove this child
			newChildren := append(children[:i], children[i+1:]...)
			parent.SetChildren(newChildren)
			return true
		}
		// Recursively search in children
		if removeNodeFromTree(child, target) {
			return true
		}
	}
	return false
}

// SetExpanded expands or collapses a node
func (s *TreeService) SetExpanded(ctx context.Context, req *pb.SetExpandedRequest) (*pb.SetExpandedResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.NodeId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, and node_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id
	nodeID := req.NodeId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		_, ok = primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Get the node
		registry := s.getOrCreateRegistry(widgetID)
		node, ok := registry.get(nodeID)
		if !ok {
			err = status.Error(codes.NotFound, "node not found")
			return
		}

		// Set expanded state
		node.SetExpanded(req.Expanded)

		slog.Info("Tree node expanded state set", "widget_id", widgetID, "node_id", nodeID, "expanded", req.Expanded)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetExpandedResponse{
		Success: true,
	}, nil
}

// GetSelected gets the currently selected node
func (s *TreeService) GetSelected(ctx context.Context, req *pb.GetTreeSelectedRequest) (*pb.GetTreeSelectedResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id and widget_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var nodeID string
	var text string
	var reference string
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		tree, ok := primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Get the current node
		currentNode := tree.GetCurrentNode()
		if currentNode == nil {
			err = status.Error(codes.NotFound, "no node selected")
			return
		}

		// Get the node ID from the reference
		ref := currentNode.GetReference()
		if ref != nil {
			if id, ok := ref.(string); ok {
				nodeID = id
			}
		}

		text = currentNode.GetText()
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetTreeSelectedResponse{
		NodeId:    &pb.TreeNodeId{Id: nodeID},
		Text:      text,
		Reference: reference,
	}, nil
}

// SetSelected sets the currently selected node
func (s *TreeService) SetSelected(ctx context.Context, req *pb.SetTreeSelectedRequest) (*pb.SetTreeSelectedResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.NodeId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, and node_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id
	nodeID := req.NodeId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		tree, ok := primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Get the node
		registry := s.getOrCreateRegistry(widgetID)
		node, ok := registry.get(nodeID)
		if !ok {
			err = status.Error(codes.NotFound, "node not found")
			return
		}

		// Set as current node
		tree.SetCurrentNode(node)

		slog.Info("Tree node selected", "widget_id", widgetID, "node_id", nodeID)
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.SetTreeSelectedResponse{
		Success: true,
	}, nil
}

// GetChildren gets the children of a node
func (s *TreeService) GetChildren(ctx context.Context, req *pb.GetChildrenRequest) (*pb.GetChildrenResponse, error) {
	if req.SessionId == nil || req.WidgetId == nil || req.NodeId == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, widget_id, and node_id are required")
	}

	sessionID := req.SessionId.Id
	widgetID := req.WidgetId.Id
	nodeID := req.NodeId.Id

	// Verify ownership
	owner, ok := s.server.Widgets().GetOwner(widgetID)
	if !ok {
		return nil, status.Error(codes.NotFound, "widget not found")
	}
	if owner != sessionID {
		return nil, status.Error(codes.PermissionDenied, "widget not owned by session")
	}

	var children []*pb.TreeNodeInfo
	var err error
	done := make(chan struct{})
	s.server.App().QueueUpdateDraw(func() {
		defer close(done)

		primitive, ok := s.server.Widgets().Get(widgetID)
		if !ok {
			err = status.Error(codes.NotFound, "widget not found")
			return
		}

		_, ok = primitive.(*tview.TreeView)
		if !ok {
			err = status.Error(codes.InvalidArgument, "widget is not a tree view")
			return
		}

		// Get the node
		registry := s.getOrCreateRegistry(widgetID)
		node, ok := registry.get(nodeID)
		if !ok {
			err = status.Error(codes.NotFound, "node not found")
			return
		}

		// Get children
		childNodes := node.GetChildren()
		children = make([]*pb.TreeNodeInfo, 0, len(childNodes))

		for _, child := range childNodes {
			var childNodeID string
			var childReference string

			// Get the node ID from the reference
			ref := child.GetReference()
			if ref != nil {
				if id, ok := ref.(string); ok {
					childNodeID = id
				}
			}

			children = append(children, &pb.TreeNodeInfo{
				NodeId:    &pb.TreeNodeId{Id: childNodeID},
				Text:      child.GetText(),
				Reference: childReference,
				Expanded:  child.IsExpanded(),
			})
		}
	})
	<-done

	if err != nil {
		return nil, err
	}

	return &pb.GetChildrenResponse{
		Children: children,
	}, nil
}

