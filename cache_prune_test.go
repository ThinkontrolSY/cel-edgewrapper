package celedgewrapper

import (
	"testing"
	"time"
)

func TestCache_Pruning(t *testing.T) {
	maxDur := 100 * time.Millisecond

	t.Run("Prune all old but keep 2", func(t *testing.T) {
		c := NewCache("init") // 1 item
		// Add 4 more items
		for i := 0; i < 4; i++ {
			c.Add(i, maxDur)
			time.Sleep(10 * time.Millisecond)
		}
		// Total 5 items. Timestamps are close.
		if c.Len() != 5 {
			t.Fatalf("expected 5 items, got %d", c.Len())
		}

		// Sleep longer than maxDur
		time.Sleep(200 * time.Millisecond)

		// All 5 items are now old.
		// Add a new item.
		// Pruning should happen.
		// We should keep at least 2 of the old items.
		// So 2 old + 1 new = 3 items.
		c.Add("new", maxDur)

		if c.Len() != 3 {
			t.Errorf("expected 3 items (2 old kept + 1 new), got %d", c.Len())
		}
	})

	t.Run("Prune some old", func(t *testing.T) {
		c := NewCache("init")
		// Add items with delay
		// Item 0: T=0
		// Sleep 150ms
		// Item 1: T=150
		// Item 2: T=150
		// MaxDur = 100ms.
		// At T=200. Item 0 is old (200ms > 100ms). Item 1, 2 are new (50ms < 100ms).

		time.Sleep(150 * time.Millisecond)
		c.Add("item1", maxDur)
		c.Add("item2", maxDur)

		// Current state: [init(old), item1(new), item2(new)]
		// Wait a bit but not enough to expire item1, item2
		time.Sleep(10 * time.Millisecond)

		// Add item3.
		// Pruning:
		// init is old.
		// item1, item2 are new.
		// We have 2 new items.
		// So we can drop init.
		// Remaining: item1, item2.
		// Add item3.
		// Result: item1, item2, item3. Size 3.

		c.Add("item3", maxDur)

		if c.Len() != 3 {
			t.Errorf("expected 3 items, got %d", c.Len())
		}

		// Verify content
		if c.data[0].Var != "item1" {
			t.Errorf("expected first item to be item1, got %v", c.data[0].Var)
		}
	})

	t.Run("Keep all if few", func(t *testing.T) {
		c := NewCache("init")
		time.Sleep(200 * time.Millisecond)
		// 1 item, old.
		// Keep at least 2 -> keep 1.
		c.Add("new", maxDur)
		if c.Len() != 2 {
			t.Errorf("expected 2 items, got %d", c.Len())
		}
	})
}
