package terminal

import (
	"github.com/rivo/tview"
)

func NewDialogComponents() (*tview.TextView, *tview.InputField) {
	dialogView := tview.NewTextView().
		SetText("AI Chat Terminal").
		SetDynamicColors(true).
		SetScrollable(true)
	dialogInput := tview.NewInputField().
		SetLabel("You: ")

	return dialogView, dialogInput
}
