package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func GetFileName(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return info.Name()
}

func GetParentDir(path string) string {
	pathComponents := strings.Split(path, "/")
	return strings.Join(pathComponents[:len(pathComponents)-1], "/")
}

func CopyAll(src string, dstDir string, exclude []string, ignoreEmptyFiles bool) error {
	if !FileExists(src) {
		return fmt.Errorf("source file/directory %s does not exist", src)
	}

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dstDir, err)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file/directory %s: %w", src, err)
	}
	if info.IsDir() {
		fmt.Println("Copying dependency directory", src, "to", dstDir)
		children, err := os.ReadDir(src)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", src, err)
		}

		for _, child := range children {
			if contains(exclude, child.Name()) {
				fmt.Println("Skipping excluded file", child.Name())
				continue
			}

			var dst string
			if child.IsDir() {
				dst = dstDir + "/" + child.Name()
			} else {
				dst = dstDir
			}

			if err := CopyAll(src+"/"+child.Name(), dst, exclude, ignoreEmptyFiles); err != nil {
				return err
			}
		}
	} else {
		dstFile := dstDir + "/" + info.Name()
		fmt.Println("Copying file", src, "to", dstFile)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", src, err)
		}
		if ignoreEmptyFiles && len(data) == 0 {
			fmt.Println("Skipping empty file", src)
			return nil
		}
		if err := os.WriteFile(dstFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", dstFile, err)
		}
	}

	return nil
}

func contains(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

func RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		if errb.Len() != 0 {
			return "", fmt.Errorf("%s", errb.String())
		} else {
			return "", fmt.Errorf("%s", outb.String())
		}
	} else {
		return outb.String(), nil
	}
}
