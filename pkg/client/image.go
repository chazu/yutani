package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Image represents an image display widget.
type Image struct {
	baseWidget
}

// ImageBuilder provides a fluent interface for creating image widgets.
type ImageBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewImage creates a new image builder.
func (c *Client) NewImage() *ImageBuilder {
	return &ImageBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the image's title (border title).
func (b *ImageBuilder) Title(title string) *ImageBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the image has a border.
func (b *ImageBuilder) Border(border bool) *ImageBuilder {
	b.props.Border = &border
	return b
}

// ImageData sets the image data (raw bytes).
func (b *ImageBuilder) ImageData(data []byte) *ImageBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{},
		}
	}
	b.props.GetImage().ImageData = data
	return b
}

// ImagePath sets the path to an image file.
func (b *ImageBuilder) ImagePath(path string) *ImageBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{},
		}
	}
	b.props.GetImage().ImagePath = &path
	return b
}

// Colors sets the number of colors to use for rendering.
func (b *ImageBuilder) Colors(colors int) *ImageBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{},
		}
	}
	c := int32(colors)
	b.props.GetImage().Colors = &c
	return b
}

// Dithering enables or disables dithering.
func (b *ImageBuilder) Dithering(enabled bool) *ImageBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{},
		}
	}
	b.props.GetImage().Dithering = &enabled
	return b
}

// AspectRatio sets how aspect ratio is handled.
func (b *ImageBuilder) AspectRatio(ratio pb.ImageAspectRatio) *ImageBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{},
		}
	}
	b.props.GetImage().AspectRatio = &ratio
	return b
}

// Build creates the image widget on the server.
func (b *ImageBuilder) Build() (*Image, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_IMAGE,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	image := &Image{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(image)
	return image, nil
}

// SetImageData sets the image data.
func (img *Image) SetImageData(data []byte) error {
	return img.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{
				ImageData: data,
			},
		},
	})
}

// SetColors sets the number of colors for rendering.
func (img *Image) SetColors(colors int) error {
	c := int32(colors)
	return img.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{
				Colors: &c,
			},
		},
	})
}

// SetDithering enables or disables dithering.
func (img *Image) SetDithering(enabled bool) error {
	return img.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_Image{
			Image: &pb.ImageProperties{
				Dithering: &enabled,
			},
		},
	})
}

// GetColors returns the current color count setting.
func (img *Image) GetColors() (int, error) {
	props, err := img.GetProperties()
	if err != nil {
		return 0, err
	}
	if imgProps := props.GetImage(); imgProps != nil {
		if imgProps.Colors != nil {
			return int(*imgProps.Colors), nil
		}
	}
	return 0, nil
}
