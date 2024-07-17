package celedgewrapper

import (
	"testing"
	"time"
)

func TestCache_Add(t *testing.T) {
	cache := NewCache("initial")
	maxDur := 3 * time.Second

	// 测试添加新的值
	cache.Add("Hello", maxDur)
	if cache.Len() != 2 {
		t.Errorf("expected 2 items in cache, got %d", cache.Len())
	}

	// 等待一秒，添加相同的值
	time.Sleep(1 * time.Second)
	cache.Add("Hello", maxDur)
	if cache.Len() != 2 {
		t.Errorf("expected 2 items in cache after adding same value, got %d", cache.Len())
	}

	t.Log(cache.Len())

	// 检查时间戳更新
	firstTimestamp := cache.data[1].Timestamp
	t.Log(firstTimestamp)
	time.Sleep(1 * time.Second)
	cache.Add("Hello", maxDur)
	t.Log(cache.Len())
	if cache.data[1].Timestamp == firstTimestamp {
		t.Errorf("expected timestamp to be updated, but it was not")
	}

	// 添加不同的值
	cache.Add(42, maxDur)
	if cache.Len() != 3 {
		t.Errorf("expected 3 items in cache, got %d", cache.Len())
	}

	// 测试删除过期缓存值
	time.Sleep(3 * time.Second)
	cache.Add("New Value", maxDur)
	t.Log(cache.Len())

	// 测试锁定期间无法添加新值
	cache.Lock()
	cache.Add("Locked Value", maxDur)
	t.Log(cache.Len())

	// 解锁后添加新值
	cache.Unlock()
	cache.Add("Unlocked Value", maxDur)
	t.Log(cache.Len())
}
