package agent

import (
	"encoding/json"
	"fmt"
	"testing"
)

// The browser render pin consumes text from the real status handler.
func TestBrandStatusRenderFixture(t *testing.T) {
	a := &Agent{}
	data, err := json.Marshal(map[string]string{"en": a.handleStatus("en"), "zh": a.handleStatus("zh"), "id": a.handleStatus("id")})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("BRAND_STATUS_JSON:%s\n", data)
}
