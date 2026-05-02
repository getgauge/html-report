/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

// Package mdgen generate.go: filesystem-side of the report generation. Walks
// a SuiteResult, writes index.md + one .md per spec, and copies screenshots
// into images/. Pure rendering lives in render.go.
package mdgen

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/logger"
)

// indexFileName is the markdown report's entry point, matching the GitHub
// convention so a pushed report directory renders as a folder readme.
const indexFileName = "index.md"

// GenerateReports writes the full markdown report tree to reportsDir:
// index.md at the root, one .md per spec under specs/, and any screenshots
// referenced by failing steps into images/.
//
// When BeforeSuiteHookFailure is set, only the index page is rendered: the
// spec results never executed, so emitting empty per-spec pages would be
// misleading. After-suite failures don't have this constraint — the spec
// results are valid, so we render them as usual.
func GenerateReports(res *SuiteResult, reportsDir string) error {
	if err := os.MkdirAll(reportsDir, common.NewDirectoryPermissions); err != nil {
		return err
	}

	indexPath := filepath.Join(reportsDir, indexFileName)
	if err := writeRendered(indexPath, func(buf *bytes.Buffer) error {
		return RenderIndex(buf, res)
	}); err != nil {
		return err
	}

	if res.BeforeSuiteHookFailure == nil {
		if err := writeSpecPages(res, reportsDir); err != nil {
			return err
		}
		if env.ShouldUseNestedSpecs() {
			if err := writeNestedIndexPages(res, reportsDir); err != nil {
				return err
			}
		}
	}

	if err := copyScreenshotFiles(reportsDir); err != nil {
		// Screenshot copy failures are reported but don't abort the whole
		// report — the markdown is still useful without them.
		logger.Warnf("Failed to copy one or more screenshots: %s", err.Error())
	}
	return nil
}

// writeSpecPages fans the spec renders out across goroutines. The original
// HTML pipeline used a similar pattern; markdown is still I/O-bound on large
// suites so the fan-out earns its keep.
func writeSpecPages(res *SuiteResult, reportsDir string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(res.SpecResults))

	for _, s := range res.SpecResults {
		s := s
		outPath := filepath.Join(reportsDir, indexLinkHref(s))
		if err := os.MkdirAll(filepath.Dir(outPath), common.NewDirectoryPermissions); err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := writeRendered(outPath, func(buf *bytes.Buffer) error {
				return RenderSpec(buf, res, s)
			}); err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

// writeNestedIndexPages emits a per-directory index.md for each spec
// subdirectory, replicating the HTML report's nested-mode behavior. Each
// nested index lists only the specs at or below that directory.
func writeNestedIndexPages(res *SuiteResult, reportsDir string) error {
	dirs := nestedSpecDirs(res)
	for d := range dirs {
		nested := toNestedSuiteResult(d, res)
		outDir := filepath.Join(reportsDir, d)
		if err := os.MkdirAll(outDir, common.NewDirectoryPermissions); err != nil {
			return err
		}
		outPath := filepath.Join(outDir, indexFileName)
		if err := writeRendered(outPath, func(buf *bytes.Buffer) error {
			return RenderIndex(buf, nested)
		}); err != nil {
			return err
		}
	}
	return nil
}

// nestedSpecDirs returns every distinct spec subdirectory (relative to the
// project root) that contains at least one spec result.
func nestedSpecDirs(res *SuiteResult) map[string]struct{} {
	dirs := make(map[string]struct{})
	for _, s := range res.SpecResults {
		rel, err := filepath.Rel(projectRoot, filepath.Dir(s.FileName))
		if err != nil || rel == "." || rel == "" {
			continue
		}
		// Walk up the directory chain so a spec under a/b/c registers a, a/b,
		// and a/b/c. This matches the HTML report's per-level index layout.
		cur := rel
		for cur != "." && cur != "" {
			dirs[cur] = struct{}{}
			parent := filepath.Dir(cur)
			if parent == cur {
				break
			}
			cur = parent
		}
	}
	return dirs
}

// writeRendered runs fn against a buffer, then writes the buffer to path.
// Buffering up-front means a render failure doesn't leave a half-written
// file on disk.
func writeRendered(path string, fn func(*bytes.Buffer) error) error {
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), common.NewFilePermissions)
}

// copyScreenshotFiles copies every screenshot recorded during transform into
// reportsDir/images/. Lifted from the HTML generator with one change: missing
// source files are warned-about rather than fatal, since regenerated reports
// may run from a machine that no longer has the originals.
func copyScreenshotFiles(reportsDir string) error {
	if len(screenshotFiles) == 0 {
		return nil
	}
	imagesDir := filepath.Join(reportsDir, "images")
	if err := os.MkdirAll(imagesDir, common.NewDirectoryPermissions); err != nil {
		return err
	}
	src := os.Getenv(env.ScreenshotsDirName)
	for _, name := range screenshotFiles {
		srcfp := filepath.Join(src, name)
		dstfp := filepath.Join(imagesDir, filepath.Base(name))
		fileBytes, err := os.ReadFile(srcfp)
		if err != nil {
			logger.Warnf("Failed to read screenshot %s: %s", srcfp, err.Error())
			continue
		}
		if err := os.WriteFile(dstfp, fileBytes, common.NewFilePermissions); err != nil {
			logger.Warnf("Failed to write screenshot %s: %s", dstfp, err.Error())
		}
	}
	return nil
}
