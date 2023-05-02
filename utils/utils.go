package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CopyAll(src string, dst string, exclude []string, ignoreEmptyFiles bool) error {
	if !FileExists(src) {
		return fmt.Errorf("source file/directory %s does not exist", src)
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file/directory %s: %w", src, err)
	}
	if info.IsDir() {
		fmt.Println("Copying dependency directory", src, "to", dst)
		children, err := os.ReadDir(src)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %w", src, err)
		}

		if err := os.MkdirAll(dst, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
		}

		for _, child := range children {
			if contains(exclude, child.Name()) {
				fmt.Println("Skipping excluded file", child.Name())
				continue
			}
			if err := CopyAll(src+"/"+child.Name(), dst+"/"+child.Name(), exclude, ignoreEmptyFiles); err != nil {
				return err
			}
		}
	} else {
		fmt.Println("Copying file", src, "to", dst)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", src, err)
		}
		if ignoreEmptyFiles && len(data) == 0 {
			fmt.Println("Skipping empty file", src)
			return nil
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", dst, err)
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
