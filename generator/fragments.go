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

type fragment interface {
	fragmentKind() fragmentKind
}

type textFragment struct {
	Text string
}

func (t *textFragment) fragmentKind() fragmentKind {
	return textFragmentKind
}

type staticFragment struct {
	Text string
}

func (s *staticFragment) fragmentKind() fragmentKind {
	return staticFragmentKind
}

type dynamicFragment struct {
	Text string
}

func (s *dynamicFragment) fragmentKind() fragmentKind {
	return dynamicFragmentKind
}

type specialStringFragment struct {
	Name string
	Text string
}

func (s *specialStringFragment) fragmentKind() fragmentKind {
	return specialStringFragmentKind
}

type specialTableFragment struct {
	Name  string
	Table *table
}

func (s *specialTableFragment) fragmentKind() fragmentKind {
	return specialTableFragmentKind
}

type tableFragment struct {
	Table *table
}

func (s *tableFragment) fragmentKind() fragmentKind {
	return tableFragmentKind
}
