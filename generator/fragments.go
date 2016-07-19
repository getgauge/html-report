package generator

const (
	textFragmentKind fragmentKind = iota
	staticFragmentKind
	dynamicFragmentKind
	specialStringFragmentKind
	specialTableFragmentKind
	tableFragmentKind
)

type fragmentKind int

type fragment struct {
	FragmentKind fragmentKind
	Text         string
	Name         string
	Table        *table
}
