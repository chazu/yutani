package services

import (
	"fmt"

	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"github.com/rivo/tview"
)

// Widget creation functions

func (s *WidgetService) createBox(props *pb.WidgetProperties) *tview.Box {
	box := tview.NewBox()
	if props != nil {
		s.applyBoxProperties(box, props)
	}
	return box
}

func (s *WidgetService) createTextView(props *pb.WidgetProperties) *tview.TextView {
	tv := tview.NewTextView()
	if props != nil {
		s.applyBoxProperties(tv.Box, props)
		if props.TypeProperties != nil {
			if tvProps, ok := props.TypeProperties.(*pb.WidgetProperties_TextView); ok && tvProps.TextView != nil {
				if tvProps.TextView.Text != nil {
					tv.SetText(*tvProps.TextView.Text)
				}
				if tvProps.TextView.DynamicColors != nil {
					tv.SetDynamicColors(*tvProps.TextView.DynamicColors)
				}
				if tvProps.TextView.WordWrap != nil {
					tv.SetWordWrap(*tvProps.TextView.WordWrap)
				}
				if tvProps.TextView.Scrollable != nil {
					tv.SetScrollable(*tvProps.TextView.Scrollable)
				}
				if tvProps.TextView.TextColor != nil {
					tv.SetTextColor(convertColor(tvProps.TextView.TextColor))
				}
			}
		}
	}
	return tv
}

func (s *WidgetService) createInputField(props *pb.WidgetProperties) *tview.InputField {
	input := tview.NewInputField()
	if props != nil {
		s.applyBoxProperties(input.Box, props)
		if props.TypeProperties != nil {
			if inputProps, ok := props.TypeProperties.(*pb.WidgetProperties_InputField); ok && inputProps.InputField != nil {
				if inputProps.InputField.Label != nil {
					input.SetLabel(*inputProps.InputField.Label)
				}
				if inputProps.InputField.Placeholder != nil {
					input.SetPlaceholder(*inputProps.InputField.Placeholder)
				}
				if inputProps.InputField.Text != nil {
					input.SetText(*inputProps.InputField.Text)
				}
				if inputProps.InputField.FieldWidth != nil {
					input.SetFieldWidth(int(*inputProps.InputField.FieldWidth))
				}
				if inputProps.InputField.LabelColor != nil {
					input.SetLabelColor(convertColor(inputProps.InputField.LabelColor))
				}
				if inputProps.InputField.FieldTextColor != nil {
					input.SetFieldTextColor(convertColor(inputProps.InputField.FieldTextColor))
				}
				if inputProps.InputField.FieldBackgroundColor != nil {
					input.SetFieldBackgroundColor(convertColor(inputProps.InputField.FieldBackgroundColor))
				}
			}
		}
	}
	return input
}

func (s *WidgetService) createButton(props *pb.WidgetProperties) *tview.Button {
	label := "Button"
	if props != nil && props.TypeProperties != nil {
		if btnProps, ok := props.TypeProperties.(*pb.WidgetProperties_Button); ok && btnProps.Button != nil {
			if btnProps.Button.Label != nil {
				label = *btnProps.Button.Label
			}
		}
	}

	btn := tview.NewButton(label)
	if props != nil {
		s.applyBoxProperties(btn.Box, props)
		if props.TypeProperties != nil {
			if btnProps, ok := props.TypeProperties.(*pb.WidgetProperties_Button); ok && btnProps.Button != nil {
				if btnProps.Button.LabelColor != nil {
					btn.SetLabelColor(convertColor(btnProps.Button.LabelColor))
				}
				if btnProps.Button.BackgroundColor != nil {
					btn.SetBackgroundColor(convertColor(btnProps.Button.BackgroundColor))
				}
			}
		}
	}
	return btn
}

func (s *WidgetService) createCheckbox(props *pb.WidgetProperties) *tview.Checkbox {
	label := ""
	checked := false
	if props != nil && props.TypeProperties != nil {
		if cbProps, ok := props.TypeProperties.(*pb.WidgetProperties_Checkbox); ok && cbProps.Checkbox != nil {
			if cbProps.Checkbox.Label != nil {
				label = *cbProps.Checkbox.Label
			}
			if cbProps.Checkbox.Checked != nil {
				checked = *cbProps.Checkbox.Checked
			}
		}
	}

	cb := tview.NewCheckbox()
	cb.SetLabel(label)
	cb.SetChecked(checked)

	if props != nil {
		s.applyBoxProperties(cb.Box, props)
		if props.TypeProperties != nil {
			if cbProps, ok := props.TypeProperties.(*pb.WidgetProperties_Checkbox); ok && cbProps.Checkbox != nil {
				if cbProps.Checkbox.LabelColor != nil {
					cb.SetLabelColor(convertColor(cbProps.Checkbox.LabelColor))
				}
			}
		}
	}
	return cb
}

