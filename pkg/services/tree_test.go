package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
)

func TestTreeService_SetRoot(t *testing.T) {
	srv := setupTestServer(t)

	treeService := NewTreeService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a tree widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Test SetRoot
	resp, err := treeService.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Node: &pb.TreeNode{
			Text: "Root Node",
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.NodeId == nil || resp.NodeId.Id == "" {
		t.Error("Expected node_id to be set")
	}
}

func TestTreeService_AddChild(t *testing.T) {
	srv := setupTestServer(t)

	treeService := NewTreeService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a tree widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Set root
	rootResp, err := treeService.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Node: &pb.TreeNode{
			Text: "Root",
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	rootNodeID := rootResp.NodeId.Id

	// Test AddChild
	resp, err := treeService.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		ParentId:  &pb.TreeNodeId{Id: rootNodeID},
		Node: &pb.TreeNode{
			Text: "Child Node",
		},
	})
	if err != nil {
		t.Fatalf("AddChild failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
	if resp.NodeId == nil || resp.NodeId.Id == "" {
		t.Error("Expected node_id to be set")
	}

	// Add another child
	resp2, err := treeService.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		ParentId:  &pb.TreeNodeId{Id: rootNodeID},
		Node: &pb.TreeNode{
			Text: "Child Node 2",
		},
	})
	if err != nil {
		t.Fatalf("AddChild (2nd) failed: %v", err)
	}
	if resp2.NodeId.Id == resp.NodeId.Id {
		t.Error("Expected different node IDs for different children")
	}
}

func TestTreeService_RemoveNode(t *testing.T) {
	srv := setupTestServer(t)

	treeService := NewTreeService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a tree widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Set root
	rootResp, err := treeService.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Node: &pb.TreeNode{
			Text: "Root",
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	rootNodeID := rootResp.NodeId.Id

	// Add a child
	childResp, err := treeService.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		ParentId:  &pb.TreeNodeId{Id: rootNodeID},
		Node: &pb.TreeNode{
			Text: "Child",
		},
	})
	if err != nil {
		t.Fatalf("AddChild failed: %v", err)
	}
	childNodeID := childResp.NodeId.Id

	// Test RemoveNode
	resp, err := treeService.RemoveNode(ctx, &pb.RemoveTreeNodeRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		NodeId:    &pb.TreeNodeId{Id: childNodeID},
	})
	if err != nil {
		t.Fatalf("RemoveNode failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestTreeService_SetExpanded(t *testing.T) {
	srv := setupTestServer(t)

	treeService := NewTreeService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a tree widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Set root
	rootResp, err := treeService.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Node: &pb.TreeNode{
			Text: "Root",
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	rootNodeID := rootResp.NodeId.Id

	// Test SetExpanded - expand
	resp, err := treeService.SetExpanded(ctx, &pb.TreeSetExpandedRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		NodeId:    &pb.TreeNodeId{Id: rootNodeID},
		Expanded:  true,
	})
	if err != nil {
		t.Fatalf("SetExpanded (expand) failed: %v", err)
	}
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Test SetExpanded - collapse
	resp2, err := treeService.SetExpanded(ctx, &pb.TreeSetExpandedRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		NodeId:    &pb.TreeNodeId{Id: rootNodeID},
		Expanded:  false,
	})
	if err != nil {
		t.Fatalf("SetExpanded (collapse) failed: %v", err)
	}
	if !resp2.Success {
		t.Error("Expected success to be true")
	}
}

func TestTreeService_GetSetSelection(t *testing.T) {
	srv := setupTestServer(t)

	treeService := NewTreeService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a tree widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Set root
	rootResp, err := treeService.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Node: &pb.TreeNode{
			Text: "Root",
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	rootNodeID := rootResp.NodeId.Id

	// Add a child
	childResp, err := treeService.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		ParentId:  &pb.TreeNodeId{Id: rootNodeID},
		Node: &pb.TreeNode{
			Text: "Child Node",
		},
	})
	if err != nil {
		t.Fatalf("AddChild failed: %v", err)
	}
	childNodeID := childResp.NodeId.Id

	// Test SetSelection
	setResp, err := treeService.SetSelection(ctx, &pb.TreeSetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		NodeId:    &pb.TreeNodeId{Id: childNodeID},
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}
	if !setResp.Success {
		t.Error("Expected success to be true")
	}

	// Test GetSelection
	getResp, err := treeService.GetSelection(ctx, &pb.TreeGetSelectionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	if getResp.NodeId.Id != childNodeID {
		t.Errorf("Expected selected node ID %s, got %s", childNodeID, getResp.NodeId.Id)
	}
	if getResp.Text != "Child Node" {
		t.Errorf("Expected text 'Child Node', got %q", getResp.Text)
	}
}

func TestTreeService_GetChildren(t *testing.T) {
	srv := setupTestServer(t)

	treeService := NewTreeService(srv)
	widgetService := NewWidgetService(srv)

	ctx := context.Background()

	// Create a session
	session, err := srv.Sessions().Create("test-client", server.SessionPreferences{})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	sessionID := session.ID

	// Create a tree widget
	createResp, err := widgetService.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
	})
	if err != nil {
		t.Fatalf("CreateWidget failed: %v", err)
	}
	widgetID := createResp.WidgetId.Id

	// Set root
	rootResp, err := treeService.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		Node: &pb.TreeNode{
			Text: "Root",
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	rootNodeID := rootResp.NodeId.Id

	// Test GetChildren with no children
	resp, err := treeService.GetChildren(ctx, &pb.TreeGetChildrenRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		NodeId:    &pb.TreeNodeId{Id: rootNodeID},
	})
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}
	if len(resp.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(resp.Children))
	}

	// Add some children
	_, err = treeService.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		ParentId:  &pb.TreeNodeId{Id: rootNodeID},
		Node: &pb.TreeNode{
			Text: "Child 1",
		},
	})
	if err != nil {
		t.Fatalf("AddChild (1) failed: %v", err)
	}

	_, err = treeService.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		ParentId:  &pb.TreeNodeId{Id: rootNodeID},
		Node: &pb.TreeNode{
			Text: "Child 2",
		},
	})
	if err != nil {
		t.Fatalf("AddChild (2) failed: %v", err)
	}

	// Test GetChildren with 2 children
	resp2, err := treeService.GetChildren(ctx, &pb.TreeGetChildrenRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: widgetID},
		NodeId:    &pb.TreeNodeId{Id: rootNodeID},
	})
	if err != nil {
		t.Fatalf("GetChildren (2nd) failed: %v", err)
	}
	if len(resp2.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(resp2.Children))
	}
	if resp2.Children[0].Text != "Child 1" {
		t.Errorf("Expected first child text 'Child 1', got %q", resp2.Children[0].Text)
	}
	if resp2.Children[1].Text != "Child 2" {
		t.Errorf("Expected second child text 'Child 2', got %q", resp2.Children[1].Text)
	}
}

