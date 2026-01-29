package e2e

import (
	"context"
	"testing"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/testutil"
)

// TestAcceptance_SessionLifecycle exercises the full session lifecycle:
// create, verify listed, ping, create widget, set root, destroy, verify gone.
func TestAcceptance_SessionLifecycle(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)

	// Step 1: Create a session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-lifecycle",
		Preferences: &pb.SessionPreferences{
			ReceiveKeyEvents:    true,
			ReceiveMouseEvents:  true,
			ReceiveResizeEvents: true,
		},
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	if sessionID.Id == "" {
		t.Fatal("Expected non-empty session ID")
	}

	// Step 2: Verify the session appears in ListSessions
	listResp, err := sessionClient.ListSessions(ctx, &pb.ListSessionsRequest{})
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	found := false
	for _, s := range listResp.Sessions {
		if s.SessionId.Id == sessionID.Id {
			found = true
			if s.ClientName != "acceptance-lifecycle" {
				t.Errorf("Expected client name 'acceptance-lifecycle', got '%s'", s.ClientName)
			}
			break
		}
	}
	if !found {
		t.Error("Session not found in ListSessions after creation")
	}

	// Step 3: Ping the server
	pingResp, err := sessionClient.Ping(ctx, &pb.PingRequest{
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if pingResp.ServerTimestamp == 0 {
		t.Error("Expected non-zero server timestamp from Ping")
	}

	// Step 4: Create a Box widget
	boxResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Lifecycle Box"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Box) failed: %v", err)
	}
	boxID := boxResp.WidgetId
	if boxID.Id == "" {
		t.Fatal("Expected non-empty widget ID for Box")
	}

	// Step 5: Set the Box as root
	setRootResp, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
		SessionId: sessionID,
		WidgetId:  boxID,
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	if !setRootResp.Success {
		t.Error("Expected successful SetRoot")
	}

	// Step 6: Destroy the session
	destroyResp, err := sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("DestroySession failed: %v", err)
	}
	if !destroyResp.Success {
		t.Error("Expected successful DestroySession")
	}

	// Step 7: Verify the session is gone from ListSessions
	listResp2, err := sessionClient.ListSessions(ctx, &pb.ListSessionsRequest{})
	if err != nil {
		t.Fatalf("ListSessions (after destroy) failed: %v", err)
	}
	for _, s := range listResp2.Sessions {
		if s.SessionId.Id == sessionID.Id {
			t.Error("Session still present in ListSessions after destruction")
		}
	}
}

