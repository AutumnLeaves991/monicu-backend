package discord

type Uint64Set map[uint64]struct{}

func NewUint64Set(s []uint64) Uint64Set {
	set := make(Uint64Set, len(s))
	for _, i := range s {
		set[i] = struct{}{}
	}
	return set
}

func (s Uint64Set) Contains(i uint64) bool {
	_, exists := s[i]
	return exists
}
