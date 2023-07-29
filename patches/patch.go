// Package patches contains patches for Lithium EPUB Reader 0.24.1.
package patches

import (
	"io"

	"github.com/pgaskin/lithiumpatch/patches/internal/patchdef"

	_ "github.com/pgaskin/lithiumpatch/patches/internal/cleanupunused"
	_ "github.com/pgaskin/lithiumpatch/patches/internal/nofeedback"
	_ "github.com/pgaskin/lithiumpatch/patches/internal/signature"
	_ "github.com/pgaskin/lithiumpatch/patches/internal/version"
)

type Patch interface {
	Name() string
	Apply(apk string, diffwriter io.Writer) error
}

func Get() []Patch {
	p := patchdef.Patches()
	v := make([]Patch, len(p))
	for i, x := range p {
		v[i] = x
	}
	return v
}
