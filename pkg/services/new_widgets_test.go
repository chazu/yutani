package services

import (
	"context"
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/rivo/tview"
)

func setupTestWidgetService(t *testing.T) (*WidgetService, *SessionService, string) {
	t.Helper()

	srv := setupTestServer(t)

	widgetService := NewWidgetService(srv)
	sessionService := NewSessionService(srv)

	// Create a session
	resp, err := sessionService.CreateSession(context.Background(), &pb.CreateSessionRequest{
		ClientName: "test-client",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	return widgetService, sessionService, resp.SessionId.Id
}

func TestCreateTextArea(t *testing.T) {
	widgetService, _, sessionID := setupTestWidgetService(t)

	text := "Hello, World!"
	placeholder := "Enter text..."
	wordWrap := true

	resp, err := widgetService.CreateWidget(context.Background(), &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_TEXT_AREA,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_TextArea{
				TextArea: &pb.TextAreaProperties{
					Text:        &text,
					Placeholder: &placeholder,
					WordWrap:    &wordWrap,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Failed to create text area: %v", err)
	}
	if resp.WidgetId == nil {
		t.Fatal("Widget ID is nil")
	}
	if !resp.Success {
		t.Fatal("Expected success to be true")
	}
}

func TestCreateModal(t *testing.T) {
	widgetService, _, sessionID := setupTestWidgetService(t)

	text := "Are you sure?"
	buttons := []string{"Yes", "No", "Cancel"}

	resp, err := widgetService.CreateWidget(context.Background(), &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_MODAL,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Modal{
				Modal: &pb.ModalProperties{
					Text:    &text,
					Buttons: buttons,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Failed to create modal: %v", err)
	}
	if resp.WidgetId == nil {
		t.Fatal("Widget ID is nil")
	}
	if !resp.Success {
		t.Fatal("Expected success to be true")
	}
}

func TestCreateImage(t *testing.T) {
	widgetService, _, sessionID := setupTestWidgetService(t)

	colors := int32(256)
	dithering := true

	resp, err := widgetService.CreateWidget(context.Background(), &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_IMAGE,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_Image{
				Image: &pb.ImageProperties{
					Colors:    &colors,
					Dithering: &dithering,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Failed to create image: %v", err)
	}
	if resp.WidgetId == nil {
		t.Fatal("Widget ID is nil")
	}
	if !resp.Success {
		t.Fatal("Expected success to be true")
	}
}

func TestCreateProgressBar(t *testing.T) {
	widgetService, _, sessionID := setupTestWidgetService(t)

	progress := int32(50)
	max := int32(100)
	label := "Loading..."
	showPercentage := true

	resp, err := widgetService.CreateWidget(context.Background(), &pb.CreateWidgetRequest{
		SessionId: &pb.SessionId{Id: sessionID},
		Type:      pb.WidgetType_WIDGET_PROGRESS_BAR,
		Properties: &pb.WidgetProperties{
			TypeProperties: &pb.WidgetProperties_ProgressBar{
				ProgressBar: &pb.ProgressBarProperties{
					Progress:       &progress,
					Max:            &max,
					Label:          &label,
					ShowPercentage: &showPercentage,
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Failed to create progress bar: %v", err)
	}
	if resp.WidgetId == nil {
		t.Fatal("Widget ID is nil")
	}
	if !resp.Success {
		t.Fatal("Expected success to be true")
	}
}

func TestProgressBarUpdate(t *testing.T) {
	pb := newProgressBar()
	pb.SetMax(100)
	pb.SetProgress(50)
	pb.SetLabel("Test")
	pb.SetShowPercentage(true)

	// Verify internal state
	if pb.progress != 50 {
		t.Errorf("Expected progress 50, got %d", pb.progress)
	}
	if pb.max != 100 {
		t.Errorf("Expected max 100, got %d", pb.max)
	}
	if pb.label != "Test" {
		t.Errorf("Expected label 'Test', got '%s'", pb.label)
	}
	if !pb.showPercentage {
		t.Error("Expected showPercentage to be true")
	}

	// Test bounds
	pb.SetProgress(-10)
	if pb.progress != 0 {
		t.Errorf("Expected progress clamped to 0, got %d", pb.progress)
	}

	pb.SetProgress(150)
	if pb.progress != 100 {
		t.Errorf("Expected progress clamped to max, got %d", pb.progress)
	}
}

func TestProgressBarRender(t *testing.T) {
	pb := newProgressBar()
	pb.SetMax(100)
	pb.SetProgress(50)
	pb.SetShowPercentage(true)

	// The text should contain percentage
	text := pb.TextView.GetText(true)
	if text == "" {
		t.Error("Expected non-empty text")
	}
}

func TestTextAreaProperties(t *testing.T) {
	srv := setupTestServer(t)
	widgetService := NewWidgetService(srv)

	// Test createTextArea
	text := "Test content"
	placeholder := "Enter here"
	maxLength := int32(1000)
	wordWrap := true

	props := &pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_TextArea{
			TextArea: &pb.TextAreaProperties{
				Text:        &text,
				Placeholder: &placeholder,
				MaxLength:   &maxLength,
				WordWrap:    &wordWrap,
			},
		},
	}

	ta := widgetService.createTextArea(props)
	if ta == nil {
		t.Fatal("Expected non-nil TextArea")
	}

	// Verify it's a valid tview primitive
	var _ tview.Primitive = ta
}

func TestModalProperties(t *testing.T) {
	srv := setupTestServer(t)
	widgetService := NewWidgetService(srv)

	text := "Confirm action?"
	buttons := []string{"OK", "Cancel"}

	props := &pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Modal{
			Modal: &pb.ModalProperties{
				Text:    &text,
				Buttons: buttons,
			},
		},
	}

	modal := widgetService.createModal(props)
	if modal == nil {
		t.Fatal("Expected non-nil Modal")
	}

	// Verify it's a valid tview primitive
	var _ tview.Primitive = modal
}

func TestImageProperties(t *testing.T) {
	srv := setupTestServer(t)
	widgetService := NewWidgetService(srv)

	colors := int32(256)
	dithering := true

	props := &pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{
				Colors:    &colors,
				Dithering: &dithering,
			},
		},
	}

	img := widgetService.createImage(props)
	if img == nil {
		t.Fatal("Expected non-nil Image")
	}

	// Verify it's a valid tview primitive
	var _ tview.Primitive = img
}

func TestColorToString(t *testing.T) {
	tests := []struct {
		name     string
		color    *pb.Color
		expected string
	}{
		{
			name:     "nil color",
			color:    nil,
			expected: "white",
		},
		{
			name: "named color",
			color: &pb.Color{
				Color: &pb.Color_Name{Name: "red"},
			},
			expected: "red",
		},
		{
			name: "index color",
			color: &pb.Color{
				Color: &pb.Color_Index{Index: 42},
			},
			expected: "color42",
		},
		{
			name: "RGB color",
			color: &pb.Color{
				Color: &pb.Color_Rgb{
					Rgb: &pb.RGB{R: 255, G: 128, B: 0},
				},
			},
			expected: "#ff8000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorToString(tt.color)
			if result != tt.expected {
				t.Errorf("colorToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}
