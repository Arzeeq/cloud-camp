package pool

type Pooler interface {
	Get() (string, error)
	// GetAll return URLs of all alive and dead servers
	GetAll() []string
	// Enable returns true if server were marked as dead before
	Enable(string) bool
	// Disable returns true if server were marked as alive before
	Disable(string) bool
}