// applyBoxProperties applies common Box properties
func (s *WidgetService) applyBoxProperties(box *tview.Box, props *pb.WidgetProperties) {
	if props.Border != nil {
		box.SetBorder(*props.Border)
	}
	if props.Title != nil {
		box.SetTitle(*props.Title)
	}
	if props.TitleAlign != nil {
		box.SetTitleAlign(convertAlignment(*props.TitleAlign))
	}
	if props.BackgroundColor != nil {
		box.SetBackgroundColor(convertColor(props.BackgroundColor))
	}
	if props.TitleColor != nil {
		box.SetTitleColor(convertColor(props.TitleColor))
	}
}

// applyTextViewProperties applies TextView-specific properties
func (s *WidgetService) applyTextViewProperties(tv *tview.TextView, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	tvProps, ok := props.TypeProperties.(*pb.WidgetProperties_TextView)
	if !ok || tvProps.TextView == nil {
		return nil
	}

	if tvProps.TextView.Text != nil {
		tv.SetText(*tvProps.TextView.Text)
	}
	if tvProps.TextView.DynamicColors != nil {
		tv.SetDynamicColors(*tvProps.TextView.DynamicColors)
	}
	if tvProps.TextView.WordWrap != nil {
		tv.SetWordWrap(*tvProps.TextView.WordWrap)
	}
	if tvProps.TextView.Scrollable != nil {
		tv.SetScrollable(*tvProps.TextView.Scrollable)
	}
	if tvProps.TextView.TextColor != nil {
		tv.SetTextColor(convertColor(tvProps.TextView.TextColor))
	}

	return nil
}

// applyInputFieldProperties applies InputField-specific properties
func (s *WidgetService) applyInputFieldProperties(input *tview.InputField, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	inputProps, ok := props.TypeProperties.(*pb.WidgetProperties_InputField)
	if !ok || inputProps.InputField == nil {
		return nil
	}

	if inputProps.InputField.Label != nil {
		input.SetLabel(*inputProps.InputField.Label)
	}
	if inputProps.InputField.Placeholder != nil {
		input.SetPlaceholder(*inputProps.InputField.Placeholder)
	}
	if inputProps.InputField.Text != nil {
		input.SetText(*inputProps.InputField.Text)
	}
	if inputProps.InputField.FieldWidth != nil {
		input.SetFieldWidth(int(*inputProps.InputField.FieldWidth))
	}
	if inputProps.InputField.LabelColor != nil {
		input.SetLabelColor(convertColor(inputProps.InputField.LabelColor))
	}
	if inputProps.InputField.FieldTextColor != nil {
		input.SetFieldTextColor(convertColor(inputProps.InputField.FieldTextColor))
	}
	if inputProps.InputField.FieldBackgroundColor != nil {
		input.SetFieldBackgroundColor(convertColor(inputProps.InputField.FieldBackgroundColor))
	}

	return nil
}

// applyButtonProperties applies Button-specific properties
func (s *WidgetService) applyButtonProperties(btn *tview.Button, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	btnProps, ok := props.TypeProperties.(*pb.WidgetProperties_Button)
	if !ok || btnProps.Button == nil {
		return nil
	}

	if btnProps.Button.Label != nil {
		btn.SetLabel(*btnProps.Button.Label)
	}
	if btnProps.Button.LabelColor != nil {
		btn.SetLabelColor(convertColor(btnProps.Button.LabelColor))
	}
	if btnProps.Button.BackgroundColor != nil {
		btn.SetBackgroundColor(convertColor(btnProps.Button.BackgroundColor))
	}

	return nil
}

// applyCheckboxProperties applies Checkbox-specific properties
func (s *WidgetService) applyCheckboxProperties(cb *tview.Checkbox, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	cbProps, ok := props.TypeProperties.(*pb.WidgetProperties_Checkbox)
	if !ok || cbProps.Checkbox == nil {
		return nil
	}

	if cbProps.Checkbox.Label != nil {
		cb.SetLabel(*cbProps.Checkbox.Label)
	}
	if cbProps.Checkbox.Checked != nil {
		cb.SetChecked(*cbProps.Checkbox.Checked)
	}
	if cbProps.Checkbox.LabelColor != nil {
		cb.SetLabelColor(convertColor(cbProps.Checkbox.LabelColor))
	}

	return nil
}