// TestAcceptance_InteractiveForm exercises a complete form workflow:
// create form, add fields, add button, set/get values, clear.
func TestAcceptance_InteractiveForm(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	formClient := pb.NewFormServiceClient(conn)

	// Create session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-form",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})

	// Create Form widget
	formResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FORM,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Registration Form"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Form) failed: %v", err)
	}
	formID := formResp.WidgetId

	// Add field 1: text input
	addField1, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
		SessionId:    sessionID,
		WidgetId:     formID,
		Label:        "Full Name",
		FieldType:    pb.FormFieldType_FORM_FIELD_INPUT,
		FieldWidth:   testutil.Int32Ptr(30),
		InitialValue: testutil.StrPtr("John Doe"),
	})
	if err != nil {
		t.Fatalf("AddField (text input) failed: %v", err)
	}
	if addField1.FieldIndex != 0 {
		t.Errorf("Expected field index 0, got %d", addField1.FieldIndex)
	}

	// Add field 2: password
	addField2, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		Label:      "Password",
		FieldType:  pb.FormFieldType_FORM_FIELD_PASSWORD,
		FieldWidth: testutil.Int32Ptr(30),
	})
	if err != nil {
		t.Fatalf("AddField (password) failed: %v", err)
	}
	if addField2.FieldIndex != 1 {
		t.Errorf("Expected field index 1, got %d", addField2.FieldIndex)
	}

	// Add field 3: checkbox
	addField3, err := formClient.AddField(ctx, &pb.FormAddFieldRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Accept Terms",
		FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
	})
	if err != nil {
		t.Fatalf("AddField (checkbox) failed: %v", err)
	}
	if addField3.FieldIndex != 2 {
		t.Errorf("Expected field index 2, got %d", addField3.FieldIndex)
	}

	// Add a button
	addBtnResp, err := formClient.AddButton(ctx, &pb.FormAddButtonRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Submit",
	})
	if err != nil {
		t.Fatalf("AddButton failed: %v", err)
	}
	if !addBtnResp.Success {
		t.Error("Expected successful AddButton")
	}

	// Set field values
	_, err = formClient.SetFieldValue(ctx, &pb.FormSetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 0,
		Value:      "Jane Smith",
	})
	if err != nil {
		t.Fatalf("SetFieldValue (name) failed: %v", err)
	}

	_, err = formClient.SetFieldValue(ctx, &pb.FormSetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 1,
		Value:      "s3cureP@ss",
	})
	if err != nil {
		t.Fatalf("SetFieldValue (password) failed: %v", err)
	}

	// Get field values and verify
	nameVal, err := formClient.GetFieldValue(ctx, &pb.FormGetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 0,
	})
	if err != nil {
		t.Fatalf("GetFieldValue (name) failed: %v", err)
	}
	if nameVal.Value != "Jane Smith" {
		t.Errorf("Expected name 'Jane Smith', got '%s'", nameVal.Value)
	}

	passVal, err := formClient.GetFieldValue(ctx, &pb.FormGetFieldValueRequest{
		SessionId:  sessionID,
		WidgetId:   formID,
		FieldIndex: 1,
	})
	if err != nil {
		t.Fatalf("GetFieldValue (password) failed: %v", err)
	}
	if passVal.Value != "s3cureP@ss" {
		t.Errorf("Expected password 's3cureP@ss', got '%s'", passVal.Value)
	}

	// Clear the form
	clearResp, err := formClient.Clear(ctx, &pb.FormClearRequest{
		SessionId: sessionID,
		WidgetId:  formID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful Clear")
	}

	// Verify item count is 0 after clear
	countResp, err := formClient.GetItemCount(ctx, &pb.FormGetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  formID,
	})
	if err != nil {
		t.Fatalf("GetItemCount failed: %v", err)
	}
	if countResp.Count != 0 {
		t.Errorf("Expected item count 0 after clear, got %d", countResp.Count)
	}
}

