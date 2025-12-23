package client

import (
	"testing"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

func TestNamedColor(t *testing.T) {
	color := NamedColor("red")
	if color == nil {
		t.Fatal("Expected non-nil color")
	}

	name, ok := color.Color.(*pb.Color_Name)
	if !ok {
		t.Fatal("Expected Color_Name type")
	}
	if name.Name != "red" {
		t.Errorf("Expected 'red', got '%s'", name.Name)
	}
}

func TestIndexColor(t *testing.T) {
	color := IndexColor(42)
	if color == nil {
		t.Fatal("Expected non-nil color")
	}

	idx, ok := color.Color.(*pb.Color_Index)
	if !ok {
		t.Fatal("Expected Color_Index type")
	}
	if idx.Index != 42 {
		t.Errorf("Expected 42, got %d", idx.Index)
	}
}

func TestRGBColor(t *testing.T) {
	color := RGBColor(255, 128, 64)
	if color == nil {
		t.Fatal("Expected non-nil color")
	}

	rgb, ok := color.Color.(*pb.Color_Rgb)
	if !ok {
		t.Fatal("Expected Color_Rgb type")
	}
	if rgb.Rgb.R != 255 {
		t.Errorf("Expected R=255, got %d", rgb.Rgb.R)
	}
	if rgb.Rgb.G != 128 {
		t.Errorf("Expected G=128, got %d", rgb.Rgb.G)
	}
	if rgb.Rgb.B != 64 {
		t.Errorf("Expected B=64, got %d", rgb.Rgb.B)
	}
}

func TestHexColor(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expectedR int32
		expectedG int32
		expectedB int32
	}{
		{"with hash", "#FF8040", 255, 128, 64},
		{"without hash", "FF8040", 255, 128, 64},
		{"lowercase", "#ff8040", 255, 128, 64},
		{"short form", "#F84", 255, 136, 68},
		{"black", "#000000", 0, 0, 0},
		{"white", "#FFFFFF", 255, 255, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := HexColor(tt.hex)
			if color == nil {
				t.Fatal("Expected non-nil color")
			}

			rgb, ok := color.Color.(*pb.Color_Rgb)
			if !ok {
				t.Fatal("Expected Color_Rgb type")
			}
			if rgb.Rgb.R != tt.expectedR {
				t.Errorf("Expected R=%d, got %d", tt.expectedR, rgb.Rgb.R)
			}
			if rgb.Rgb.G != tt.expectedG {
				t.Errorf("Expected G=%d, got %d", tt.expectedG, rgb.Rgb.G)
			}
			if rgb.Rgb.B != tt.expectedB {
				t.Errorf("Expected B=%d, got %d", tt.expectedB, rgb.Rgb.B)
			}
		})
	}
}

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme
	if theme == nil {
		t.Fatal("DefaultTheme is nil")
	}

	if theme.PrimaryColor == nil {
		t.Error("PrimaryColor is nil")
	}
	if theme.BackgroundColor == nil {
		t.Error("BackgroundColor is nil")
	}
	if theme.TextColor == nil {
		t.Error("TextColor is nil")
	}
	if theme.SuccessColor == nil {
		t.Error("SuccessColor is nil")
	}
	if theme.ErrorColor == nil {
		t.Error("ErrorColor is nil")
	}
}

func TestAllThemesExist(t *testing.T) {
	themes := []*Theme{
		DefaultTheme,
		LightTheme,
		MonochromeTheme,
		OceanTheme,
		ForestTheme,
		SunsetTheme,
	}

	for i, theme := range themes {
		if theme == nil {
			t.Errorf("Theme %d is nil", i)
			continue
		}

		// Check all required colors are set
		if theme.PrimaryColor == nil {
			t.Errorf("Theme %d: PrimaryColor is nil", i)
		}
		if theme.BackgroundColor == nil {
			t.Errorf("Theme %d: BackgroundColor is nil", i)
		}
		if theme.TextColor == nil {
			t.Errorf("Theme %d: TextColor is nil", i)
		}
	}
}

func TestHexToInt(t *testing.T) {
	tests := []struct {
		hex      string
		expected int
	}{
		{"00", 0},
		{"FF", 255},
		{"ff", 255},
		{"80", 128},
		{"0A", 10},
		{"a0", 160},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			result := hexToInt(tt.hex)
			if result != tt.expected {
				t.Errorf("hexToInt(%s) = %d, want %d", tt.hex, result, tt.expected)
			}
		})
	}
}