// convertAlignment converts proto Alignment to tview alignment
func convertAlignment(align pb.Alignment) int {
	switch align {
	case pb.Alignment_ALIGN_LEFT:
		return tview.AlignLeft
	case pb.Alignment_ALIGN_CENTER:
		return tview.AlignCenter
	case pb.Alignment_ALIGN_RIGHT:
		return tview.AlignRight
	default:
		return tview.AlignLeft
	}
}

// Complex widget creation functions

func (s *WidgetService) createList(props *pb.WidgetProperties) *tview.List {
	list := tview.NewList()
	if props != nil {
		s.applyBoxProperties(list.Box, props)
		if props.TypeProperties != nil {
			if listProps, ok := props.TypeProperties.(*pb.WidgetProperties_List); ok && listProps.List != nil {
				if listProps.List.ShowSecondaryText != nil {
					list.ShowSecondaryText(*listProps.List.ShowSecondaryText)
				}
				if listProps.List.MainTextColor != nil {
					list.SetMainTextColor(convertColor(listProps.List.MainTextColor))
				}
				if listProps.List.SecondaryTextColor != nil {
					list.SetSecondaryTextColor(convertColor(listProps.List.SecondaryTextColor))
				}
				if listProps.List.SelectedTextColor != nil {
					list.SetSelectedTextColor(convertColor(listProps.List.SelectedTextColor))
				}
				if listProps.List.SelectedBackgroundColor != nil {
					list.SetSelectedBackgroundColor(convertColor(listProps.List.SelectedBackgroundColor))
				}
			}
		}
	}
	return list
}

func (s *WidgetService) createTable(props *pb.WidgetProperties) *tview.Table {
	table := tview.NewTable()
	if props != nil {
		s.applyBoxProperties(table.Box, props)
		if props.TypeProperties != nil {
			if tableProps, ok := props.TypeProperties.(*pb.WidgetProperties_Table); ok && tableProps.Table != nil {
				if tableProps.Table.Borders != nil {
					table.SetBorders(*tableProps.Table.Borders)
				}
				if tableProps.Table.Selectable != nil {
					table.SetSelectable(*tableProps.Table.Selectable, *tableProps.Table.Selectable)
				}
				if tableProps.Table.BordersColor != nil {
					table.SetBordersColor(convertColor(tableProps.Table.BordersColor))
				}
			}
		}
	}
	return table
}

func (s *WidgetService) createTreeView(props *pb.WidgetProperties) *tview.TreeView {
	tree := tview.NewTreeView()
	if props != nil {
		s.applyBoxProperties(tree.Box, props)
		if props.TypeProperties != nil {
			if treeProps, ok := props.TypeProperties.(*pb.WidgetProperties_Tree); ok && treeProps.Tree != nil {
				if treeProps.Tree.NodeTextColor != nil {
					tree.SetGraphicsColor(convertColor(treeProps.Tree.NodeTextColor))
				}
				// Note: tview.TreeView doesn't have SetSelectedTextColor/SetSelectedBackgroundColor
				// These are set on individual nodes
				if treeProps.Tree.TopLevelPrefix != nil {
					// SetTopLevel expects an int (number of levels), not a string
					// We'll skip this for now as the API doesn't match
				}
				if treeProps.Tree.ShowGraphics != nil {
					tree.SetGraphics(*treeProps.Tree.ShowGraphics)
				}
			}
		}
	}
	return tree
}

func (s *WidgetService) createForm(props *pb.WidgetProperties) *tview.Form {
	form := tview.NewForm()
	if props != nil {
		s.applyBoxProperties(form.Box, props)
		if props.TypeProperties != nil {
			if formProps, ok := props.TypeProperties.(*pb.WidgetProperties_Form); ok && formProps.Form != nil {
				// Note: tview.Form doesn't have SetLabelWidth/SetFieldWidth methods
				// These would need to be set when adding individual fields
				if formProps.Form.LabelColor != nil {
					form.SetLabelColor(convertColor(formProps.Form.LabelColor))
				}
				if formProps.Form.FieldTextColor != nil {
					form.SetFieldTextColor(convertColor(formProps.Form.FieldTextColor))
				}
				if formProps.Form.FieldBackgroundColor != nil {
					form.SetFieldBackgroundColor(convertColor(formProps.Form.FieldBackgroundColor))
				}
				if formProps.Form.ButtonBackgroundColor != nil {
					form.SetButtonBackgroundColor(convertColor(formProps.Form.ButtonBackgroundColor))
				}
				if formProps.Form.ButtonTextColor != nil {
					form.SetButtonTextColor(convertColor(formProps.Form.ButtonTextColor))
				}
				if formProps.Form.Horizontal != nil {
					form.SetHorizontal(*formProps.Form.Horizontal)
				}
			}
		}
	}
	return form
}