// TestAcceptance_ListManagement exercises a complete list workflow:
// create list, add items, verify count, selection, get/remove items, clear.
func TestAcceptance_ListManagement(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	listClient := pb.NewListServiceClient(conn)

	// Create session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-list",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})

	// Create List widget
	listResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_LIST,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Task List"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (List) failed: %v", err)
	}
	listID := listResp.WidgetId

	// Add 5 items with main_text and secondary_text
	items := []struct {
		main      string
		secondary string
	}{
		{"Buy groceries", "Milk, eggs, bread"},
		{"Write report", "Q4 financial summary"},
		{"Call dentist", "Schedule appointment"},
		{"Fix bug #42", "Null pointer in parser"},
		{"Deploy v2.0", "Production release"},
	}

	for i, item := range items {
		addResp, err := listClient.AddItem(ctx, &pb.ListAddItemRequest{
			SessionId:     sessionID,
			WidgetId:      listID,
			MainText:      item.main,
			SecondaryText: item.secondary,
		})
		if err != nil {
			t.Fatalf("AddItem (%d) failed: %v", i, err)
		}
		if addResp.Index != int32(i) {
			t.Errorf("Expected index %d, got %d", i, addResp.Index)
		}
	}

	// Verify item count is 5
	countResp, err := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetItemCount failed: %v", err)
	}
	if countResp.Count != 5 {
		t.Errorf("Expected count 5, got %d", countResp.Count)
	}

	// Get selection (should be 0 for a fresh list with items, or -1 if no selection)
	// tview lists default to selecting the first item when items are added
	getSelResp, err := listClient.GetSelection(ctx, &pb.ListGetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	// A fresh list with items selects index 0 by default in tview
	t.Logf("Initial selection index: %d", getSelResp.Index)

	// Set selection to index 2
	setSelResp, err := listClient.SetSelection(ctx, &pb.ListSetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  listID,
		Index:     2,
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}
	if !setSelResp.Success {
		t.Error("Expected successful SetSelection")
	}

	// Get selection and verify it's 2
	getSelResp2, err := listClient.GetSelection(ctx, &pb.ListGetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetSelection (after set) failed: %v", err)
	}
	if getSelResp2.Index != 2 {
		t.Errorf("Expected selection index 2, got %d", getSelResp2.Index)
	}

	// Get item at index 2 and verify text
	itemResp, err := listClient.GetItem(ctx, &pb.ListGetItemRequest{
		SessionId: sessionID,
		WidgetId:  listID,
		Index:     2,
	})
	if err != nil {
		t.Fatalf("GetItem (index 2) failed: %v", err)
	}
	if itemResp.MainText != "Call dentist" {
		t.Errorf("Expected main text 'Call dentist', got '%s'", itemResp.MainText)
	}
	if itemResp.SecondaryText != "Schedule appointment" {
		t.Errorf("Expected secondary text 'Schedule appointment', got '%s'", itemResp.SecondaryText)
	}

	// Remove item at index 0
	removeResp, err := listClient.RemoveItem(ctx, &pb.ListRemoveItemRequest{
		SessionId: sessionID,
		WidgetId:  listID,
		Index:     0,
	})
	if err != nil {
		t.Fatalf("RemoveItem failed: %v", err)
	}
	if !removeResp.Success {
		t.Error("Expected successful RemoveItem")
	}

	// Verify count decreased to 4
	countResp2, err := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetItemCount (after remove) failed: %v", err)
	}
	if countResp2.Count != 4 {
		t.Errorf("Expected count 4 after removal, got %d", countResp2.Count)
	}

	// Clear the list
	clearResp, err := listClient.Clear(ctx, &pb.ListClearRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful Clear")
	}

	// Verify count is 0
	countResp3, err := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetItemCount (after clear) failed: %v", err)
	}
	if countResp3.Count != 0 {
		t.Errorf("Expected count 0 after clear, got %d", countResp3.Count)
	}
}

