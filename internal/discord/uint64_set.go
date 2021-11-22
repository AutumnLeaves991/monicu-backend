package discord

// Uint64Set is a simple map-based set of unique uint64 values.
type Uint64Set struct {
	backingMap map[uint64]struct{}
}

// NewUint64Set creates a new Uint64Set from the specified array of uint64.
func NewUint64Set(s []uint64) *Uint64Set {
	set := &Uint64Set{make(map[uint64]struct{}, len(s))}
	for _, i := range s {
		set.backingMap[i] = struct{}{}
	}
	return set
}

// Contains checks if this Uint64Set contains the specified uint64.
func (s *Uint64Set) Contains(i uint64) bool {
	_, exists := s.backingMap[i]
	return exists
}

// Values return values contained by this Uint64Set as array.
func (s *Uint64Set) Values() []uint64 {
	v := make([]uint64, 0, len(s.backingMap))
	for k := range s.backingMap {
		v = append(v, k)
	}

	return v
}