func (s *WidgetService) createFlex(props *pb.WidgetProperties) *tview.Flex {
	flex := tview.NewFlex()
	if props != nil {
		s.applyBoxProperties(flex.Box, props)
		if props.TypeProperties != nil {
			if flexProps, ok := props.TypeProperties.(*pb.WidgetProperties_Flex); ok && flexProps.Flex != nil {
				if flexProps.Flex.Direction != nil {
					if *flexProps.Flex.Direction == pb.FlexDirection_FLEX_COLUMN {
						flex.SetDirection(tview.FlexColumn)
					} else {
						flex.SetDirection(tview.FlexRow)
					}
				}
				if flexProps.Flex.FullScreen != nil {
					flex.SetFullScreen(*flexProps.Flex.FullScreen)
				}
			}
		}
	}
	return flex
}

func (s *WidgetService) createGrid(props *pb.WidgetProperties) *tview.Grid {
	grid := tview.NewGrid()
	if props != nil {
		s.applyBoxProperties(grid.Box, props)
		if props.TypeProperties != nil {
			if gridProps, ok := props.TypeProperties.(*pb.WidgetProperties_Grid); ok && gridProps.Grid != nil {
				if gridProps.Grid.Rows != nil {
					grid.SetRows(int(*gridProps.Grid.Rows))
				}
				if gridProps.Grid.Columns != nil {
					grid.SetColumns(int(*gridProps.Grid.Columns))
				}
				if gridProps.Grid.MinWidth != nil && gridProps.Grid.MinHeight != nil {
					grid.SetMinSize(int(*gridProps.Grid.MinWidth), int(*gridProps.Grid.MinHeight))
				}
			}
		}
	}
	return grid
}

func (s *WidgetService) createPages(props *pb.WidgetProperties) *tview.Pages {
	pages := tview.NewPages()
	if props != nil {
		s.applyBoxProperties(pages.Box, props)
		// Pages doesn't have many configurable properties in tview
	}
	return pages
}

func (s *WidgetService) createDropdown(props *pb.WidgetProperties) *tview.DropDown {
	dd := tview.NewDropDown()
	if props != nil {
		s.applyBoxProperties(dd.Box, props)
		if props.TypeProperties != nil {
			if ddProps, ok := props.TypeProperties.(*pb.WidgetProperties_Dropdown); ok && ddProps.Dropdown != nil {
				if ddProps.Dropdown.Label != nil {
					dd.SetLabel(*ddProps.Dropdown.Label)
				}
				if len(ddProps.Dropdown.Options) > 0 {
					dd.SetOptions(ddProps.Dropdown.Options, nil)
				}
				if ddProps.Dropdown.SelectedIndex != nil {
					dd.SetCurrentOption(int(*ddProps.Dropdown.SelectedIndex))
				}
				if ddProps.Dropdown.FieldWidth != nil {
					dd.SetFieldWidth(int(*ddProps.Dropdown.FieldWidth))
				}
				if ddProps.Dropdown.LabelColor != nil {
					dd.SetLabelColor(convertColor(ddProps.Dropdown.LabelColor))
				}
				if ddProps.Dropdown.FieldTextColor != nil {
					dd.SetFieldTextColor(convertColor(ddProps.Dropdown.FieldTextColor))
				}
				if ddProps.Dropdown.FieldBackgroundColor != nil {
					dd.SetFieldBackgroundColor(convertColor(ddProps.Dropdown.FieldBackgroundColor))
				}
			}
		}
	}
	return dd
}

