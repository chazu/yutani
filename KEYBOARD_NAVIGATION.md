# Keyboard Navigation in Yutani

## Overview

Yutani uses tview for its TUI, which provides standard keyboard navigation patterns.

## 🎮 Navigation Keys

### Global Keys

- **Ctrl+C** - Exit the server gracefully
- **Tab** - Move focus to next widget
- **Shift+Tab** (Backtab) - Move focus to previous widget

### List Widget

When a List widget has focus:
- **↑ / k** - Move selection up
- **↓ / j** - Move selection down
- **Home / g** - Jump to first item
- **End / G** - Jump to last item
- **Enter** - Select/activate item
- **Page Up** - Scroll up one page
- **Page Down** - Scroll down one page

### Table Widget

When a Table widget has focus:
- **↑ / k** - Move selection up
- **↓ / j** - Move selection down
- **← / h** - Move selection left
- **→ / l** - Move selection right
- **Home** - Jump to first row
- **End** - Jump to last row
- **Enter** - Select/activate cell
- **Page Up** - Scroll up one page
- **Page Down** - Scroll down one page

### Form Widget

When a Form widget has focus:
- **Tab** - Move to next field
- **Shift+Tab** - Move to previous field
- **Enter** - Submit form (when on button)
- **Space** - Toggle checkbox
- **Type** - Enter text in input fields

### Tree Widget

When a Tree widget has focus:
- **↑ / k** - Move selection up
- **↓ / j** - Move selection down
- **Enter / Space** - Expand/collapse node
- **Home / g** - Jump to root
- **End / G** - Jump to last visible node

## 🔍 Focus Management

### What is Focus?

In a TUI, only ONE widget can have "focus" at a time. The focused widget:
- Receives keyboard input
- Usually has a highlighted border or selection
- Can be interacted with

### Setting Focus

When widgets are added to layouts, you can specify which one gets focus:

```go
// In FlexAddItem
layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
    SessionId:  sessionID,
    FlexId:     containerID,
    ItemId:     widgetID,
    Proportion: 1,
    Focus:      true,  // This widget gets focus
})
```

### Changing Focus Programmatically

Use the SetFocus RPC:

```go
widgetClient.SetFocus(ctx, &pb.SetFocusRequest{
    SessionId: sessionID,
    WidgetId:  targetWidgetID,
})
```

### Checking Current Focus

Use the GetFocus RPC:

```go
resp, err := widgetClient.GetFocus(ctx, &pb.GetFocusRequest{
    SessionId: sessionID,
})
// resp.WidgetId contains the currently focused widget
```

## 🎯 Phase 4 Demo Navigation

After running the demo (`./bin/phase4-demo`), the List widget in the top-left has focus.

### Try These Actions:

1. **Navigate the List**
   - Press **↓** or **j** to move down through the phases
   - Press **↑** or **k** to move up
   - Press **Enter** to select an item

2. **Switch to Table** (future feature)
   - Press **Tab** to move focus to the Table widget
   - Use arrow keys to navigate cells

3. **Switch to Form** (future feature)
   - Press **Tab** multiple times to reach the Form
   - Use **Tab** within the form to move between fields

4. **Switch to Tree** (future feature)
   - Press **Tab** to reach the Tree widget
   - Press **Enter** to expand/collapse nodes

## 🐛 Troubleshooting

### Problem: Can't interact with any widgets

**Cause:** No widget has focus.

**Solution:** Make sure at least one widget is added with `Focus: true`:

```go
layoutClient.FlexAddItem(ctx, &pb.FlexAddItemRequest{
    // ...
    Focus: true,  // At least ONE widget needs this
})
```

### Problem: Tab doesn't switch between widgets

**Cause:** Tab navigation between widgets in a Flex/Grid layout requires additional setup.

**Status:** This is a known limitation. Currently, you need to:
1. Use SetFocus RPC to change focus programmatically, OR
2. Implement custom Tab handling in the client

**Future:** Tab navigation will be automatic in Phase 5.

### Problem: Arrow keys don't work

**Cause:** The widget doesn't have focus, or the widget type doesn't support arrow keys.

**Solution:** 
1. Make sure the widget has focus (check with GetFocus)
2. Verify the widget type supports keyboard input (List, Table, Form, Tree do; TextView doesn't)

## 📝 Widget Interaction Support

| Widget Type | Keyboard Input | Mouse Input | Focusable |
|-------------|---------------|-------------|-----------|
| TextView    | ❌ No         | ✅ Scroll   | ⚠️ Limited |
| InputField  | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| Button      | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| Checkbox    | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| List        | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| Table       | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| Form        | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| Tree        | ✅ Yes        | ✅ Yes      | ✅ Yes    |
| Flex        | ❌ Container  | ❌ Container| ❌ No     |
| Grid        | ❌ Container  | ❌ Container| ❌ No     |
| Pages       | ❌ Container  | ❌ Container| ❌ No     |

## 🎨 Visual Focus Indicators

When a widget has focus, tview typically shows:
- **Highlighted border** (if the widget has a border)
- **Different color** for the selection
- **Cursor** (for input fields)

The exact appearance depends on your terminal's color scheme.

## 🚀 Best Practices

1. **Always set focus on at least one widget** when creating a layout
2. **Use Tab/Shift+Tab** for navigation between widgets
3. **Use arrow keys** for navigation within widgets
4. **Provide visual feedback** by setting borders and titles on focusable widgets
5. **Test keyboard navigation** before deploying to users

## 📚 Next Steps

- See `RUN_SERVER.md` for how to run the server
- See `QUICKSTART.md` for a complete walkthrough
- See `FINAL_FIX_SUMMARY.md` for technical details

