package generator

// KeyGenerator creates short keys for new URLs.
type KeyGenerator interface {
	NextKey() string
}
