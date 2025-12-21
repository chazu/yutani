package main

import (
	"context"
	"log"
	"time"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:7755", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create service clients
	sessionClient := pb.NewSessionServiceClient(conn)
	widgetClient := pb.NewWidgetServiceClient(conn)
	listClient := pb.NewListServiceClient(conn)
	tableClient := pb.NewTableServiceClient(conn)
	formClient := pb.NewFormServiceClient(conn)
	treeClient := pb.NewTreeServiceClient(conn)
	layoutClient := pb.NewLayoutServiceClient(conn)

	ctx := context.Background()

	// Create session
	log.Println("Creating session...")
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "phase4-demo",
		Preferences: &pb.SessionPreferences{
			ReceiveMouseEvents:  true,
			ReceiveKeyEvents:    true,
			ReceiveResizeEvents: true,
		},
	})
	if err != nil {
		log.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := &pb.SessionId{Id: sessionResp.SessionId.Id}
	log.Printf("✓ Session created: %s\n", sessionID.Id)

	// Create main layout container (Flex with column direction)
	log.Println("\n=== Creating Main Layout ===")
	flexDir := pb.FlexDirection_FLEX_COLUMN
	mainFlexResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FLEX,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Flex{
				Flex: &pb.FlexProperties{
					Direction: &flexDir,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create main flex: %v", err)
	}
	mainFlexID := mainFlexResp.WidgetId
	log.Printf("✓ Main flex container created: %s\n", mainFlexID.Id)

	// Set as root
	_, err = widgetClient.SetRoot(ctx, &pb.SetRootRequest{
		SessionId: sessionID,
		WidgetId:  mainFlexID,
	})
	if err != nil {
		log.Fatalf("Failed to set root: %v", err)
	}
	log.Println("✓ Set as root widget")

	// Create title text
	titleText := "[yellow::b]Yutani Phase 4 Demo - Complex Widgets[-::-]\n"
	dynamicColors := true
	titleResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_TextView{
				TextView: &pb.TextViewProperties{
					Text:          &titleText,
					DynamicColors: &dynamicColors,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create title: %v", err)
	}
	log.Println("✓ Title created")

	// Add title to main flex (fixed size 1 line)
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     mainFlexID,
		ItemId:     titleResp.WidgetId,
		FixedSize:  1,
		Focus:      false,
	})
	if err != nil {
		log.Fatalf("Failed to add title to flex: %v", err)
	}
	log.Println("✓ Title added to layout")

	// Create horizontal flex for top row (List and Table)
	flexRowDir := pb.FlexDirection_FLEX_ROW
	topRowResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FLEX,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Flex{
				Flex: &pb.FlexProperties{
					Direction: &flexRowDir,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create top row flex: %v", err)
	}
	topRowID := topRowResp.WidgetId
	log.Println("✓ Top row flex created")

	// Add top row to main flex (proportion 1)
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     mainFlexID,
		ItemId:     topRowID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		log.Fatalf("Failed to add top row to main flex: %v", err)
	}
	log.Println("✓ Top row added to main layout")

	// Demo List widget
	log.Println("\n=== Creating List Widget ===")
	if err := demoList(ctx, sessionID, widgetClient, listClient, layoutClient, topRowID); err != nil {
		log.Fatalf("List demo failed: %v", err)
	}



	// Demo Table widget
	log.Println("\n=== Creating Table Widget ===")
	if err := demoTable(ctx, sessionID, widgetClient, tableClient, layoutClient, topRowID); err != nil {
		log.Fatalf("Table demo failed: %v", err)
	}

	// Create bottom row flex for Form and Tree
	bottomRowResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FLEX,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Flex{
				Flex: &pb.FlexProperties{
					Direction: &flexRowDir,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create bottom row flex: %v", err)
	}
	bottomRowID := bottomRowResp.WidgetId
	log.Println("✓ Bottom row flex created")

	// Add bottom row to main flex (proportion 1)
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     mainFlexID,
		ItemId:     bottomRowID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		log.Fatalf("Failed to add bottom row to main flex: %v", err)
	}
	log.Println("✓ Bottom row added to main layout")

	// Demo Form widget
	log.Println("\n=== Creating Form Widget ===")
	if err := demoForm(ctx, sessionID, widgetClient, formClient, layoutClient, bottomRowID); err != nil {
		log.Fatalf("Form demo failed: %v", err)
	}

	// Demo Tree widget
	log.Println("\n=== Creating Tree Widget ===")
	if err := demoTree(ctx, sessionID, widgetClient, treeClient, layoutClient, bottomRowID); err != nil {
		log.Fatalf("Tree demo failed: %v", err)
	}

	log.Println("\n=== Demo Complete ===")
	log.Println("✓ All Phase 4 widgets created and displayed!")
	log.Println("✓ Layout hierarchy:")
	log.Println("  - Main Flex (column)")
	log.Println("    - Title (TextView)")
	log.Println("    - Top Row Flex (row)")
	log.Println("      - List Widget")
	log.Println("      - Table Widget")
	log.Println("    - Bottom Row Flex (row)")
	log.Println("      - Form Widget")
	log.Println("      - Tree Widget")
	log.Println("\nPress Ctrl+C to exit...")

	// Keep the demo running
	time.Sleep(60 * time.Second)
}

