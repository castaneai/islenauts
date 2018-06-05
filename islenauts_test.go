package islenauts

import (
	"context"
	"net/http"
	"testing"
)

func TestGetItems(t *testing.T) {
	c, err := NewClient(&http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	items, err := c.GetItems(ctx, "tags:抱き枕カバー")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		t.Logf("%+v", item)
	}
}