// applyDropdownProperties applies Dropdown-specific properties
func (s *WidgetService) applyDropdownProperties(dd *tview.DropDown, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	ddProps, ok := props.TypeProperties.(*pb.WidgetProperties_Dropdown)
	if !ok || ddProps.Dropdown == nil {
		return nil
	}

	if ddProps.Dropdown.Label != nil {
		dd.SetLabel(*ddProps.Dropdown.Label)
	}
	if len(ddProps.Dropdown.Options) > 0 {
		dd.SetOptions(ddProps.Dropdown.Options, nil)
	}
	if ddProps.Dropdown.SelectedIndex != nil {
		dd.SetCurrentOption(int(*ddProps.Dropdown.SelectedIndex))
	}
	if ddProps.Dropdown.FieldWidth != nil {
		dd.SetFieldWidth(int(*ddProps.Dropdown.FieldWidth))
	}
	if ddProps.Dropdown.LabelColor != nil {
		dd.SetLabelColor(convertColor(ddProps.Dropdown.LabelColor))
	}
	if ddProps.Dropdown.FieldTextColor != nil {
		dd.SetFieldTextColor(convertColor(ddProps.Dropdown.FieldTextColor))
	}
	if ddProps.Dropdown.FieldBackgroundColor != nil {
		dd.SetFieldBackgroundColor(convertColor(ddProps.Dropdown.FieldBackgroundColor))
	}

	return nil
}

// TextArea widget

func (s *WidgetService) createTextArea(props *pb.WidgetProperties) *tview.TextArea {
	ta := tview.NewTextArea()
	if props != nil {
		s.applyBoxProperties(ta.Box, props)
		if props.TypeProperties != nil {
			if taProps, ok := props.TypeProperties.(*pb.WidgetProperties_TextArea); ok && taProps.TextArea != nil {
				if taProps.TextArea.Text != nil {
					ta.SetText(*taProps.TextArea.Text, true)
				}
				if taProps.TextArea.Placeholder != nil {
					ta.SetPlaceholder(*taProps.TextArea.Placeholder)
				}
				if taProps.TextArea.MaxLength != nil {
					ta.SetMaxLength(int(*taProps.TextArea.MaxLength))
				}
				if taProps.TextArea.WordWrap != nil {
					ta.SetWordWrap(*taProps.TextArea.WordWrap)
				}
				if taProps.TextArea.TextColor != nil {
					ta.SetTextStyle(ta.GetTextStyle().Foreground(convertColor(taProps.TextArea.TextColor)))
				}
				if taProps.TextArea.PlaceholderColor != nil {
					ta.SetPlaceholderStyle(ta.GetPlaceholderStyle().Foreground(convertColor(taProps.TextArea.PlaceholderColor)))
				}
			}
		}
	}
	return ta
}

func (s *WidgetService) applyTextAreaProperties(ta *tview.TextArea, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	taProps, ok := props.TypeProperties.(*pb.WidgetProperties_TextArea)
	if !ok || taProps.TextArea == nil {
		return nil
	}

	if taProps.TextArea.Text != nil {
		ta.SetText(*taProps.TextArea.Text, true)
	}
	if taProps.TextArea.Placeholder != nil {
		ta.SetPlaceholder(*taProps.TextArea.Placeholder)
	}
	if taProps.TextArea.MaxLength != nil {
		ta.SetMaxLength(int(*taProps.TextArea.MaxLength))
	}
	if taProps.TextArea.WordWrap != nil {
		ta.SetWordWrap(*taProps.TextArea.WordWrap)
	}
	if taProps.TextArea.TextColor != nil {
		ta.SetTextStyle(ta.GetTextStyle().Foreground(convertColor(taProps.TextArea.TextColor)))
	}
	if taProps.TextArea.PlaceholderColor != nil {
		ta.SetPlaceholderStyle(ta.GetPlaceholderStyle().Foreground(convertColor(taProps.TextArea.PlaceholderColor)))
	}

	return nil
}

// Modal widget

