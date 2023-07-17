package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const gitignore = ".gitignore"

type IgnorePatternsMap map[string][]string

func isIgnored(path string, ignorePatternsMap IgnorePatternsMap) bool {
	dir := filepath.Dir(path)
	for {
		if ignorePatterns, found := ignorePatternsMap[dir]; found {
			for _, pattern := range ignorePatterns {
				matched, err := filepath.Match(pattern, filepath.Base(path))
				if err == nil && matched {
					return true
				}
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}
	return false
}

func parseGitIgnore(gitignorePath string) ([]string, error) {
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var ignorePatterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern != "" && !strings.HasPrefix(pattern, "#") {
			ignorePatterns = append(ignorePatterns, pattern)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignorePatterns, nil
}

func buildIgnorePatternsMap(rootPath string) (IgnorePatternsMap, error) {
	ignorePatternsMap := make(IgnorePatternsMap)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		gitignorePath := filepath.Join(path, gitignore)
		ignorePatterns, err := parseGitIgnore(gitignorePath)
		if err != nil {
			return err
		}

		ignorePatternsMap[path] = ignorePatterns
		return nil
	})

	return ignorePatternsMap, err
}

func printTree(path string, ignorePatternsMap IgnorePatternsMap) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if isIgnored(path, ignorePatternsMap) {
		return
	}
	fmt.Println(fileInfo.Name())
	filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if subpath != path && !isIgnored(subpath, ignorePatternsMap) {
			relativePath := strings.Replace(subpath, path, "", 1)
			elements := strings.Split(relativePath, string(os.PathSeparator))
			indentation := strings.Repeat("   ", len(elements)-1)

			fmt.Print(indentation, "├──", info.Name(), "\n")
		}
		return nil
	})
}

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	ignorePatternsMap, err := buildIgnorePatternsMap(currentDir)
	if err != nil {
		fmt.Println("Error building ignore patterns map:", err)
		return
	}

	fmt.Println("Directory tree for:", currentDir)
	printTree(currentDir, ignorePatternsMap)
}
