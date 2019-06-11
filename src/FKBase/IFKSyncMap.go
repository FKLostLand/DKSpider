package FKBase

/*
	go 自带的Sync.Map，但增加了Len对象，可以便利获得Map大小
*/
// Map is a concurrent map with loads, stores, and deletes.
// It is safe for multiple goroutines to call a Map's methods concurrently.
type SyncMap interface {
	// Load returns the value stored in the map for a key, or nil if no
	// value is present.
	// The ok result indicates whether value was found in the map.
	Load(key interface{}) (value interface{}, ok bool)
	// Store sets the value for a key.
	Store(key, value interface{})
	// LoadOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored.
	LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
	// Range calls f sequentially for each key and value present in the map.
	// If f returns false, range stops the iteration.
	Range(f func(key, value interface{}) bool)
	// Random returns a pair kv randomly.
	// If exist=false, no kv data is exist.
	Random() (key, value interface{}, exist bool)
	// Delete deletes the value for a key.
	Delete(key interface{})
	// Clear clears all current data in the map.
	Clear()
	// Len returns the length of the map.
	Len() int
}

// AtomicMap creates a concurrent map with amortized-constant-time loads, stores, and deletes.
// It is safe for multiple goroutines to call a atomicMap's methods concurrently.
// From go v1.9 sync.Map.
// 对外接口
func CreateSyncMap() SyncMap {
	return new(atomicMap)
}