func demoList(ctx context.Context, sessionID *pb.SessionId, widgetClient pb.WidgetServiceClient,
	listClient pb.ListServiceClient, layoutClient pb.LayoutServiceClient, containerID *pb.WidgetId) error {

	// Create list widget with border
	border := true
	title := "Task List"
	showSecondary := true
	listResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_LIST,
		Properties: &pb.WidgetProperties{
			Border: &border,
			Title:  &title,
			TypeProperties: &pb.WidgetProperties_List{
				List: &pb.ListProperties{
					ShowSecondaryText: &showSecondary,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	listID := listResp.WidgetId
	log.Println("✓ List widget created")

	// Add list to container
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     containerID,
		ItemId:     listID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		return err
	}
	log.Println("✓ List added to layout")

	// Add items to list
	items := []struct {
		main, secondary string
	}{
		{"✓ Phase 1: Core Infrastructure", "Sessions, Events, Screen"},
		{"✓ Phase 2: Widget System", "Factory, Registry, Lifecycle"},
		{"✓ Phase 3: Basic Widgets", "Box, TextView, Input, Button"},
		{"✓ Phase 4: Complex Widgets", "List, Table, Form, Tree, Layout"},
		{"⏳ Phase 5: Advanced Features", "Menus, Modals, Tabs"},
	}

	for _, item := range items {
		_, err = listClient.AddItem(ctx, &pb.AddItemRequest{
			SessionId: sessionID,
			WidgetId:  listID,
			MainText:  item.main,
			SecondaryText: item.secondary,
		})
		if err != nil {
			return err
		}
	}
	log.Printf("✓ Added %d items to list\n", len(items))

	return nil
}

func demoTable(ctx context.Context, sessionID *pb.SessionId, widgetClient pb.WidgetServiceClient,
	tableClient pb.TableServiceClient, layoutClient pb.LayoutServiceClient, containerID *pb.WidgetId) error {

	// Create table widget with border
	tableBorder := true
	tableTitle := "Service Stats"
	tableBorders := true
	tableSelectable := true
	tableResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TABLE,
		Properties: &pb.WidgetProperties{
			Border: &tableBorder,
			Title:  &tableTitle,
			TypeProperties: &pb.WidgetProperties_Table{
				Table: &pb.TableProperties{
					Borders:    &tableBorders,
					Selectable: &tableSelectable,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	tableID := tableResp.WidgetId
	log.Println("✓ Table widget created")

	// Add table to container
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     containerID,
		ItemId:     tableID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		return err
	}
	log.Println("✓ Table added to layout")

	// Set header row
	headers := []string{"Service", "RPCs", "Status"}
	for col, header := range headers {
		align := pb.Alignment_ALIGN_CENTER
		_, err = tableClient.SetCell(ctx, &pb.SetTableCellRequest{
			SessionId: sessionID,
			WidgetId:  tableID,
			Row:       0,
			Column:    int32(col),
			Cell: &pb.TableCell{
				Text: header,
				Color: &pb.Color{
					Color: &pb.Color_Name{Name: "yellow"},
				},
				Align: &align,
			},
		})
		if err != nil {
			return err
		}
	}

	// Set fixed header row
	_, err = tableClient.SetFixed(ctx, &pb.SetFixedRequest{
		SessionId:   sessionID,
		WidgetId:    tableID,
		FixedRows:   1,
		FixedColumns: 0,
	})
	if err != nil {
		return err
	}

	// Add data rows
	data := [][]string{
		{"ListService", "5", "✓"},
		{"TableService", "8", "✓"},
		{"FormService", "6", "✓"},
		{"TreeService", "7", "✓"},
		{"LayoutService", "11", "✓"},
	}

	for row, rowData := range data {
		for col, cell := range rowData {
			_, err = tableClient.SetCell(ctx, &pb.SetTableCellRequest{
				SessionId: sessionID,
				WidgetId:  tableID,
				Row:       int32(row + 1),
				Column:    int32(col),
				Cell: &pb.TableCell{
					Text: cell,
				},
			})
			if err != nil {
				return err
			}
		}
	}
	log.Printf("✓ Added %d rows to table\n", len(data))

	return nil
}

func demoForm(ctx context.Context, sessionID *pb.SessionId, widgetClient pb.WidgetServiceClient,
	formClient pb.FormServiceClient, layoutClient pb.LayoutServiceClient, containerID *pb.WidgetId) error {

	// Create form widget with border
	formBorder := true
	formTitle := "User Settings"
	formHorizontal := false
	formResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_FORM,
		Properties: &pb.WidgetProperties{
			Border: &formBorder,
			Title:  &formTitle,
			TypeProperties: &pb.WidgetProperties_Form{
				Form: &pb.FormProperties{
					Horizontal: &formHorizontal,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	formID := formResp.WidgetId
	log.Println("✓ Form widget created")

	// Add form to container
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     containerID,
		ItemId:     formID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		return err
	}
	log.Println("✓ Form added to layout")

	// Add input field
	_, err = formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Username:",
		FieldType: pb.FormFieldType_FORM_FIELD_INPUT,
	})
	if err != nil {
		return err
	}

	// Add password field
	_, err = formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Password:",
		FieldType: pb.FormFieldType_FORM_FIELD_PASSWORD,
	})
	if err != nil {
		return err
	}

	// Add checkbox
	_, err = formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Remember me:",
		FieldType: pb.FormFieldType_FORM_FIELD_CHECKBOX,
	})
	if err != nil {
		return err
	}

	// Add dropdown
	_, err = formClient.AddField(ctx, &pb.AddFieldRequest{
		SessionId:       sessionID,
		WidgetId:        formID,
		Label:           "Theme:",
		FieldType:       pb.FormFieldType_FORM_FIELD_DROPDOWN,
		DropdownOptions: []string{"Dark", "Light", "Auto"},
	})
	if err != nil {
		return err
	}

	// Add buttons
	_, err = formClient.AddButton(ctx, &pb.AddButtonRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Submit",
	})
	if err != nil {
		return err
	}

	_, err = formClient.AddButton(ctx, &pb.AddButtonRequest{
		SessionId: sessionID,
		WidgetId:  formID,
		Label:     "Cancel",
	})
	if err != nil {
		return err
	}

	log.Println("✓ Form fields and buttons added")

	return nil
}

