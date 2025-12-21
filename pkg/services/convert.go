package services

import (
	"github.com/gdamore/tcell/v2"
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// convertColor converts a protobuf Color to tcell.Color
func convertColor(pbColor *pb.Color) tcell.Color {
	if pbColor == nil {
		return tcell.ColorDefault
	}

	switch c := pbColor.Color.(type) {
	case *pb.Color_Name:
		return tcell.GetColor(c.Name)
	case *pb.Color_Index:
		return tcell.PaletteColor(int(c.Index))
	case *pb.Color_Rgb:
		return tcell.NewRGBColor(int32(c.Rgb.R), int32(c.Rgb.G), int32(c.Rgb.B))
	default:
		return tcell.ColorDefault
	}
}

// applyAttribute applies a protobuf Attribute to a tcell.Style
func applyAttribute(style tcell.Style, attr pb.Attribute) tcell.Style {
	switch attr {
	case pb.Attribute_ATTR_BOLD:
		return style.Bold(true)
	case pb.Attribute_ATTR_ITALIC:
		return style.Italic(true)
	case pb.Attribute_ATTR_UNDERLINE:
		return style.Underline(true)
	case pb.Attribute_ATTR_REVERSE:
		return style.Reverse(true)
	case pb.Attribute_ATTR_BLINK:
		return style.Blink(true)
	case pb.Attribute_ATTR_DIM:
		return style.Dim(true)
	case pb.Attribute_ATTR_STRIKETHROUGH:
		return style.StrikeThrough(true)
	default:
		return style
	}
}

// convertColorToProto converts a tcell.Color to protobuf Color
func convertColorToProto(color tcell.Color) *pb.Color {
	if color == tcell.ColorDefault {
		return nil
	}

	// Try to get RGB values
	r, g, b := color.RGB()
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

// convertStyleToProto converts a tcell.Style to protobuf Style
func convertStyleToProto(style tcell.Style) *pb.Style {
	fg, bg, attrs := style.Decompose()

	pbStyle := &pb.Style{
		Foreground: convertColorToProto(fg),
		Background: convertColorToProto(bg),
		Attributes: []pb.Attribute{},
	}

	// Convert attributes
	if attrs&tcell.AttrBold != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_BOLD)
	}
	if attrs&tcell.AttrItalic != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_ITALIC)
	}
	if attrs&tcell.AttrUnderline != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_UNDERLINE)
	}
	if attrs&tcell.AttrReverse != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_REVERSE)
	}
	if attrs&tcell.AttrBlink != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_BLINK)
	}
	if attrs&tcell.AttrDim != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_DIM)
	}
	if attrs&tcell.AttrStrikeThrough != 0 {
		pbStyle.Attributes = append(pbStyle.Attributes, pb.Attribute_ATTR_STRIKETHROUGH)
	}

	return pbStyle
}