// TestAcceptance_TableData exercises a complete table workflow:
// create table, set cells (bulk), get dimensions, get cell, selection, fixed rows, clear.
func TestAcceptance_TableData(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	tableClient := pb.NewTableServiceClient(conn)

	// Create session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-table",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})

	// Create Table widget
	tableResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TABLE,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Data Table"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Table) failed: %v", err)
	}
	tableID := tableResp.WidgetId

	// Set cells in bulk: 3x3 grid with headers in row 0
	setCellsResp, err := tableClient.SetCells(ctx, &pb.SetTableCellsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Cells: []*pb.TableCellUpdate{
			// Header row
			{Row: 0, Column: 0, Cell: &pb.TableCell{Text: "Name"}},
			{Row: 0, Column: 1, Cell: &pb.TableCell{Text: "Age"}},
			{Row: 0, Column: 2, Cell: &pb.TableCell{Text: "City"}},
			// Data row 1
			{Row: 1, Column: 0, Cell: &pb.TableCell{Text: "Alice"}},
			{Row: 1, Column: 1, Cell: &pb.TableCell{Text: "30"}},
			{Row: 1, Column: 2, Cell: &pb.TableCell{Text: "New York"}},
			// Data row 2
			{Row: 2, Column: 0, Cell: &pb.TableCell{Text: "Bob"}},
			{Row: 2, Column: 1, Cell: &pb.TableCell{Text: "25"}},
			{Row: 2, Column: 2, Cell: &pb.TableCell{Text: "London"}},
		},
	})
	if err != nil {
		t.Fatalf("SetCells failed: %v", err)
	}
	if !setCellsResp.Success {
		t.Error("Expected successful SetCells")
	}
	if setCellsResp.CellsSet != 9 {
		t.Errorf("Expected 9 cells set, got %d", setCellsResp.CellsSet)
	}

	// Get dimensions and verify 3 rows, 3 columns
	dimsResp, err := tableClient.GetDimensions(ctx, &pb.TableGetDimensionsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("GetDimensions failed: %v", err)
	}
	if dimsResp.Rows != 3 {
		t.Errorf("Expected 3 rows, got %d", dimsResp.Rows)
	}
	if dimsResp.Columns != 3 {
		t.Errorf("Expected 3 columns, got %d", dimsResp.Columns)
	}

	// Get a specific cell and verify content
	cellResp, err := tableClient.GetCell(ctx, &pb.GetTableCellRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       1,
		Column:    2,
	})
	if err != nil {
		t.Fatalf("GetCell failed: %v", err)
	}
	if cellResp.Cell.Text != "New York" {
		t.Errorf("Expected cell (1,2) text 'New York', got '%s'", cellResp.Cell.Text)
	}

	// Set selection to row 1, col 1
	setSelResp, err := tableClient.SetSelection(ctx, &pb.TableSetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Row:       1,
		Column:    1,
	})
	if err != nil {
		t.Fatalf("SetSelection failed: %v", err)
	}
	if !setSelResp.Success {
		t.Error("Expected successful SetSelection")
	}

	// Get selection and verify
	getSelResp, err := tableClient.GetSelection(ctx, &pb.TableGetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	if getSelResp.Row != 1 || getSelResp.Column != 1 {
		t.Errorf("Expected selection (1,1), got (%d,%d)", getSelResp.Row, getSelResp.Column)
	}

	// Set fixed rows=1 for header
	setFixedResp, err := tableClient.SetFixed(ctx, &pb.TableSetFixedRequest{
		SessionId:    sessionID,
		WidgetId:     tableID,
		FixedRows:    1,
		FixedColumns: 0,
	})
	if err != nil {
		t.Fatalf("SetFixed failed: %v", err)
	}
	if !setFixedResp.Success {
		t.Error("Expected successful SetFixed")
	}

	// Clear the table
	clearResp, err := tableClient.Clear(ctx, &pb.TableClearRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if !clearResp.Success {
		t.Error("Expected successful Clear")
	}

	// Verify dimensions are 0x0 after clear
	dimsResp2, err := tableClient.GetDimensions(ctx, &pb.TableGetDimensionsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("GetDimensions (after clear) failed: %v", err)
	}
	if dimsResp2.Rows != 0 || dimsResp2.Columns != 0 {
		t.Errorf("Expected 0x0 after clear, got %dx%d", dimsResp2.Rows, dimsResp2.Columns)
	}
}

// TestAcceptance_TreeNavigation exercises a complete tree workflow:
// create tree, set root, add children, get children, expand/collapse, selection, remove.
func TestAcceptance_TreeNavigation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	treeClient := pb.NewTreeServiceClient(conn)

	// Create session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-tree",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})

	// Create TreeView widget
	treeResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("File Browser"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (TreeView) failed: %v", err)
	}
	treeID := treeResp.WidgetId

	// Set root node
	setRootResp, err := treeClient.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		Node: &pb.TreeNode{
			Text:       "root",
			Selectable: testutil.BoolPtr(true),
			Expanded:   testutil.BoolPtr(true),
			Reference:  testutil.StrPtr("root-ref"),
		},
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	if !setRootResp.Success {
		t.Error("Expected successful SetRoot")
	}
	rootID := setRootResp.NodeId

	// Add 3 child nodes to root
	childNames := []string{"src", "pkg", "test"}
	childIDs := make([]*pb.TreeNodeId, 3)

	for i, name := range childNames {
		addChildResp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
			SessionId: sessionID,
			WidgetId:  treeID,
			ParentId:  rootID,
			Node: &pb.TreeNode{
				Text:       name,
				Selectable: testutil.BoolPtr(true),
				Reference:  testutil.StrPtr(name + "-ref"),
			},
		})
		if err != nil {
			t.Fatalf("AddChild (%s) failed: %v", name, err)
		}
		if !addChildResp.Success {
			t.Errorf("Expected successful AddChild for %s", name)
		}
		childIDs[i] = addChildResp.NodeId
	}

	// Get children of root and verify 3
	getChildrenResp, err := treeClient.GetChildren(ctx, &pb.TreeGetChildrenRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
	})
	if err != nil {
		t.Fatalf("GetChildren failed: %v", err)
	}
	if len(getChildrenResp.Children) != 3 {
		t.Errorf("Expected 3 children of root, got %d", len(getChildrenResp.Children))
	}

	// Set expanded=false on root (collapse)
	setExpandedResp, err := treeClient.SetExpanded(ctx, &pb.TreeSetExpandedRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
		Expanded:  false,
	})
	if err != nil {
		t.Fatalf("SetExpanded failed: %v", err)
	}
	if !setExpandedResp.Success {
		t.Error("Expected successful SetExpanded")
	}

	// Re-expand root so children are visible for selection
	_, err = treeClient.SetExpanded(ctx, &pb.TreeSetExpandedRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
		Expanded:  true,
	})
	if err != nil {
		t.Fatalf("SetExpanded (re-expand) failed: %v", err)
	}

	// Explicitly select the root node
	setSelRoot, err := treeClient.SetSelection(ctx, &pb.TreeSetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
	})
	if err != nil {
		t.Fatalf("SetSelection (root) failed: %v", err)
	}
	if !setSelRoot.Success {
		t.Error("Expected successful SetSelection for root")
	}

	// Get selection and verify it is the root
	getSelResp, err := treeClient.GetSelection(ctx, &pb.TreeGetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
	})
	if err != nil {
		t.Fatalf("GetSelection failed: %v", err)
	}
	if getSelResp.NodeId.Id != rootID.Id {
		t.Errorf("Expected root node selected, got node %s", getSelResp.NodeId.Id)
	}
	if getSelResp.Text != "root" {
		t.Errorf("Expected selected node text 'root', got '%s'", getSelResp.Text)
	}

	// Set selection to a child node (the second child: "pkg")
	setSelResp, err := treeClient.SetSelection(ctx, &pb.TreeSetSelectionRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    childIDs[1],
	})
	if err != nil {
		t.Fatalf("SetSelection (child) failed: %v", err)
	}
	if !setSelResp.Success {
		t.Error("Expected successful SetSelection")
	}

	// Remove one child ("src" at index 0)
	removeResp, err := treeClient.RemoveNode(ctx, &pb.RemoveTreeNodeRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    childIDs[0],
	})
	if err != nil {
		t.Fatalf("RemoveNode failed: %v", err)
	}
	if !removeResp.Success {
		t.Error("Expected successful RemoveNode")
	}

	// Get children and verify 2 remaining
	getChildrenResp2, err := treeClient.GetChildren(ctx, &pb.TreeGetChildrenRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		NodeId:    rootID,
	})
	if err != nil {
		t.Fatalf("GetChildren (after remove) failed: %v", err)
	}
	if len(getChildrenResp2.Children) != 2 {
		t.Errorf("Expected 2 children after removal, got %d", len(getChildrenResp2.Children))
	}
}

