package sdd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/assets"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestInjectSkillDirectoryForSkillsWritesRoutedSubsetAndSharedFilesOnce(t *testing.T) {
	root := t.TempDir()
	ids := []model.SkillID{model.SkillSDDInit, model.SkillJudgmentDay}

	result, err := InjectSkillDirectoryForSkills(root, ids, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed {
		t.Fatal("first routed SDD injection changed = false, want true")
	}
	for _, path := range []string{
		filepath.Join(root, "sdd-init", "SKILL.md"),
		filepath.Join(root, "judgment-day", "SKILL.md"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("routed SDD file %q: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "sdd-apply", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("unselected routed SDD skill exists: %v", err)
	}

	sharedFiles, err := assets.SharedSkillFileNames()
	if err != nil {
		t.Fatal(err)
	}
	for _, fileName := range sharedFiles {
		path := filepath.Join(root, "_shared", filepath.FromSlash(fileName))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("shared SDD file %q: %v", path, err)
		}
		if occurrences(result.Files, path) != 1 {
			t.Fatalf("shared SDD file %q appears %d times in result, want once", path, occurrences(result.Files, path))
		}
	}

	second, err := InjectSkillDirectoryForSkills(root, ids, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if second.Changed {
		t.Fatal("second routed SDD injection changed = true, want idempotent false")
	}
}

func occurrences(paths []string, path string) int {
	count := 0
	for _, current := range paths {
		if current == path {
			count++
		}
	}
	return count
}