func demoTree(ctx context.Context, sessionID *pb.SessionId, widgetClient pb.WidgetServiceClient,
	treeClient pb.TreeServiceClient, layoutClient pb.LayoutServiceClient, containerID *pb.WidgetId) error {

	// Create tree widget with border
	treeBorder := true
	treeTitle := "Project Structure"
	treeResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: sessionID,
		Type:      pb.WidgetType_WIDGET_TREE_VIEW,
		Properties: &pb.WidgetProperties{
			Border: &treeBorder,
			Title:  &treeTitle,
		},
	})
	if err != nil {
		return err
	}
	treeID := treeResp.WidgetId
	log.Println("✓ Tree widget created")

	// Add tree to container
	_, err = layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
		SessionId:  sessionID,
		FlexId:     containerID,
		ItemId:     treeID,
		Proportion: 1,
		Focus:      false,
	})
	if err != nil {
		return err
	}
	log.Println("✓ Tree added to layout")

	// Set root node
	rootResp, err := treeClient.SetRoot(ctx, &pb.SetTreeRootRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		Node: &pb.TreeNode{
			Text: "yutani/",
		},
	})
	if err != nil {
		return err
	}
	rootID := rootResp.NodeId

	// Add child nodes
	pkgResp, err := treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  rootID,
		Node: &pb.TreeNode{
			Text: "pkg/",
		},
	})
	if err != nil {
		return err
	}
	pkgID := pkgResp.NodeId

	// Add grandchildren
	_, err = treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  pkgID,
		Node: &pb.TreeNode{
			Text: "server/",
		},
	})
	if err != nil {
		return err
	}

	_, err = treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  pkgID,
		Node: &pb.TreeNode{
			Text: "services/",
		},
	})
	if err != nil {
		return err
	}

	_, err = treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  pkgID,
		Node: &pb.TreeNode{
			Text: "proto/",
		},
	})
	if err != nil {
		return err
	}

	// Add cmd node
	_, err = treeClient.AddChild(ctx, &pb.AddTreeChildRequest{
		SessionId: sessionID,
		WidgetId:  treeID,
		ParentId:  rootID,
		Node: &pb.TreeNode{
			Text: "cmd/",
		},
	})
	if err != nil {
		return err
	}

	log.Println("✓ Tree nodes added")

	return nil
}