// TestAcceptance_LayoutComposition exercises layout composition with Flex and Pages:
// create flex with children, set direction, create pages with multiple pages, switch pages.
func TestAcceptance_LayoutComposition(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	layoutClient := pb.NewLayoutServiceClient(conn)

	// Create session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-layout",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})

	// --- Flex layout ---

	// Create Flex widget
	flexResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FLEX,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Main Flex"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Flex) failed: %v", err)
	}
	flexID := flexResp.WidgetId

	// Create 2 Box widgets for flex children
	box1Resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Sidebar"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Box 1) failed: %v", err)
	}
	box1ID := box1Resp.WidgetId

	box2Resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Content"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Box 2) failed: %v", err)
	}
	box2ID := box2Resp.WidgetId

	// Add both boxes to Flex with proportions
	addFlex1, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     box1ID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (sidebar) failed: %v", err)
	}
	if !addFlex1.Success {
		t.Error("Expected successful FlexAddItem for sidebar")
	}

	addFlex2, err := layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     box2ID,
		Proportion: 3,
		Focus:      true,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (content) failed: %v", err)
	}
	if !addFlex2.Success {
		t.Error("Expected successful FlexAddItem for content")
	}

	// Set flex direction to vertical (column)
	setDirResp, err := layoutClient.FlexSetDirection(ctx, &pb.FlexSetDirectionRequest{
		SessionId: sessionID,
		FlexId:    flexID,
		Direction: pb.FlexDirection_FLEX_COLUMN,
	})
	if err != nil {
		t.Fatalf("FlexSetDirection failed: %v", err)
	}
	if !setDirResp.Success {
		t.Error("Expected successful FlexSetDirection")
	}

	// --- Pages layout ---

	// Create Pages widget
	pagesResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_PAGES,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Tab Pages"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Pages) failed: %v", err)
	}
	pagesID := pagesResp.WidgetId

	// Create 2 more Box widgets for pages
	pageBox1Resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: testutil.StrPtr("Home Content"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (page box 1) failed: %v", err)
	}
	pageBox1ID := pageBox1Resp.WidgetId

	pageBox2Resp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_BOX,
		Properties: &pb.WidgetProperties{
			Title: testutil.StrPtr("Settings Content"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (page box 2) failed: %v", err)
	}
	pageBox2ID := pageBox2Resp.WidgetId

	// Add as pages
	addPage1, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "home",
		ItemId:    pageBox1ID,
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (home) failed: %v", err)
	}
	if !addPage1.Success {
		t.Error("Expected successful PagesAddPage for home")
	}

	addPage2, err := layoutClient.PagesAddPage(ctx, &pb.PagesAddPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "settings",
		ItemId:    pageBox2ID,
		Resize:    true,
		Visible:   true,
	})
	if err != nil {
		t.Fatalf("PagesAddPage (settings) failed: %v", err)
	}
	if !addPage2.Success {
		t.Error("Expected successful PagesAddPage for settings")
	}

	// Show the second page
	showResp, err := layoutClient.PagesShowPage(ctx, &pb.PagesShowPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
		PageName:  "settings",
	})
	if err != nil {
		t.Fatalf("PagesShowPage failed: %v", err)
	}
	if !showResp.Success {
		t.Error("Expected successful PagesShowPage")
	}

	// Get current page and verify
	getCurrentResp, err := layoutClient.PagesGetCurrentPage(ctx, &pb.PagesGetCurrentPageRequest{
		SessionId: sessionID,
		PagesId:   pagesID,
	})
	if err != nil {
		t.Fatalf("PagesGetCurrentPage failed: %v", err)
	}
	if getCurrentResp.PageName != "settings" {
		t.Errorf("Expected current page 'settings', got '%s'", getCurrentResp.PageName)
	}
}

