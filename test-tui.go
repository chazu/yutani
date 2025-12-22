package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	
	textView := tview.NewTextView().
		SetText("[yellow::b]Test TUI[-::-]\n\n" +
			"If you can see this, the TUI works!\n\n" +
			"Press [red::b]Ctrl+C[-::-] to exit").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	textView.SetBorder(true).SetTitle(" Test ")
	
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		return event
	})
	
	if err := app.SetRoot(textView, true).Run(); err != nil {
		panic(err)
	}
}

