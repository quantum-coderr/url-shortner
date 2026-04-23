package generator

import "sync/atomic"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62Generator turns increasing integer IDs into Base62 keys.
type Base62Generator struct {
	counter uint64
}

func NewBase62Generator(start uint64) *Base62Generator {
	return &Base62Generator{counter: start}
}

func (g *Base62Generator) NextKey() string {
	id := atomic.AddUint64(&g.counter, 1)
	return encodeBase62(id)
}

func encodeBase62(n uint64) string {
	if n == 0 {
		return "0"
	}

	var out [11]byte
	i := len(out)
	for n > 0 {
		i--
		out[i] = alphabet[n%62]
		n /= 62
	}

	return string(out[i:])
}
