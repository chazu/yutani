package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// ProgressBar represents a progress bar widget.
type ProgressBar struct {
	baseWidget
}

// ProgressBarBuilder provides a fluent interface for creating progress bar widgets.
type ProgressBarBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewProgressBar creates a new progress bar builder.
func (c *Client) NewProgressBar() *ProgressBarBuilder {
	return &ProgressBarBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the progress bar's title (border title).
func (b *ProgressBarBuilder) Title(title string) *ProgressBarBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the progress bar has a border.
func (b *ProgressBarBuilder) Border(border bool) *ProgressBarBuilder {
	b.props.Border = &border
	return b
}

// Progress sets the current progress value.
func (b *ProgressBarBuilder) Progress(progress int) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	p := int32(progress)
	b.props.GetProgressBar().Progress = &p
	return b
}

// Max sets the maximum progress value.
func (b *ProgressBarBuilder) Max(max int) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	m := int32(max)
	b.props.GetProgressBar().Max = &m
	return b
}

// Label sets the progress bar label.
func (b *ProgressBarBuilder) Label(label string) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	b.props.GetProgressBar().Label = &label
	return b
}

// ShowPercentage enables or disables percentage display.
func (b *ProgressBarBuilder) ShowPercentage(show bool) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	b.props.GetProgressBar().ShowPercentage = &show
	return b
}

// FilledColor sets the color of the filled portion.
func (b *ProgressBarBuilder) FilledColor(color *pb.Color) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	b.props.GetProgressBar().FilledColor = color
	return b
}

// EmptyColor sets the color of the empty portion.
func (b *ProgressBarBuilder) EmptyColor(color *pb.Color) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	b.props.GetProgressBar().EmptyColor = color
	return b
}

// FilledChar sets the character used for the filled portion.
func (b *ProgressBarBuilder) FilledChar(char string) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	b.props.GetProgressBar().FilledChar = &char
	return b
}

// EmptyChar sets the character used for the empty portion.
func (b *ProgressBarBuilder) EmptyChar(char string) *ProgressBarBuilder {
	if b.props.TypeProperties == nil {
		b.props.TypeProperties = &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{},
		}
	}
	b.props.GetProgressBar().EmptyChar = &char
	return b
}

// Build creates the progress bar widget on the server.
func (b *ProgressBarBuilder) Build() (*ProgressBar, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_PROGRESS_BAR,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	progressBar := &ProgressBar{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(progressBar)
	return progressBar, nil
}

// SetProgress sets the current progress value.
func (pbar *ProgressBar) SetProgress(progress int) error {
	p := int32(progress)
	return pbar.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{
				Progress: &p,
			},
		},
	})
}

// GetProgress returns the current progress value.
func (pbar *ProgressBar) GetProgress() (int, error) {
	props, err := pbar.GetProperties()
	if err != nil {
		return 0, err
	}
	if pbarProps := props.GetProgressBar(); pbarProps != nil {
		if pbarProps.Progress != nil {
			return int(*pbarProps.Progress), nil
		}
	}
	return 0, nil
}

// SetMax sets the maximum progress value.
func (pbar *ProgressBar) SetMax(max int) error {
	m := int32(max)
	return pbar.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{
				Max: &m,
			},
		},
	})
}

// GetMax returns the maximum progress value.
func (pbar *ProgressBar) GetMax() (int, error) {
	props, err := pbar.GetProperties()
	if err != nil {
		return 100, err
	}
	if pbarProps := props.GetProgressBar(); pbarProps != nil {
		if pbarProps.Max != nil {
			return int(*pbarProps.Max), nil
		}
	}
	return 100, nil
}

// SetLabel sets the progress bar label.
func (pbar *ProgressBar) SetLabel(label string) error {
	return pbar.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{
				Label: &label,
			},
		},
	})
}

// GetLabel returns the progress bar label.
func (pbar *ProgressBar) GetLabel() (string, error) {
	props, err := pbar.GetProperties()
	if err != nil {
		return "", err
	}
	if pbarProps := props.GetProgressBar(); pbarProps != nil {
		if pbarProps.Label != nil {
			return *pbarProps.Label, nil
		}
	}
	return "", nil
}

// SetShowPercentage enables or disables percentage display.
func (pbar *ProgressBar) SetShowPercentage(show bool) error {
	return pbar.setProperty(&pb.WidgetProperties{
		TypeProperties: &pb.WidgetProperties_ProgressBar{
			ProgressBar: &pb.ProgressBarProperties{
				ShowPercentage: &show,
			},
		},
	})
}
