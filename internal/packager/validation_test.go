// Copyright 2024 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package packager

import (
	"os"
	"path/filepath"
	"testing"
)

func createFile(t *testing.T, dir, name, content string) string {
    t.Helper()
    tmpFile := filepath.Join(dir, name)
    err := os.WriteFile(tmpFile, []byte(content), 0644)
    if err != nil {
        t.Fatalf("failed to create temp file: %v", err)
    }
    return tmpFile
}

func TestIsScript_WithShebang(t *testing.T) {
    tmpDir := t.TempDir()
    path := createFile(t, tmpDir, "testfile.sh", "#!/bin/sh\n echo Hello")
    isScript, shebang, err := isScript(path)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !isScript {
        t.Errorf("expected isScript to be true")
    }
    if shebang != "/bin/sh" {
        t.Errorf("expected shebang to be '/bin/sh', got '%s'", shebang)
    }
}

func TestIsScript_WithoutShebang(t *testing.T) {
    tmpDir := t.TempDir()
    path := createFile(t, tmpDir, "testfile.sh", "echo Hello")
    isScript, shebang, err := isScript(path)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if isScript {
        t.Errorf("expected isScript to be false")
    }
    if shebang != "echo" {
        t.Errorf("expected shebang to be 'echo Hello', got '%s'", shebang)
    }
}

func TestIsScript_EmptyFile(t *testing.T) {
    tmpDir := t.TempDir()
    path := createFile(t, tmpDir, "testfile.sh", "")
    isScript, shebang, err := isScript(path)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if isScript {
        t.Errorf("expected isScript to be false")
    }
    if shebang != "" {
        t.Errorf("expected shebang to be empty, got '%s'", shebang)
    }
}

func TestIsScript_FileNotFound(t *testing.T) {
    isScript, shebang, err := isScript("nonexistent.sh")
    if err == nil {
        t.Fatalf("expected error for nonexistent file")
    }
    if isScript {
        t.Errorf("expected isScript to be false")
    }
    if shebang != "" {
        t.Errorf("expected shebang to be empty, got '%s'", shebang)
    }
}


func createBinaryFile(t *testing.T, tmpDir string, content []byte) string {
    t.Helper()
    tmpFile := filepath.Join(tmpDir, "binaryfile")
    err := os.WriteFile(tmpFile, content, 0755)
    if err != nil {
        t.Fatalf("failed to create binary file: %v", err)
    }
    return tmpFile
}

func TestIsBinary_ValidELF(t *testing.T) {
    // ELF magic number: 0x7F 0x45 0x4C 0x46
    tmpDir := t.TempDir()
    path := createBinaryFile(t, tmpDir, []byte{0x7F, 0x45, 0x4C, 0x46, 0x00})
    isBin, name, err := isBinary(path)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !isBin {
        t.Errorf("expected isBinary to be true")
    }
    if name != path {
        t.Errorf("expected file name to be '%s', got '%s'", path, name)
    }
}

func TestIsBinary_NotBinary(t *testing.T) {
    tmpDir := t.TempDir()
    path := createBinaryFile(t, tmpDir, []byte("text"))
    isBin, _, err := isBinary(path)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if isBin {
        t.Errorf("expected isBinary to be false")
    }
}

func TestIsBinary_TooSmall(t *testing.T) {
    tmpDir := t.TempDir()
    path := createBinaryFile(t, tmpDir, []byte{0x01, 0x02})
    isBin, _, err := isBinary(path)
    if err == nil {
        t.Fatalf("expected error for small file")
    }
    if isBin {
        t.Errorf("expected isBinary to be false")
    }
}

func TestIsBinary_FileNotFound(t *testing.T) {
    isBin, _, err := isBinary("nonexistent.bin")
    if err == nil {
        t.Fatalf("expected error for nonexistent file")
    }
    if isBin {
        t.Errorf("expected isBinary to be false")
    }
}


func TestIsSymlink_ValidSymlink(t *testing.T) {
    tmpDir := t.TempDir()
    targetFile := filepath.Join(tmpDir, "target.txt")
    symlink := filepath.Join(tmpDir, "link.txt")

    err := os.WriteFile(targetFile, []byte("hello"), 0644)
    if err != nil {
        t.Fatalf("failed to create target file: %v", err)
    }

    err = os.Symlink(targetFile, symlink)
    if err != nil {
        t.Fatalf("failed to create symlink: %v", err)
    }

    isLink, resolved, err := isSymlink(symlink)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !isLink {
        t.Errorf("expected isSymlink to be true")
    }
    if resolved != targetFile {
        t.Errorf("expected resolved path to be '%s', got '%s'", targetFile, resolved)
    }
}

func TestIsSymlink_NotASymlink(t *testing.T) {
    tmpDir := t.TempDir()
    regularFile := filepath.Join(tmpDir, "file.txt")

    err := os.WriteFile(regularFile, []byte("hello"), 0644)
    if err != nil {
        t.Fatalf("failed to create regular file: %v", err)
    }

    isLink, resolved, err := isSymlink(regularFile)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if isLink {
        t.Errorf("expected isSymlink to be false")
    }
    if resolved != "" {
        t.Errorf("expected resolved path to be empty, got '%s'", resolved)
    }
}

