package godirwalk

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWalkParentsUpwardTraversal(t *testing.T) {
	dir := filepath.Join(scaffolingRoot, "test_walkparents_upward", "level1", "level2", "level3")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot create test directory: %s", err)
	}

	var visited []string
	err := WalkParents(dir, func(d string) (bool, error) {
		visited = append(visited, d)
		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(visited) < 4 {
		t.Fatalf("expected at least 4 visited directories, got %d", len(visited))
	}

	cleanDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("failed to resolve absolute path: %s", err)
	}

	if visited[0] != cleanDir {
		t.Errorf("expected first visited %q, got %q", cleanDir, visited[0])
	}
	if visited[1] != filepath.Dir(cleanDir) {
		t.Errorf("expected second visited %q, got %q", filepath.Dir(cleanDir), visited[1])
	}
	if visited[2] != filepath.Dir(filepath.Dir(cleanDir)) {
		t.Errorf("expected third visited %q, got %q", filepath.Dir(filepath.Dir(cleanDir)), visited[2])
	}

	for i := 1; i < len(visited); i++ {
		expectedParent := filepath.Dir(visited[i-1])
		if visited[i] != expectedParent {
			t.Errorf("step %d: expected parent %q, got %q", i, expectedParent, visited[i])
		}
	}

	last := visited[len(visited)-1]
	if filepath.Dir(last) != last {
		t.Errorf("expected last visited directory %q to be filesystem root", last)
	}
}

func TestWalkParentsStopCondition(t *testing.T) {
	repoDir := filepath.Join(scaffolingRoot, "test_walkparents_stop", "project_root")
	deepDir := filepath.Join(repoDir, "cmd", "subcommand")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("cannot create test directory: %s", err)
	}

	markerFile := filepath.Join(repoDir, ".git")
	if err := ioutil.WriteFile(markerFile, []byte("gitdir"), 0644); err != nil {
		t.Fatalf("cannot create marker file: %s", err)
	}

	var visited []string
	var foundDir string

	err := WalkParents(deepDir, func(dir string) (bool, error) {
		visited = append(visited, dir)
		target := filepath.Join(dir, ".git")
		if _, err := os.Stat(target); err == nil {
			foundDir = dir
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	absRepoDir, err := filepath.Abs(repoDir)
	if err != nil {
		t.Fatalf("failed to resolve absolute path: %s", err)
	}

	if foundDir != absRepoDir {
		t.Errorf("expected found directory %q, got %q", absRepoDir, foundDir)
	}

	if len(visited) != 3 {
		t.Fatalf("expected 3 visited directories, got %d: %v", len(visited), visited)
	}

	if visited[len(visited)-1] != absRepoDir {
		t.Errorf("expected last visited directory %q, got %q", absRepoDir, visited[len(visited)-1])
	}
}

func TestWalkParentsRootBoundaryTermination(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get cwd: %s", err)
	}

	root := cwd
	for {
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}

	var visited []string
	err = WalkParents(root, func(dir string) (bool, error) {
		visited = append(visited, dir)
		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(visited) != 1 {
		t.Fatalf("expected 1 visit for root, got %d", len(visited))
	}
	if visited[0] != root {
		t.Errorf("expected %q, got %q", root, visited[0])
	}
}

func TestWalkParentsRelativePathResolution(t *testing.T) {
	var visited []string
	err := WalkParents(".", func(dir string) (bool, error) {
		visited = append(visited, dir)
		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get cwd: %s", err)
	}

	if len(visited) == 0 {
		t.Fatal("expected at least one visited directory")
	}

	if visited[0] != cwd {
		t.Errorf("expected first visit %q, got %q", cwd, visited[0])
	}

	for i, d := range visited {
		if !filepath.IsAbs(d) {
			t.Errorf("visit %d (%q) is not absolute", i, d)
		}
	}
}

func TestWalkParentsAbsolutePathResolution(t *testing.T) {
	base := filepath.Join(scaffolingRoot, "test_walkparents_clean", "a", "b")
	if err := os.MkdirAll(base, 0755); err != nil {
		t.Fatalf("cannot create test directory: %s", err)
	}

	dirtyPath := filepath.Join(scaffolingRoot, "test_walkparents_clean", "a", "..", "a", "b", ".")
	cleanAbs, err := filepath.Abs(dirtyPath)
	if err != nil {
		t.Fatalf("cannot resolve absolute path: %s", err)
	}

	var first string
	err = WalkParents(dirtyPath, func(dir string) (bool, error) {
		first = dir
		return true, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if first != cleanAbs {
		t.Errorf("expected clean path %q, got %q", cleanAbs, first)
	}
}

func TestWalkParentsFileTarget(t *testing.T) {
	dir := filepath.Join(scaffolingRoot, "test_walkparents_file")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot create directory: %s", err)
	}

	filePath := filepath.Join(dir, "target.txt")
	if err := ioutil.WriteFile(filePath, []byte("sample"), 0644); err != nil {
		t.Fatalf("cannot create file: %s", err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("failed to resolve absolute path: %s", err)
	}

	var first string
	err = WalkParents(filePath, func(d string) (bool, error) {
		first = d
		return true, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if first != absDir {
		t.Errorf("expected first visited directory to be %q, got %q", absDir, first)
	}
}

func TestWalkParentsErrorPropagation(t *testing.T) {
	dir := filepath.Join(scaffolingRoot, "test_walkparents_err", "a", "b")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot create directory: %s", err)
	}

	errExpected := errors.New("abort traversal")
	count := 0
	err := WalkParents(dir, func(d string) (bool, error) {
		count++
		return false, errExpected
	})

	if err != errExpected {
		t.Errorf("expected error %v, got %v", errExpected, err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 visit, got %d", count)
	}
}

func TestWalkParentsNilCallback(t *testing.T) {
	err := WalkParents("/some/path", nil)
	if err == nil {
		t.Fatal("expected error when callback is nil, got nil")
	}
}

func TestWalkParentsNonExistentPath(t *testing.T) {
	fakePath := filepath.FromSlash("/nonexistent_synthetic_dir/nested/child")
	cleanFake, err := filepath.Abs(fakePath)
	if err != nil {
		t.Fatalf("cannot resolve abs: %s", err)
	}

	var visited []string
	err = WalkParents(fakePath, func(d string) (bool, error) {
		visited = append(visited, d)
		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(visited) == 0 {
		t.Fatal("expected non-empty visited slice")
	}
	if visited[0] != cleanFake {
		t.Errorf("expected first visit %q, got %q", cleanFake, visited[0])
	}
	last := visited[len(visited)-1]
	if filepath.Dir(last) != last {
		t.Errorf("expected last visited %q to be filesystem root", last)
	}
}

func TestWalkAncestorsAlias(t *testing.T) {
	dir := filepath.Join(scaffolingRoot, "test_walkancestors_alias", "a", "b")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot create directory: %s", err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("failed to resolve absolute path: %s", err)
	}

	var first string
	err = WalkAncestors(dir, func(d string) (bool, error) {
		first = d
		return true, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if first != absDir {
		t.Errorf("expected %q, got %q", absDir, first)
	}
}
