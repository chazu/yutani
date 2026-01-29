package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// TreeView represents a tree view widget.
type TreeView struct {
	baseWidget
}

// TreeViewBuilder provides a fluent interface for creating tree view widgets.
type TreeViewBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewTreeView creates a new tree view builder.
func (c *Client) NewTreeView() *TreeViewBuilder {
	return &TreeViewBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the tree view title.
func (b *TreeViewBuilder) Title(title string) *TreeViewBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the tree view has a border.
func (b *TreeViewBuilder) Border(border bool) *TreeViewBuilder {
	b.props.Border = &border
	return b
}

// NodeTextColor sets the node text color.
func (b *TreeViewBuilder) NodeTextColor(color *pb.Color) *TreeViewBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Tree{
			Tree: &pb.TreeProperties{},
		}
	}
	b.props.GetTree().NodeTextColor = color
	return b
}

// SelectedTextColor sets the selected node text color.
func (b *TreeViewBuilder) SelectedTextColor(color *pb.Color) *TreeViewBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Tree{
			Tree: &pb.TreeProperties{},
		}
	}
	b.props.GetTree().SelectedTextColor = color
	return b
}

// SelectedBackgroundColor sets the selected node background color.
func (b *TreeViewBuilder) SelectedBackgroundColor(color *pb.Color) *TreeViewBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Tree{
			Tree: &pb.TreeProperties{},
		}
	}
	b.props.GetTree().SelectedBackgroundColor = color
	return b
}

// TopLevelPrefix sets the prefix for top-level nodes.
func (b *TreeViewBuilder) TopLevelPrefix(prefix string) *TreeViewBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Tree{
			Tree: &pb.TreeProperties{},
		}
	}
	b.props.GetTree().TopLevelPrefix = &prefix
	return b
}

// ShowGraphics sets whether to show tree graphics.
func (b *TreeViewBuilder) ShowGraphics(show bool) *TreeViewBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Tree{
			Tree: &pb.TreeProperties{},
		}
	}
	b.props.GetTree().ShowGraphics = &show
	return b
}

// Build creates the tree view widget on the server.
func (b *TreeViewBuilder) Build() (*TreeView, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_TREE_VIEW,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	treeView := &TreeView{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(treeView)
	return treeView, nil
}

// SetRoot sets the root node of the tree.
func (tv *TreeView) SetRoot(node *pb.TreeNode) (*pb.TreeNodeId, error) {
	resp, err := tv.client.treeClient.SetRoot(tv.client.ctx, &pb.SetTreeRootRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
		Node:      node,
	})
	if err != nil {
		return nil, err
	}
	return resp.NodeId, nil
}

// AddChild adds a child node to a parent node.
func (tv *TreeView) AddChild(parentID *pb.TreeNodeId, node *pb.TreeNode) (*pb.TreeNodeId, error) {
	resp, err := tv.client.treeClient.AddChild(tv.client.ctx, &pb.AddTreeChildRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
		ParentId:  parentID,
		Node:      node,
	})
	if err != nil {
		return nil, err
	}
	return resp.NodeId, nil
}

// RemoveNode removes a node from the tree.
func (tv *TreeView) RemoveNode(nodeID *pb.TreeNodeId) error {
	_, err := tv.client.treeClient.RemoveNode(tv.client.ctx, &pb.RemoveTreeNodeRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
		NodeId:    nodeID,
	})
	return err
}

// SetExpanded expands or collapses a node.
func (tv *TreeView) SetExpanded(nodeID *pb.TreeNodeId, expanded bool) error {
	_, err := tv.client.treeClient.SetExpanded(tv.client.ctx, &pb.TreeSetExpandedRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
		NodeId:    nodeID,
		Expanded:  expanded,
	})
	return err
}

// GetSelection gets the currently selected node.
func (tv *TreeView) GetSelection() (*pb.TreeNodeId, string, string, error) {
	resp, err := tv.client.treeClient.GetSelection(tv.client.ctx, &pb.TreeGetSelectionRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
	})
	if err != nil {
		return nil, "", "", err
	}
	return resp.NodeId, resp.Text, resp.Reference, nil
}

// SetSelection sets the currently selected node.
func (tv *TreeView) SetSelection(nodeID *pb.TreeNodeId) error {
	_, err := tv.client.treeClient.SetSelection(tv.client.ctx, &pb.TreeSetSelectionRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
		NodeId:    nodeID,
	})
	return err
}

// GetChildren gets the children of a node.
func (tv *TreeView) GetChildren(nodeID *pb.TreeNodeId) ([]*pb.TreeNodeInfo, error) {
	resp, err := tv.client.treeClient.GetChildren(tv.client.ctx, &pb.TreeGetChildrenRequest{
		SessionId: tv.client.sessionID,
		WidgetId:  tv.widgetID,
		NodeId:    nodeID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Children, nil
}

// NewTreeNode creates a new tree node with the given text.
func NewTreeNode(text string) *pb.TreeNode {
	return &pb.TreeNode{
		Text: text,
	}
}

// TreeNodeWithColor creates a new tree node with text and color.
func TreeNodeWithColor(text string, color *pb.Color) *pb.TreeNode {
	return &pb.TreeNode{
		Text:  text,
		Color: color,
	}
}

// TreeNodeWithOptions creates a new tree node with all options.
func TreeNodeWithOptions(text string, color *pb.Color, selectable, expanded bool, reference string) *pb.TreeNode {
	node := &pb.TreeNode{
		Text: text,
	}
	if color != nil {
		node.Color = color
	}
	node.Selectable = &selectable
	node.Expanded = &expanded
	if reference != "" {
		node.Reference = &reference
	}
	return node
}

