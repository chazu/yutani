package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Theme defines a consistent color scheme for widgets.
type Theme struct {
	// Primary colors
	PrimaryColor      *pb.Color
	SecondaryColor    *pb.Color
	BackgroundColor   *pb.Color
	SurfaceColor      *pb.Color

	// Text colors
	TextColor         *pb.Color
	TextSecondaryColor *pb.Color
	TextDisabledColor *pb.Color

	// Interactive element colors
	ButtonColor       *pb.Color
	ButtonTextColor   *pb.Color
	InputBackground   *pb.Color
	InputText         *pb.Color

	// Status colors
	SuccessColor      *pb.Color
	WarningColor      *pb.Color
	ErrorColor        *pb.Color
	InfoColor         *pb.Color

	// Border and divider colors
	BorderColor       *pb.Color
	DividerColor      *pb.Color

	// Selection colors
	SelectionColor    *pb.Color
	HighlightColor    *pb.Color
}

// Color helpers for creating pb.Color instances

// NamedColor creates a color from a named color string.
func NamedColor(name string) *pb.Color {
	return &pb.Color{
		Color: &pb.Color_Name{Name: name},
	}
}

// IndexColor creates a color from a 256-color palette index.
func IndexColor(index int) *pb.Color {
	return &pb.Color{
		Color: &pb.Color_Index{Index: int32(index)},
	}
}

// RGBColor creates a true color from RGB values.
func RGBColor(r, g, b int) *pb.Color {
	return &pb.Color{
		Color: &pb.Color_Rgb{
			Rgb: &pb.RGB{
				R: int32(r),
				G: int32(g),
				B: int32(b),
			},
		},
	}
}

// HexColor creates a color from a hex string (e.g., "#FF0000" or "FF0000").
func HexColor(hex string) *pb.Color {
	// Remove # prefix if present
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse hex values
	var r, g, b int
	if len(hex) == 6 {
		r = hexToInt(hex[0:2])
		g = hexToInt(hex[2:4])
		b = hexToInt(hex[4:6])
	} else if len(hex) == 3 {
		// Short form: #RGB -> #RRGGBB
		r = hexToInt(string(hex[0]) + string(hex[0]))
		g = hexToInt(string(hex[1]) + string(hex[1]))
		b = hexToInt(string(hex[2]) + string(hex[2]))
	}

	return RGBColor(r, g, b)
}

func hexToInt(hex string) int {
	val := 0
	for _, c := range hex {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += int(c - '0')
		case c >= 'a' && c <= 'f':
			val += int(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			val += int(c - 'A' + 10)
		}
	}
	return val
}

// Predefined themes

// DefaultTheme provides a standard dark theme.
var DefaultTheme = &Theme{
	PrimaryColor:       NamedColor("blue"),
	SecondaryColor:     NamedColor("cyan"),
	BackgroundColor:    NamedColor("black"),
	SurfaceColor:       IndexColor(235),

	TextColor:          NamedColor("white"),
	TextSecondaryColor: NamedColor("gray"),
	TextDisabledColor:  IndexColor(242),

	ButtonColor:        NamedColor("blue"),
	ButtonTextColor:    NamedColor("white"),
	InputBackground:    IndexColor(236),
	InputText:          NamedColor("white"),

	SuccessColor:       NamedColor("green"),
	WarningColor:       NamedColor("yellow"),
	ErrorColor:         NamedColor("red"),
	InfoColor:          NamedColor("blue"),

	BorderColor:        NamedColor("white"),
	DividerColor:       IndexColor(238),

	SelectionColor:     NamedColor("blue"),
	HighlightColor:     NamedColor("yellow"),
}

// LightTheme provides a light color scheme.
var LightTheme = &Theme{
	PrimaryColor:       HexColor("#1976D2"),
	SecondaryColor:     HexColor("#00796B"),
	BackgroundColor:    NamedColor("white"),
	SurfaceColor:       IndexColor(255),

	TextColor:          NamedColor("black"),
	TextSecondaryColor: IndexColor(240),
	TextDisabledColor:  IndexColor(245),

	ButtonColor:        HexColor("#1976D2"),
	ButtonTextColor:    NamedColor("white"),
	InputBackground:    IndexColor(254),
	InputText:          NamedColor("black"),

	SuccessColor:       HexColor("#388E3C"),
	WarningColor:       HexColor("#F57C00"),
	ErrorColor:         HexColor("#D32F2F"),
	InfoColor:          HexColor("#1976D2"),

	BorderColor:        IndexColor(240),
	DividerColor:       IndexColor(250),

	SelectionColor:     HexColor("#1976D2"),
	HighlightColor:     HexColor("#FFC107"),
}

