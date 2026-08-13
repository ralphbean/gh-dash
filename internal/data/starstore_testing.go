package data

// NewStarStoreForTesting creates a StarStore backed by the given file path.
func NewStarStoreForTesting(filePath string) *StarStore {
	return &StarStore{
		entries:  make(map[string]struct{}),
		filePath: filePath,
	}
}

// OverrideStarStoreForTesting replaces the singleton StarStore with the
// given store. It returns a function that restores the original store.
func OverrideStarStoreForTesting(store *StarStore) func() {
	// Ensure the singleton is initialized so sync.Once has fired.
	GetStarStore()
	old := starStore
	starStore = store
	return func() { starStore = old }
}
