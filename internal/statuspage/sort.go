package statuspage

// ComponentSort sorts components based on ID and name, the only
// guarantee being that it is self consistent across those fields.
type ComponentSort []Component

// Len is a part of sort.Interface, returning length of sort target
func (s ComponentSort) Len() int {
	return len(s)
}

// Swap is a part of sort.Interface, swapping items in sort target
func (s ComponentSort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is a part of sort.Interface, evaluating whether one item is less than another in sort target
func (s ComponentSort) Less(i, j int) bool {
	if s[i].ID != s[j].ID {
		return s[i].ID < s[j].ID
	} else {
		return s[i].Name < s[j].Name
	}
}
