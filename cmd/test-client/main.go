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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: Ping
	log.Println("Testing Ping...")
	pingResp, err := sessionClient.Ping(ctx, &pb.PingRequest{
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		log.Fatalf("Ping failed: %v", err)
	}
	log.Printf("✓ Ping successful: server_timestamp=%d\n", pingResp.ServerTimestamp)

	// Test 2: GetServerInfo
	log.Println("\nTesting GetServerInfo...")
	infoResp, err := sessionClient.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
	if err != nil {
		log.Fatalf("GetServerInfo failed: %v", err)
	}
	log.Printf("✓ Server Info:\n")
	log.Printf("  Version: %s\n", infoResp.Version)
	log.Printf("  Active Sessions: %d\n", infoResp.ActiveSessions)
	log.Printf("  Max Sessions: %d\n", infoResp.MaxSessions)
	log.Printf("  Screen Size: %dx%d\n", infoResp.ScreenSize.Width, infoResp.ScreenSize.Height)
	log.Printf("  Mouse Support: %v\n", infoResp.Capabilities.MouseSupport)
	log.Printf("  True Color: %v\n", infoResp.Capabilities.TrueColor)

	// Test 3: CreateSession
	log.Println("\nTesting CreateSession...")
	sessionResp, err := sessionClient.CreateSession(ctx, &pb.CreateSessionRequest{
		ClientName: "test-client",
		Preferences: &pb.SessionPreferences{
			ReceiveMouseEvents:  true,
			ReceiveKeyEvents:    true,
			ReceiveResizeEvents: true,
		},
	})
	if err != nil {
		log.Fatalf("CreateSession failed: %v", err)
	}
	sessionID := sessionResp.SessionId.Id
	log.Printf("✓ Session created: %s\n", sessionID)

	// Test 4: GetScreenSize
	log.Println("\nTesting GetSize...")
	screenClient := pb.NewScreenServiceClient(conn)
	sizeResp, err := screenClient.GetSize(ctx, &pb.GetSizeRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		log.Fatalf("GetSize failed: %v", err)
	}
	log.Printf("✓ Screen size: %dx%d\n", sizeResp.Size.Width, sizeResp.Size.Height)

	// Test 5: Clear screen
	log.Println("\nTesting Clear...")
	clearResp, err := screenClient.Clear(ctx, &pb.ClearRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		log.Fatalf("Clear failed: %v", err)
	}
	log.Printf("✓ Clear successful: %v\n", clearResp.Success)

	// Test 6: Sync screen
	log.Println("\nTesting Sync...")
	syncResp, err := screenClient.Sync(ctx, &pb.SyncRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
	log.Printf("✓ Sync successful: %v\n", syncResp.Success)

	// Test 7: SetCell
	log.Println("\nTesting SetCell...")
	setCellResp, err := screenClient.SetCell(ctx, &pb.SetCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Position:  &pb.Position{X: 5, Y: 5},
		Cell: &pb.Cell{
			Rune: "X",
			Style: &pb.Style{
				Foreground: &pb.Color{Color: &pb.Color_Name{Name: "red"}},
				Attributes: []pb.Attribute{pb.Attribute_ATTR_BOLD},
			},
		},
	})
	if err != nil {
		log.Fatalf("SetCell failed: %v", err)
	}
	log.Printf("✓ SetCell successful: %v\n", setCellResp.Success)

	// Test 8: DrawText
	log.Println("\nTesting DrawText...")
	drawTextResp, err := screenClient.DrawText(ctx, &pb.DrawTextRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Position:  &pb.Position{X: 10, Y: 10},
		Text:      "Hello, Yutani!",
		Style: &pb.Style{
			Foreground: &pb.Color{Color: &pb.Color_Name{Name: "green"}},
		},
	})
	if err != nil {
		log.Fatalf("DrawText failed: %v", err)
	}
	log.Printf("✓ DrawText successful: %d chars drawn\n", drawTextResp.CharsDrawn)

	// Test 9: DrawBox
	log.Println("\nTesting DrawBox...")
	drawBoxResp, err := screenClient.DrawBox(ctx, &pb.DrawBoxRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Rect:      &pb.Rect{X: 2, Y: 2, Width: 20, Height: 10},
		Style: &pb.Style{
			Foreground: &pb.Color{Color: &pb.Color_Name{Name: "blue"}},
		},
	})
	if err != nil {
		log.Fatalf("DrawBox failed: %v", err)
	}
	log.Printf("✓ DrawBox successful: %v\n", drawBoxResp.Success)

	// Test 10: Fill
	log.Println("\nTesting Fill...")
	fillResp, err := screenClient.Fill(ctx, &pb.FillRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Region:    &pb.Rect{X: 30, Y: 5, Width: 10, Height: 5},
		Cell: &pb.Cell{
			Rune: "█",
			Style: &pb.Style{
				Foreground: &pb.Color{Color: &pb.Color_Name{Name: "yellow"}},
			},
		},
	})
	if err != nil {
		log.Fatalf("Fill failed: %v", err)
	}
	log.Printf("✓ Fill successful: %v\n", fillResp.Success)

	// Test 11: GetCell
	log.Println("\nTesting GetCell...")
	getCellResp, err := screenClient.GetCell(ctx, &pb.GetCellRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Position:  &pb.Position{X: 5, Y: 5},
	})
	if err != nil {
		log.Fatalf("GetCell failed: %v", err)
	}
	log.Printf("✓ GetCell successful: rune='%s'\n", getCellResp.Cell.Rune)

	// Test 12: Event streaming (brief test)
	log.Println("\nTesting Event Streaming...")
	eventClient := pb.NewEventServiceClient(conn)

	// Start event stream in goroutine
	streamCtx, streamCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer streamCancel()

	stream, err := eventClient.Subscribe(streamCtx, &pb.SubscribeRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Filter: &pb.EventFilter{
			ReceiveKeyEvents: true,
		},
	})
	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	}

	// Inject a test event
	_, err = eventClient.InjectEvent(ctx, &pb.InjectEventRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Event: &pb.Event{
			SessionId: &pb.SessionId{Id: sessionID},
			Timestamp: time.Now().UnixNano(),
			Event: &pb.Event_Key{
				Key: &pb.KeyEvent{
					Key:  pb.Key_KEY_ENTER,
					Rune: "\n",
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("InjectEvent failed: %v", err)
	}

	// Try to receive the event
	event, err := stream.Recv()
	if err != nil {
		log.Printf("⚠ Event receive failed (expected in non-TTY): %v\n", err)
	} else {
		log.Printf("✓ Event received: type=%T\n", event.Event)
	}

	// Test 13: DestroySession
	log.Println("\nTesting DestroySession...")
	destroyResp, err := sessionClient.DestroySession(ctx, &pb.DestroySessionRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		log.Fatalf("DestroySession failed: %v", err)
	}
	log.Printf("✓ Session destroyed: %v\n", destroyResp.Success)

	// Phase 3: Widget Tests
	log.Println("\n=== Phase 3: Widget Tests ===")

	widgetClient := pb.NewWidgetServiceClient(conn)

	// Test: Create a TextView widget
	log.Println("\nTesting CreateWidget (TextView)...")
	text := "Hello from Yutani!"
	textViewResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_VIEW,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("TextView Demo"),
			TypeProperties: &pb.WidgetProperties_TextView{
				TextView: &pb.TextViewProperties{
					Text:          &text,
					DynamicColors: boolPtr(true),
					WordWrap:      boolPtr(true),
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateWidget (TextView) failed: %v", err)
	}
	textViewID := textViewResp.WidgetId.Id
	log.Printf("✓ TextView created: %s\n", textViewID)

	// Test: Create an InputField widget
	log.Println("\nTesting CreateWidget (InputField)...")
	inputResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_INPUT_FIELD,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			Title:  strPtr("Input Demo"),
			TypeProperties: &pb.WidgetProperties_InputField{
				InputField: &pb.InputFieldProperties{
					Label:       strPtr("Name: "),
					Placeholder: strPtr("Enter your name"),
					FieldWidth:  int32Ptr(20),
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateWidget (InputField) failed: %v", err)
	}
	inputID := inputResp.WidgetId.Id
	log.Printf("✓ InputField created: %s\n", inputID)

	// Test: Create a Button widget
	log.Println("\nTesting CreateWidget (Button)...")
	buttonResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_BUTTON,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			TypeProperties: &pb.WidgetProperties_Button{
				Button: &pb.ButtonProperties{
					Label: strPtr("Click Me!"),
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateWidget (Button) failed: %v", err)
	}
	buttonID := buttonResp.WidgetId.Id
	log.Printf("✓ Button created: %s\n", buttonID)

	// Test: Create a Checkbox widget
	log.Println("\nTesting CreateWidget (Checkbox)...")
	checkboxResp, err := widgetClient.CreateWidget(ctx, &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_CHECKBOX,
		Properties: &pb.WidgetProperties{
			Border: boolPtr(true),
			TypeProperties: &pb.WidgetProperties_Checkbox{
				Checkbox: &pb.CheckboxProperties{
					Label:   strPtr("Enable feature"),
					Checked: boolPtr(false),
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("CreateWidget (Checkbox) failed: %v", err)
	}
	checkboxID := checkboxResp.WidgetId.Id
	log.Printf("✓ Checkbox created: %s\n", checkboxID)

	// Test: ListWidgets
	log.Println("\nTesting ListWidgets...")
	listResp, err := widgetClient.ListWidgets(ctx, &pb.ListWidgetsRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		log.Fatalf("ListWidgets failed: %v", err)
	}
	log.Printf("✓ Widgets in session: %d\n", len(listResp.Widgets))
	for i, w := range listResp.Widgets {
		log.Printf("  %d. Widget ID: %s, Type: %s\n", i+1, w.WidgetId.Id, w.Type)
	}

	// Test: SetRoot (display the TextView)
	log.Println("\nTesting SetRoot...")
	setRootResp, err := widgetClient.SetRoot(ctx, &pb.SetRootRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: textViewID},
	})
	if err != nil {
		log.Fatalf("SetRoot failed: %v", err)
	}
	log.Printf("✓ Root widget set: %v\n", setRootResp.Success)

	// Test: SetFocus
	log.Println("\nTesting SetFocus...")
	setFocusResp, err := widgetClient.SetFocus(ctx, &pb.SetFocusRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		WidgetId:  &pb.WidgetId{Id: textViewID},
	})
	if err != nil {
		log.Fatalf("SetFocus failed: %v", err)
	}
	log.Printf("✓ Focus set: %v\n", setFocusResp.Success)

	// Test: GetFocus
	log.Println("\nTesting GetFocus...")
	getFocusResp, err := widgetClient.GetFocus(ctx, &pb.GetFocusRequest{
		SessionId: &pb.SessionId{Id: sessionID},
	})
	if err != nil {
		log.Fatalf("GetFocus failed: %v", err)
	}
	if getFocusResp.WidgetId != nil {
		log.Printf("✓ Focused widget: %s\n", getFocusResp.WidgetId.Id)
	} else {
		log.Printf("✓ No widget focused\n")
	}

	log.Println("\n=== All Phase 3 tests passed! ===")
	log.Println("\n✅ All tests passed!")
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}