func (s *WidgetService) createModal(props *pb.WidgetProperties) *tview.Modal {
	modal := tview.NewModal()
	if props != nil {
		// Modal doesn't expose Box directly, but we can set title via SetTitle if border is set
		if props.Title != nil {
			modal.SetTitle(*props.Title)
		}
		if props.Border != nil {
			modal.SetBorder(*props.Border)
		}
		if props.TypeProperties != nil {
			if modalProps, ok := props.TypeProperties.(*pb.WidgetProperties_Modal); ok && modalProps.Modal != nil {
				if modalProps.Modal.Text != nil {
					modal.SetText(*modalProps.Modal.Text)
				}
				if len(modalProps.Modal.Buttons) > 0 {
					modal.AddButtons(modalProps.Modal.Buttons)
				}
				if modalProps.Modal.TextColor != nil {
					modal.SetTextColor(convertColor(modalProps.Modal.TextColor))
				}
				if modalProps.Modal.BackgroundColor != nil {
					modal.SetBackgroundColor(convertColor(modalProps.Modal.BackgroundColor))
				}
				if modalProps.Modal.ButtonBackgroundColor != nil {
					modal.SetButtonBackgroundColor(convertColor(modalProps.Modal.ButtonBackgroundColor))
				}
				if modalProps.Modal.ButtonTextColor != nil {
					modal.SetButtonTextColor(convertColor(modalProps.Modal.ButtonTextColor))
				}
			}
		}
	}
	return modal
}

func (s *WidgetService) applyModalProperties(modal *tview.Modal, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	modalProps, ok := props.TypeProperties.(*pb.WidgetProperties_Modal)
	if !ok || modalProps.Modal == nil {
		return nil
	}

	if modalProps.Modal.Text != nil {
		modal.SetText(*modalProps.Modal.Text)
	}
	if len(modalProps.Modal.Buttons) > 0 {
		modal.ClearButtons()
		modal.AddButtons(modalProps.Modal.Buttons)
	}
	if modalProps.Modal.TextColor != nil {
		modal.SetTextColor(convertColor(modalProps.Modal.TextColor))
	}
	if modalProps.Modal.BackgroundColor != nil {
		modal.SetBackgroundColor(convertColor(modalProps.Modal.BackgroundColor))
	}
	if modalProps.Modal.ButtonBackgroundColor != nil {
		modal.SetButtonBackgroundColor(convertColor(modalProps.Modal.ButtonBackgroundColor))
	}
	if modalProps.Modal.ButtonTextColor != nil {
		modal.SetButtonTextColor(convertColor(modalProps.Modal.ButtonTextColor))
	}

	return nil
}

// Image widget

func (s *WidgetService) createImage(props *pb.WidgetProperties) *tview.Image {
	img := tview.NewImage()
	if props != nil {
		s.applyBoxProperties(img.Box, props)
		if props.TypeProperties != nil {
			if imgProps, ok := props.TypeProperties.(*pb.WidgetProperties_Image); ok && imgProps.Image != nil {
				if imgProps.Image.Colors != nil {
					img.SetColors(int(*imgProps.Image.Colors))
				}
				if imgProps.Image.Dithering != nil {
					img.SetDithering(tview.DitheringFloydSteinberg)
				}
				// Note: Image data/path would need special handling to load the image
			}
		}
	}
	return img
}

func (s *WidgetService) applyImageProperties(img *tview.Image, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	imgProps, ok := props.TypeProperties.(*pb.WidgetProperties_Image)
	if !ok || imgProps.Image == nil {
		return nil
	}

	if imgProps.Image.Colors != nil {
		img.SetColors(int(*imgProps.Image.Colors))
	}
	if imgProps.Image.Dithering != nil && *imgProps.Image.Dithering {
		img.SetDithering(tview.DitheringFloydSteinberg)
	}

	return nil
}

// ProgressBar widget (implemented as styled TextView)
// tview doesn't have a native progress bar, so we use a custom wrapper

type ProgressBar struct {
	*tview.TextView
	progress       int
	max            int
	label          string
	showPercentage bool
	filledChar     rune
	emptyChar      rune
	filledColor    string
	emptyColor     string
}

func newProgressBar() *ProgressBar {
	pb := &ProgressBar{
		TextView:       tview.NewTextView(),
		progress:       0,
		max:            100,
		filledChar:     '█',
		emptyChar:      '░',
		showPercentage: true,
		filledColor:    "green",
		emptyColor:     "gray",
	}
	pb.TextView.SetDynamicColors(true)
	pb.update()
	return pb
}

func (pb *ProgressBar) SetProgress(progress int) {
	if progress < 0 {
		progress = 0
	}
	if progress > pb.max {
		progress = pb.max
	}
	pb.progress = progress
	pb.update()
}

func (pb *ProgressBar) SetMax(max int) {
	if max < 1 {
		max = 1
	}
	pb.max = max
	pb.update()
}

func (pb *ProgressBar) SetLabel(label string) {
	pb.label = label
	pb.update()
}

func (pb *ProgressBar) SetShowPercentage(show bool) {
	pb.showPercentage = show
	pb.update()
}