func TestIsSymlink_BrokenSymlink(t *testing.T) {
    tmpDir := t.TempDir()
    brokenTarget := filepath.Join(tmpDir, "nonexistent.txt")
    symlink := filepath.Join(tmpDir, "brokenlink.txt")

    err := os.Symlink(brokenTarget, symlink)
    if err != nil {
        t.Fatalf("failed to create broken symlink: %v", err)
    }

    isLink, resolved, err := isSymlink(symlink)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !isLink {
        t.Errorf("expected isSymlink to be true")
    }
    if resolved != brokenTarget {
        t.Errorf("expected resolved path to be '%s', got '%s'", brokenTarget, resolved)
    }
}

func TestIsSymlink_NonExistentPath(t *testing.T) {
    isLink, resolved, err := isSymlink("/nonexistent/path")
    if err == nil {
        t.Fatalf("expected error for nonexistent path")
    }
    if isLink {
        t.Errorf("expected isSymlink to be false")
    }
    if resolved != "" {
        t.Errorf("expected resolved path to be empty, got '%s'", resolved)
    }
}


func TestValidateFileProperties_ValidExecutable(t *testing.T) {
    tmpDir := t.TempDir()
    execFile := filepath.Join(tmpDir, "exec.sh")

    err := os.WriteFile(execFile, []byte("echo Hello"), 0755)
    if err != nil {
        t.Fatalf("failed to create executable file: %v", err)
    }

    err = validateFileProperties(execFile)
    if err != nil {
        t.Errorf("expected no error for valid executable, got: %v", err)
    }
}

func TestValidateFileProperties_NotExecutable(t *testing.T) {
    tmpDir := t.TempDir()
    nonExecFile := filepath.Join(tmpDir, "nonexec.sh")

    err := os.WriteFile(nonExecFile, []byte("echo Hello"), 0644)
    if err != nil {
        t.Fatalf("failed to create non-executable file: %v", err)
    }

    err = validateFileProperties(nonExecFile)
    if err == nil {
        t.Errorf("expected error for non-executable file")
    }
}

func TestValidateFileProperties_NotRegularFile(t *testing.T) {
    tmpDir := t.TempDir()

    err := validateFileProperties(tmpDir)
    if err == nil {
        t.Errorf("expected error for directory (not a regular file)")
    }
}

func TestValidateFileProperties_NonExistent(t *testing.T) {
    err := validateFileProperties("/nonexistent/path")
    if err == nil {
        t.Errorf("expected error for nonexistent file")
    }
}


func TestDiscoverEntrypointPath_Binary(t *testing.T) {
    tmpDir := t.TempDir()
    bin := createBinaryFile(t, tmpDir, []byte{0x7F, 0x45, 0x4C, 0x46})

    path, err := discoverEntrypointPath(bin, tmpDir)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if path != bin {
        t.Errorf("expected path to be '%s', got '%s'", bin, path)
    }
}


func createSymlink(t *testing.T, target, link string) {
    t.Helper()
    err := os.Symlink(target, link)
    if err != nil {
        t.Fatalf("failed to create symlink: %v", err)
    }
}


func TestDiscoverEntrypointPath_SymlinkToBinary(t *testing.T) {
	tmpDir := t.TempDir()
	bin := createBinaryFile(t, tmpDir, []byte{0x7F, 0x45, 0x4C, 0x46})
	tmpPathToLink := filepath.Join(tmpDir, "link")
	createSymlink(t, "binaryfile", tmpPathToLink)

	path, err := discoverEntrypointPath(tmpPathToLink, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != bin {
		t.Errorf("expected path to be '%s', got '%s'", bin, path)
	}
}

func TestDiscoverEntrypointPath_ScriptWithShebang(t *testing.T) {
    tmpDir := t.TempDir()
    bin := createBinaryFile(t, tmpDir, []byte{0x7F, 0x45, 0x4C, 0x46})
    scriptPath := createFile(t, tmpDir, "testfile.sh", "#!/binaryfile\n")

    path, err := discoverEntrypointPath(scriptPath, tmpDir)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    expected := bin
    if path != expected {
        t.Errorf("expected path to be '%s', got '%s'", expected, path)
    }
}

func TestDiscoverEntrypointPath_ChainSymlinkScriptBinary(t *testing.T) {
	tmpDir := t.TempDir()
	bin := createBinaryFile(t, tmpDir, []byte{0x7F, 0x45, 0x4C, 0x46})
	_ = createFile(t, tmpDir, "testfile.sh", "#!/binaryfile\n")
	tmpPathToLink := filepath.Join(tmpDir, "linkfile")
	createSymlink(t, "testfile.sh", tmpPathToLink)

	path, err := discoverEntrypointPath(tmpPathToLink, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != bin {
		t.Errorf("expected path to be '%s', got '%s'", bin, path)
	}
}

func TestDiscoverEntrypointPath_UnknownFileType(t *testing.T) {
    tmpDir := t.TempDir()
    unknown := createFile(t, tmpDir, "unknown", "test text")

    _, err := discoverEntrypointPath(unknown, tmpDir)
    if err == nil {
        t.Errorf("expected error for unknown file type")
    }
}
