package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

const IndexFileName = "index.html"

func SaveHTMLtoFile(htmlString string) error {
	file, err := os.Create(IndexFileName)
	if err != nil {
		return fmt.Errorf("ошибка при создании файла: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(htmlString)
	if err != nil {
		return fmt.Errorf("ошибка при записи в файл: %w", err)
	}

	return nil
}

func FileToLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LinesFromReader(f)
}

func LinesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func InsertStringToFile(filePath, str string, index int) error {
	lines, err := FileToLines(filePath)
	if err != nil {
		return err
	}

	fileContent := ""
	for i, line := range lines {
		if i == index {
			fileContent += str
		}
		fileContent += line
		fileContent += "\n"
	}

	return os.WriteFile(filePath, []byte(fileContent), 0644)
}

func EnsureDir(dirName string) error {
	err := os.Mkdir(dirName, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("ошибка при создании папки: %w", err)
	}
	return nil
}
