package static

import "testing"

func TestSPAHandler_DevNil(t *testing.T) {
	// In dev mode web/ only contains .gitkeep, so SPAHandler returns nil.
	h := SPAHandler()
	if h != nil {
		t.Fatal("expected nil SPAHandler when web/ has no index.html")
	}
}