func (pb *ProgressBar) update() {
	// Calculate filled portion
	_, _, width, _ := pb.TextView.GetInnerRect()
	if width < 10 {
		width = 20 // Default width
	}

	// Reserve space for label and percentage
	barWidth := width
	labelPart := ""
	if pb.label != "" {
		labelPart = pb.label + " "
		barWidth -= len(labelPart)
	}
	if pb.showPercentage {
		barWidth -= 5 // " 100%"
	}
	if barWidth < 5 {
		barWidth = 5
	}

	// Calculate fill
	fillWidth := 0
	if pb.max > 0 {
		fillWidth = (pb.progress * barWidth) / pb.max
	}
	emptyWidth := barWidth - fillWidth

	// Build the bar
	filled := ""
	for i := 0; i < fillWidth; i++ {
		filled += string(pb.filledChar)
	}
	empty := ""
	for i := 0; i < emptyWidth; i++ {
		empty += string(pb.emptyChar)
	}

	// Format with colors
	text := labelPart
	text += "[" + pb.filledColor + "]" + filled + "[-]"
	text += "[" + pb.emptyColor + "]" + empty + "[-]"

	if pb.showPercentage {
		pct := 0
		if pb.max > 0 {
			pct = (pb.progress * 100) / pb.max
		}
		text += fmt.Sprintf(" %3d%%", pct)
	}

	pb.TextView.SetText(text)
}

func (s *WidgetService) createProgressBar(props *pb.WidgetProperties) *ProgressBar {
	pbar := newProgressBar()
	if props != nil {
		s.applyBoxProperties(pbar.Box, props)
		if props.TypeProperties != nil {
			if pbarProps, ok := props.TypeProperties.(*pb.WidgetProperties_ProgressBar); ok && pbarProps.ProgressBar != nil {
				if pbarProps.ProgressBar.Max != nil {
					pbar.SetMax(int(*pbarProps.ProgressBar.Max))
				}
				if pbarProps.ProgressBar.Progress != nil {
					pbar.SetProgress(int(*pbarProps.ProgressBar.Progress))
				}
				if pbarProps.ProgressBar.Label != nil {
					pbar.SetLabel(*pbarProps.ProgressBar.Label)
				}
				if pbarProps.ProgressBar.ShowPercentage != nil {
					pbar.SetShowPercentage(*pbarProps.ProgressBar.ShowPercentage)
				}
				if pbarProps.ProgressBar.FilledChar != nil && len(*pbarProps.ProgressBar.FilledChar) > 0 {
					pbar.filledChar = []rune(*pbarProps.ProgressBar.FilledChar)[0]
				}
				if pbarProps.ProgressBar.EmptyChar != nil && len(*pbarProps.ProgressBar.EmptyChar) > 0 {
					pbar.emptyChar = []rune(*pbarProps.ProgressBar.EmptyChar)[0]
				}
				if pbarProps.ProgressBar.FilledColor != nil {
					pbar.filledColor = colorToString(pbarProps.ProgressBar.FilledColor)
				}
				if pbarProps.ProgressBar.EmptyColor != nil {
					pbar.emptyColor = colorToString(pbarProps.ProgressBar.EmptyColor)
				}
				pbar.update()
			}
		}
	}
	return pbar
}

func (s *WidgetService) applyProgressBarProperties(pbar *ProgressBar, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	pbarProps, ok := props.TypeProperties.(*pb.WidgetProperties_ProgressBar)
	if !ok || pbarProps.ProgressBar == nil {
		return nil
	}

	if pbarProps.ProgressBar.Max != nil {
		pbar.SetMax(int(*pbarProps.ProgressBar.Max))
	}
	if pbarProps.ProgressBar.Progress != nil {
		pbar.SetProgress(int(*pbarProps.ProgressBar.Progress))
	}
	if pbarProps.ProgressBar.Label != nil {
		pbar.SetLabel(*pbarProps.ProgressBar.Label)
	}
	if pbarProps.ProgressBar.ShowPercentage != nil {
		pbar.SetShowPercentage(*pbarProps.ProgressBar.ShowPercentage)
	}
	if pbarProps.ProgressBar.FilledChar != nil && len(*pbarProps.ProgressBar.FilledChar) > 0 {
		pbar.filledChar = []rune(*pbarProps.ProgressBar.FilledChar)[0]
	}
	if pbarProps.ProgressBar.EmptyChar != nil && len(*pbarProps.ProgressBar.EmptyChar) > 0 {
		pbar.emptyChar = []rune(*pbarProps.ProgressBar.EmptyChar)[0]
	}
	if pbarProps.ProgressBar.FilledColor != nil {
		pbar.filledColor = colorToString(pbarProps.ProgressBar.FilledColor)
	}
	if pbarProps.ProgressBar.EmptyColor != nil {
		pbar.emptyColor = colorToString(pbarProps.ProgressBar.EmptyColor)
	}
	pbar.update()

	return nil
}