// TestAcceptance_MultiWidgetDashboard exercises building a multi-widget dashboard:
// create flex layout containing a list, table, and textview, populate data,
// set as root, verify all widgets exist, clean up.
func TestAcceptance_MultiWidgetDashboard(t *testing.T) {
	ts := newTestServer(t)
	defer ts.stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := ts.dial(ctx)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	layoutClient := pb.NewLayoutServiceClient(conn)
	listClient := pb.NewListServiceClient(conn)
	tableClient := pb.NewTableServiceClient(conn)

	// Create session
	createResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "acceptance-dashboard",
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := createResp.SessionId
	defer sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{SessionId: sessionID})

	// Create the root Flex layout
	flexResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FLEX,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Dashboard"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Flex) failed: %v", err)
	}
	flexID := flexResp.WidgetId

	// Create a List widget
	listResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_LIST,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Recent Activity"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (List) failed: %v", err)
	}
	listID := listResp.WidgetId

	// Create a Table widget
	tableResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TABLE,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Metrics"),
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (Table) failed: %v", err)
	}
	tableID := tableResp.WidgetId

	// Create a TextView widget
	text := "Welcome to the dashboard. All systems operational."
	textViewResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
		Properties: &pb.WidgetProperties{
			Border: testutil.BoolPtr(true),
			Title:  testutil.StrPtr("Status"),
			TypeProperties: &pb.WidgetProperties_TextView{
				TextView: &pb.TextViewProperties{
					Text: &text,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateWidget (TextView) failed: %v", err)
	}
	textViewID := textViewResp.WidgetId

	// Add all three widgets to the Flex layout
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     listID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (List) failed: %v", err)
	}

	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     tableID,
		Proportion: 2,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (Table) failed: %v", err)
	}

	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     flexID,
		ItemId:     textViewID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		t.Fatalf("FlexAddItem (TextView) failed: %v", err)
	}

	// Populate the List with items
	activities := []string{"User login", "Data import", "Report generated"}
	for i, activity := range activities {
		_, err := listClient.AddItem(ctx, &pb.ListAddItemRequest{
			SessionId:     sessionID,
			WidgetId:      listID,
			MainText:      activity,
			SecondaryText: "Just now",
		})
		if err != nil {
			t.Fatalf("AddItem (list %d) failed: %v", i, err)
		}
	}

	// Populate the Table with cells
	_, err = tableClient.SetCells(ctx, &pb.SetTableCellsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
		Cells: []*pb.TableCellUpdate{
			{Row: 0, Column: 0, Cell: &pb.TableCell{Text: "Metric"}},
			{Row: 0, Column: 1, Cell: &pb.TableCell{Text: "Value"}},
			{Row: 1, Column: 0, Cell: &pb.TableCell{Text: "CPU"}},
			{Row: 1, Column: 1, Cell: &pb.TableCell{Text: "45%"}},
			{Row: 2, Column: 0, Cell: &pb.TableCell{Text: "Memory"}},
			{Row: 2, Column: 1, Cell: &pb.TableCell{Text: "2.1 GB"}},
		},
	})
	if err != nil {
		t.Fatalf("SetCells (table) failed: %v", err)
	}

	// Set the Flex as root
	setRootResp, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
		SessionId: sessionID,
		WidgetId:  flexID,
	})
	if err != nil {
		t.Fatalf("SetRoot failed: %v", err)
	}
	if !setRootResp.Success {
		t.Error("Expected successful SetRoot")
	}

	// Verify all widgets exist via ListWidgets
	listWidgetsResp, err := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{
		SessionId: sessionID,
	})
	if err != nil {
		t.Fatalf("ListWidgets failed: %v", err)
	}

	// We created 4 widgets: Flex, List, Table, TextView
	if len(listWidgetsResp.Widgets) != 4 {
		t.Errorf("Expected 4 widgets in dashboard, got %d", len(listWidgetsResp.Widgets))
	}

	// Verify each widget ID is present
	widgetIDs := make(map[string]bool)
	for _, w := range listWidgetsResp.Widgets {
		widgetIDs[w.WidgetId.Id] = true
	}

	expectedIDs := []struct {
		id   string
		name string
	}{
		{flexID.Id, "Flex"},
		{listID.Id, "List"},
		{tableID.Id, "Table"},
		{textViewID.Id, "TextView"},
	}

	for _, expected := range expectedIDs {
		if !widgetIDs[expected.id] {
			t.Errorf("Widget %s (%s) not found in ListWidgets", expected.name, expected.id)
		}
	}

	// Verify list has items
	listCountResp, err := listClient.GetItemCount(ctx, &pb.ListGetItemCountRequest{
		SessionId: sessionID,
		WidgetId:  listID,
	})
	if err != nil {
		t.Fatalf("GetItemCount (list) failed: %v", err)
	}
	if listCountResp.Count != 3 {
		t.Errorf("Expected 3 list items, got %d", listCountResp.Count)
	}

	// Verify table has data
	tableDimsResp, err := tableClient.GetDimensions(ctx, &pb.TableGetDimensionsRequest{
		SessionId: sessionID,
		WidgetId:  tableID,
	})
	if err != nil {
		t.Fatalf("GetDimensions (table) failed: %v", err)
	}
	if tableDimsResp.Rows != 3 {
		t.Errorf("Expected 3 table rows, got %d", tableDimsResp.Rows)
	}
	if tableDimsResp.Columns != 2 {
		t.Errorf("Expected 2 table columns, got %d", tableDimsResp.Columns)
	}

	// Clean up is handled by the deferred DestroySession
}
