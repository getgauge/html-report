/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package mdgen

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// TestParseGoldenFiles is the highest-leverage test on the renderer: every
// fixture is generated end-to-end, every produced .md is parsed with a real
// CommonMark/GFM parser, and every link/image target must resolve to a file
// on disk. This catches "looks fine, breaks in real renderers" bugs that
// substring assertions miss.
func TestParseGoldenFiles(t *testing.T) {
	saved := projectRoot
	projectRoot = fixtureProjectRoot
	t.Cleanup(func() { projectRoot = saved })

	gm := goldmark.New(goldmark.WithExtensions(extension.GFM))

	for _, f := range allFixtures() {
		f := f
		t.Run(f.name, func(t *testing.T) {
			if f.setup != nil {
				f.setup(t)
			}
			suite := f.build()
			tmp := t.TempDir()
			if err := GenerateReports(suite, tmp); err != nil {
				t.Fatalf("GenerateReports: %v", err)
			}

			// Collect every .md in the tree, then parse each.
			var mdFiles []string
			err := filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(path, ".md") {
					mdFiles = append(mdFiles, path)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walk: %v", err)
			}
			if len(mdFiles) == 0 {
				t.Fatalf("no .md files written under %s", tmp)
			}

			for _, mdPath := range mdFiles {
				assertMarkdownIsValid(t, gm, tmp, mdPath)
			}
		})
	}
}

// assertMarkdownIsValid parses the file at mdPath and walks the AST, asserting
// the file has at least one heading and that every internal link/image
// target points to a real file under reportsDir.
func assertMarkdownIsValid(t *testing.T, gm goldmark.Markdown, reportsDir, mdPath string) {
	t.Helper()
	srcBytes, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read %s: %v", mdPath, err)
	}

	doc := gm.Parser().Parse(text.NewReader(srcBytes))
	rel, _ := filepath.Rel(reportsDir, mdPath)

	var (
		headingLevels []int
		linkTargets   []string
		imageTargets  []string
	)
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Heading:
			headingLevels = append(headingLevels, v.Level)
		case *ast.Link:
			linkTargets = append(linkTargets, string(v.Destination))
		case *ast.Image:
			imageTargets = append(imageTargets, string(v.Destination))
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", mdPath, err)
	}

	if len(headingLevels) == 0 {
		t.Errorf("%s: no headings found", rel)
	}

	mdDir := filepath.Dir(mdPath)
	for _, link := range linkTargets {
		if isExternal(link) {
			continue
		}
		assertResolves(t, mdDir, link, rel)
	}
	for _, img := range imageTargets {
		if isExternal(img) {
			continue
		}
		assertResolves(t, mdDir, img, rel)
	}
}

// isExternal returns true for hrefs we shouldn't validate against the file
// tree (http://, https://, mailto:, fragment-only).
func isExternal(href string) bool {
	if strings.HasPrefix(href, "#") {
		return true
	}
	u, err := url.Parse(href)
	if err != nil {
		return false
	}
	return u.Scheme != ""
}

// assertResolves checks that href (relative to mdDir) points to an existing
// file. Fragments are stripped since our renderer never emits anchors.
func assertResolves(t *testing.T, mdDir, href, sourceRel string) {
	t.Helper()
	clean := href
	if i := strings.IndexByte(clean, '#'); i >= 0 {
		clean = clean[:i]
	}
	if clean == "" {
		return
	}
	full := filepath.Join(mdDir, clean)
	if _, err := os.Stat(full); err != nil {
		t.Errorf("%s references missing target %q (resolved to %s)", sourceRel, href, full)
	}
}

// TestParseDoesNotEmitHTMLBlocksOutsideDetails is a structural check: the
// only raw HTML our renderer emits is <details>/<summary>. If a renderer
// change accidentally introduces other HTML (e.g. an unescaped < from a
// stack trace), this catches it.
func TestParseDoesNotEmitUnexpectedHTML(t *testing.T) {
	saved := projectRoot
	projectRoot = fixtureProjectRoot
	t.Cleanup(func() { projectRoot = saved })

	gm := goldmark.New(goldmark.WithExtensions(extension.GFM))

	allowedTags := map[string]bool{
		"details": true,
		"summary": true,
	}

	for _, f := range allFixtures() {
		f := f
		t.Run(f.name, func(t *testing.T) {
			if f.setup != nil {
				f.setup(t)
			}
			suite := f.build()
			tmp := t.TempDir()
			if err := GenerateReports(suite, tmp); err != nil {
				t.Fatalf("GenerateReports: %v", err)
			}

			err := filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() || !strings.HasSuffix(path, ".md") {
					return nil
				}
				src, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				doc := gm.Parser().Parse(text.NewReader(src))
				return ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
					if !entering {
						return ast.WalkContinue, nil
					}
					switch v := n.(type) {
					case *ast.HTMLBlock:
						tag := extractTagName(string(v.Lines().Value(src)))
						if tag != "" && !allowedTags[tag] {
							t.Errorf("%s: unexpected HTML block tag %q", path, tag)
						}
					case *ast.RawHTML:
						tag := extractTagName(string(v.Segments.Value(src)))
						if tag != "" && !allowedTags[tag] {
							t.Errorf("%s: unexpected raw HTML tag %q", path, tag)
						}
					}
					return ast.WalkContinue, nil
				})
			})
			if err != nil {
				t.Fatalf("walk: %v", err)
			}
		})
	}
}

// extractTagName pulls the lowercase tag name from a raw HTML snippet — e.g.
// "<details>...</details>" → "details". Returns "" if the snippet is not an
// HTML opening tag.
func extractTagName(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "<") {
		return ""
	}
	s = s[1:]
	if strings.HasPrefix(s, "/") {
		s = s[1:]
	}
	end := len(s)
	for i, r := range s {
		if r == ' ' || r == '>' || r == '/' {
			end = i
			break
		}
	}
	return strings.ToLower(s[:end])
}
