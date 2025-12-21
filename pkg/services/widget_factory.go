package services

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
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
