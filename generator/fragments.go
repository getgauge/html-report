/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package generator

const (
	textFragmentKind fragmentKind = iota
	staticFragmentKind
	dynamicFragmentKind
	specialStringFragmentKind
	specialTableFragmentKind
	tableFragmentKind
	multilineFragmentKind
)

type fragmentKind int

type fragment struct {
	FragmentKind fragmentKind
	Text         string
	Name         string
	Table        *table
	FileName     string
}
