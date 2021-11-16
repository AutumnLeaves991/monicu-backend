package discord

type uint64Set map[uint64]struct{}

func newUint64Set(s []uint64) uint64Set {
	set := make(uint64Set, len(s))
	for _, i := range s {
		set[i] = struct{}{}
	}
	return set
}

func (s uint64Set) Contains(i uint64) bool {
	_, exists := s[i]
	return exists
}