// MonochromeTheme provides a black and white theme.
var MonochromeTheme = &Theme{
	PrimaryColor:       NamedColor("white"),
	SecondaryColor:     NamedColor("gray"),
	BackgroundColor:    NamedColor("black"),
	SurfaceColor:       NamedColor("black"),

	TextColor:          NamedColor("white"),
	TextSecondaryColor: NamedColor("gray"),
	TextDisabledColor:  NamedColor("darkgray"),

	ButtonColor:        NamedColor("white"),
	ButtonTextColor:    NamedColor("black"),
	InputBackground:    NamedColor("black"),
	InputText:          NamedColor("white"),

	SuccessColor:       NamedColor("white"),
	WarningColor:       NamedColor("white"),
	ErrorColor:         NamedColor("white"),
	InfoColor:          NamedColor("white"),

	BorderColor:        NamedColor("white"),
	DividerColor:       NamedColor("gray"),

	SelectionColor:     NamedColor("white"),
	HighlightColor:     NamedColor("white"),
}

// OceanTheme provides a cool blue theme.
var OceanTheme = &Theme{
	PrimaryColor:       HexColor("#0288D1"),
	SecondaryColor:     HexColor("#00ACC1"),
	BackgroundColor:    HexColor("#0D1B2A"),
	SurfaceColor:       HexColor("#1B2838"),

	TextColor:          HexColor("#E0E0E0"),
	TextSecondaryColor: HexColor("#90A4AE"),
	TextDisabledColor:  HexColor("#546E7A"),

	ButtonColor:        HexColor("#0288D1"),
	ButtonTextColor:    NamedColor("white"),
	InputBackground:    HexColor("#1B2838"),
	InputText:          HexColor("#E0E0E0"),

	SuccessColor:       HexColor("#00C853"),
	WarningColor:       HexColor("#FFB300"),
	ErrorColor:         HexColor("#FF5252"),
	InfoColor:          HexColor("#29B6F6"),

	BorderColor:        HexColor("#37474F"),
	DividerColor:       HexColor("#263238"),

	SelectionColor:     HexColor("#0288D1"),
	HighlightColor:     HexColor("#FFB300"),
}

// ForestTheme provides a green nature theme.
var ForestTheme = &Theme{
	PrimaryColor:       HexColor("#2E7D32"),
	SecondaryColor:     HexColor("#558B2F"),
	BackgroundColor:    HexColor("#1B2B1B"),
	SurfaceColor:       HexColor("#2A3D2A"),

	TextColor:          HexColor("#E8F5E9"),
	TextSecondaryColor: HexColor("#A5D6A7"),
	TextDisabledColor:  HexColor("#4CAF50"),

	ButtonColor:        HexColor("#2E7D32"),
	ButtonTextColor:    NamedColor("white"),
	InputBackground:    HexColor("#2A3D2A"),
	InputText:          HexColor("#E8F5E9"),

	SuccessColor:       HexColor("#66BB6A"),
	WarningColor:       HexColor("#FFCA28"),
	ErrorColor:         HexColor("#EF5350"),
	InfoColor:          HexColor("#42A5F5"),

	BorderColor:        HexColor("#388E3C"),
	DividerColor:       HexColor("#1B5E20"),

	SelectionColor:     HexColor("#2E7D32"),
	HighlightColor:     HexColor("#FFCA28"),
}

// SunsetTheme provides a warm orange/red theme.
var SunsetTheme = &Theme{
	PrimaryColor:       HexColor("#E65100"),
	SecondaryColor:     HexColor("#FF6D00"),
	BackgroundColor:    HexColor("#1A0F0A"),
	SurfaceColor:       HexColor("#2D1810"),

	TextColor:          HexColor("#FFF3E0"),
	TextSecondaryColor: HexColor("#FFCC80"),
	TextDisabledColor:  HexColor("#A1887F"),

	ButtonColor:        HexColor("#E65100"),
	ButtonTextColor:    NamedColor("white"),
	InputBackground:    HexColor("#2D1810"),
	InputText:          HexColor("#FFF3E0"),

	SuccessColor:       HexColor("#66BB6A"),
	WarningColor:       HexColor("#FFA726"),
	ErrorColor:         HexColor("#EF5350"),
	InfoColor:          HexColor("#42A5F5"),

	BorderColor:        HexColor("#BF360C"),
	DividerColor:       HexColor("#3E2723"),

	SelectionColor:     HexColor("#E65100"),
	HighlightColor:     HexColor("#FFD54F"),
}