// Window widget

func (s *WidgetService) createWindow(props *pb.WidgetProperties) *server.WindowPrimitive {
	win := server.NewWindowPrimitive(nil)
	if props != nil {
		if props.Title != nil {
			win.SetTitle(*props.Title)
		}
		s.applyWindowProperties(win, props)
	}
	return win
}

func (s *WidgetService) applyWindowProperties(win *server.WindowPrimitive, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	wProps, ok := props.TypeProperties.(*pb.WidgetProperties_Window)
	if !ok || wProps.Window == nil {
		return nil
	}

	wp := wProps.Window
	if wp.InitialRect != nil {
		win.SetRect(int(wp.InitialRect.X), int(wp.InitialRect.Y), int(wp.InitialRect.Width), int(wp.InitialRect.Height))
		win.SetRestoredRect(server.Rect{
			X: int(wp.InitialRect.X), Y: int(wp.InitialRect.Y),
			Width: int(wp.InitialRect.Width), Height: int(wp.InitialRect.Height),
		})
	}
	if wp.Resizable != nil {
		win.SetResizable(*wp.Resizable)
	}
	if wp.Movable != nil {
		win.SetMovable(*wp.Movable)
	}
	if wp.Closable != nil {
		win.SetClosable(*wp.Closable)
	}
	if wp.Minimizable != nil {
		win.SetMinimizable(*wp.Minimizable)
	}
	if wp.Maximizable != nil {
		win.SetMaximizable(*wp.Maximizable)
	}
	if wp.MinWidth != nil {
		win.SetMinWidth(int(*wp.MinWidth))
	}
	if wp.MinHeight != nil {
		win.SetMinHeight(int(*wp.MinHeight))
	}
	if wp.MaxWidth != nil {
		win.SetMaxWidth(int(*wp.MaxWidth))
	}
	if wp.MaxHeight != nil {
		win.SetMaxHeight(int(*wp.MaxHeight))
	}
	if wp.TitleBarColor != nil {
		win.SetTitleBarColor(convertColor(wp.TitleBarColor))
	}
	if wp.TitleBarTextColor != nil {
		win.SetTitleBarTextColor(convertColor(wp.TitleBarTextColor))
	}
	if wp.ActiveBorderColor != nil {
		win.SetActiveBorderColor(convertColor(wp.ActiveBorderColor))
	}
	if wp.InactiveBorderColor != nil {
		win.SetInactiveBorderColor(convertColor(wp.InactiveBorderColor))
	}

	return nil
}

// WindowManager widget

func (s *WidgetService) createWindowManager(props *pb.WidgetProperties) *server.WindowManagerPrimitive {
	wm := server.NewWindowManagerPrimitive()
	if props != nil {
		if props.Title != nil {
			wm.SetTitle(*props.Title)
		}
		s.applyWindowManagerProperties(wm, props)
	}
	return wm
}

func (s *WidgetService) applyWindowManagerProperties(wm *server.WindowManagerPrimitive, props *pb.WidgetProperties) error {
	if props.TypeProperties == nil {
		return nil
	}

	wmProps, ok := props.TypeProperties.(*pb.WidgetProperties_WindowManager)
	if !ok || wmProps.WindowManager == nil {
		return nil
	}

	if wmProps.WindowManager.BackgroundColor != nil {
		wm.SetBgColor(convertColor(wmProps.WindowManager.BackgroundColor))
	}

	return nil
}

// colorToString converts a proto Color to a tview color string
func colorToString(c *pb.Color) string {
	if c == nil {
		return "white"
	}
	switch v := c.Color.(type) {
	case *pb.Color_Name:
		return v.Name
	case *pb.Color_Index:
		return fmt.Sprintf("color%d", v.Index)
	case *pb.Color_Rgb:
		return fmt.Sprintf("#%02x%02x%02x", v.Rgb.R, v.Rgb.G, v.Rgb.B)
	}
	return "white"
}
