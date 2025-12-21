package services

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestConvertColor(t *testing.T) {
	tests := []struct {
		name     string
		pbColor  *pb.Color
		expected tcell.Color
	}{
		{
			name:     "nil color",
			pbColor:  nil,
			expected: tcell.ColorDefault,
		},
		{
			name: "named color",
			pbColor: &pb.Color{
				Color: &pb.Color_Name{Name: "red"},
			},
			expected: tcell.GetColor("red"),
		},
		{
			name: "indexed color",
			pbColor: &pb.Color{
				Color: &pb.Color_Index{Index: 5},
			},
			expected: tcell.PaletteColor(5),
		},
		{
			name: "RGB color",
			pbColor: &pb.Color{
				Color: &pb.Color_Rgb{
					Rgb: &pb.RGB{R: 255, G: 128, B: 64},
				},
			},
			expected: tcell.NewRGBColor(255, 128, 64),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertColor(tt.pbColor)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestApplyAttribute(t *testing.T) {
	tests := []struct {
		name     string
		attr     pb.Attribute
		hasAttr  func(tcell.Style) bool
	}{
		{
			name: "bold",
			attr: pb.Attribute_ATTR_BOLD,
			hasAttr: func(s tcell.Style) bool {
				_, _, attrs := s.Decompose()
				return attrs&tcell.AttrBold != 0
			},
		},
		{
			name: "italic",
			attr: pb.Attribute_ATTR_ITALIC,
			hasAttr: func(s tcell.Style) bool {
				_, _, attrs := s.Decompose()
				return attrs&tcell.AttrItalic != 0
			},
		},
		{
			name: "underline",
			attr: pb.Attribute_ATTR_UNDERLINE,
			hasAttr: func(s tcell.Style) bool {
				_, _, attrs := s.Decompose()
				return attrs&tcell.AttrUnderline != 0
			},
		},
		{
			name: "reverse",
			attr: pb.Attribute_ATTR_REVERSE,
			hasAttr: func(s tcell.Style) bool {
				_, _, attrs := s.Decompose()
				return attrs&tcell.AttrReverse != 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tcell.StyleDefault
			result := applyAttribute(style, tt.attr)
			if !tt.hasAttr(result) {
				t.Errorf("Attribute %v not applied", tt.attr)
			}
		})
	}
}

func TestConvertStyleToProto(t *testing.T) {
	// Create a style with foreground, background, and attributes
	style := tcell.StyleDefault.
		Foreground(tcell.ColorRed).
		Background(tcell.ColorBlue).
		Bold(true).
		Italic(true)

	pbStyle := convertStyleToProto(style)

	if pbStyle == nil {
		t.Fatal("Expected style, got nil")
	}

	// Check that attributes are present
	hasBold := false
	hasItalic := false
	for _, attr := range pbStyle.Attributes {
		if attr == pb.Attribute_ATTR_BOLD {
			hasBold = true
		}
		if attr == pb.Attribute_ATTR_ITALIC {
			hasItalic = true
		}
	}

	if !hasBold {
		t.Error("Expected BOLD attribute")
	}
	if !hasItalic {
		t.Error("Expected ITALIC attribute")
	}
}

func TestConvertColorToProto(t *testing.T) {
	tests := []struct {
		name     string
		color    tcell.Color
		checkRGB bool
	}{
		{
			name:     "default color",
			color:    tcell.ColorDefault,
			checkRGB: false,
		},
		{
			name:     "RGB color",
			color:    tcell.NewRGBColor(255, 128, 64),
			checkRGB: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertColorToProto(tt.color)
			if tt.checkRGB {
				if result == nil {
					t.Error("Expected color, got nil")
				}
				if rgb, ok := result.Color.(*pb.Color_Rgb); ok {
					if rgb.Rgb.R != 255 || rgb.Rgb.G != 128 || rgb.Rgb.B != 64 {
						t.Errorf("RGB values incorrect: %v", rgb.Rgb)
					}
				} else {
					t.Error("Expected RGB color")
				}
			} else {
				if result != nil {
					t.Error("Expected nil for default color")
				}
			}
		})
	}
}

