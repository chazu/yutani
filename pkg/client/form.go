package client

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
)

// Form represents a form widget.
type Form struct {
	baseWidget
}

// FormBuilder provides a fluent interface for creating form widgets.
type FormBuilder struct {
	client *Client
	props  *pb.WidgetProperties
}

// NewForm creates a new form builder.
func (c *Client) NewForm() *FormBuilder {
	return &FormBuilder{
		client: c,
		props:  &pb.WidgetProperties{},
	}
}

// Title sets the form title.
func (b *FormBuilder) Title(title string) *FormBuilder {
	b.props.Title = &title
	return b
}

// Border sets whether the form has a border.
func (b *FormBuilder) Border(border bool) *FormBuilder {
	b.props.Border = &border
	return b
}

// Build creates the form widget on the server.
func (b *FormBuilder) Build() (*Form, error) {
	resp, err := b.client.widgetClient.CreateWidget(b.client.ctx, &pb.CreateWidgetRequest{
		SessionId:  b.client.sessionID,
		Type:       pb.WidgetType_WIDGET_FORM,
		Properties: b.props,
	})
	if err != nil {
		return nil, err
	}

	form := &Form{
		baseWidget: baseWidget{
			client:   b.client,
			widgetID: resp.WidgetId,
		},
	}

	b.client.registerWidget(form)
	return form, nil
}

// AddInputField adds an input field to the form.
func (f *Form) AddInputField(label string, fieldWidth int, initialValue string) (int, error) {
	return f.addField(label, pb.FormFieldType_FORM_FIELD_INPUT, fieldWidth, &initialValue, nil)
}

// AddPasswordField adds a password field to the form.
func (f *Form) AddPasswordField(label string, fieldWidth int) (int, error) {
	return f.addField(label, pb.FormFieldType_FORM_FIELD_PASSWORD, fieldWidth, nil, nil)
}

// AddCheckbox adds a checkbox to the form.
func (f *Form) AddCheckbox(label string, initialValue bool) (int, error) {
	val := "false"
	if initialValue {
		val = "true"
	}
	return f.addField(label, pb.FormFieldType_FORM_FIELD_CHECKBOX, 0, &val, nil)
}

// AddDropdown adds a dropdown to the form.
func (f *Form) AddDropdown(label string, options []string, initialValue string) (int, error) {
	return f.addField(label, pb.FormFieldType_FORM_FIELD_DROPDOWN, 0, &initialValue, options)
}

// addField is a helper for adding fields.
func (f *Form) addField(label string, fieldType pb.FormFieldType, fieldWidth int, initialValue *string, options []string) (int, error) {
	req := &pb.FormAddFieldRequest{
		SessionId: f.client.sessionID,
		WidgetId:  f.widgetID,
		Label:     label,
		FieldType: fieldType,
	}

	if fieldWidth > 0 {
		req.FieldWidth = int32Ptr(int32(fieldWidth))
	}
	if initialValue != nil {
		req.InitialValue = initialValue
	}
	if options != nil {
		req.DropdownOptions = options
	}

	resp, err := f.client.formClient.AddField(f.client.ctx, req)
	if err != nil {
		return -1, err
	}
	return int(resp.FieldIndex), nil
}

// AddButton adds a button to the form.
func (f *Form) AddButton(label string) error {
	_, err := f.client.formClient.AddButton(f.client.ctx, &pb.FormAddButtonRequest{
		SessionId: f.client.sessionID,
		WidgetId:  f.widgetID,
		Label:     label,
	})
	return err
}

// GetFieldValue gets a field's current value.
func (f *Form) GetFieldValue(fieldIndex int) (string, error) {
	resp, err := f.client.formClient.GetFieldValue(f.client.ctx, &pb.FormGetFieldValueRequest{
		SessionId:  f.client.sessionID,
		WidgetId:   f.widgetID,
		FieldIndex: int32(fieldIndex),
	})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetFieldValue sets a field's value.
func (f *Form) SetFieldValue(fieldIndex int, value string) error {
	_, err := f.client.formClient.SetFieldValue(f.client.ctx, &pb.FormSetFieldValueRequest{
		SessionId:  f.client.sessionID,
		WidgetId:   f.widgetID,
		FieldIndex: int32(fieldIndex),
		Value:      value,
	})
	return err
}

// Clear clears all form fields.
func (f *Form) Clear() error {
	_, err := f.client.formClient.Clear(f.client.ctx, &pb.FormClearRequest{
		SessionId: f.client.sessionID,
		WidgetId:  f.widgetID,
	})
	return err
}

// GetItemCount returns the number of form items.
func (f *Form) GetItemCount() (int, error) {
	resp, err := f.client.formClient.GetItemCount(f.client.ctx, &pb.FormGetItemCountRequest{
		SessionId: f.client.sessionID,
		WidgetId:  f.widgetID,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

