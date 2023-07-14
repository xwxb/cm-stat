package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	outputDir      = "output"
	inputDir       = "input"
	projectRoot    = "test_prj"
	codeFileSuffix = ".rs"
)

var (
	commentRegex  = regexp.MustCompile(`\/\/\s*(TODO|FIXME|todo|fixme)\((\w+)\):\s*(.+)`)
	todoCounts    = make(map[string]int)
	fixmeCounts   = make(map[string]int)
	topTodoBlamer = 5
)

func main() {
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Println("Failed to create output directory:", err)
		return
	}

	err = extractComments(filepath.Join(".", inputDir, projectRoot))
	if err != nil {
		fmt.Println("Failed to extract comments:", err)
		return
	}

	err = writeTodoCounts()
	if err != nil {
		fmt.Println("Failed to write TODO counts:", err)
		return
	}

	err = writeTopTodoBlamer()
	if err != nil {
		fmt.Println("Failed to write top TODO blamer:", err)
		return
	}

	fmt.Println("Comment extraction and analysis completed.")
}

func extractComments(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = extractComments(filepath.Join(dirPath, file.Name()))
			if err != nil {
				return err
			}
			continue
		}

		if strings.HasSuffix(file.Name(), codeFileSuffix) {
			err = processFile(filepath.Join(dirPath, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processFile(filePath string) error {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileContent := string(fileBytes)
	matches := commentRegex.FindAllStringSubmatch(fileContent, -1)
	for _, match := range matches {
		commentType := match[1]
		blamer := match[2]
		// comment := match[3]

		switch strings.ToUpper(commentType) {
		case "TODO":
			todoCounts[blamer]++
		case "FIXME":
			fixmeCounts[blamer]++
		case "todo":
			todoCounts[blamer]++
		case "fixme":
			fixmeCounts[blamer]++
		}
	}

	return nil
}

func writeTodoCounts() error {
	outputPath := filepath.Join(outputDir, "todo_counts.txt")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	outputFile.WriteString("TODO\n")
	writeCounts(outputFile, todoCounts)

	outputFile.WriteString("FIXME\n")
	writeCounts(outputFile, fixmeCounts)

	return nil
}

func writeCounts(outputFile *os.File, counts map[string]int) {
	for blamer, count := range counts {
		outputFile.WriteString(fmt.Sprintf("%s: %d\n", blamer, count))
	}
	outputFile.WriteString("\n")
}

func writeTopTodoBlamer() error {
	outputPath := filepath.Join(outputDir, "top_todo_blamer.txt")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	outputFile.WriteString("Top TODO Blamer:\n")
	writeTopBlamer(outputFile, todoCounts)

	outputFile.WriteString("\nTop FIXME Blamer:\n")
	writeTopBlamer(outputFile, fixmeCounts)

	return nil
}

func writeTopBlamer(outputFile *os.File, counts map[string]int) {
	sortedBlamers := sortBlamersByCount(counts)

	for i := 0; i < topTodoBlamer && i < len(sortedBlamers); i++ {
		outputFile.WriteString(fmt.Sprintf("%d. %s\n", i+1, sortedBlamers[i]))
	}
}

func sortBlamersByCount(counts map[string]int) []string {
	blamers := make([]string, 0, len(counts))
	for blamer := range counts {
		blamers = append(blamers, blamer)
	}

	// Sort blamers based on count (descending order)
	for i := 0; i < len(blamers)-1; i++ {
		for j := i + 1; j < len(blamers); j++ {
			if counts[blamers[j]] > counts[blamers[i]] {
				blamers[i], blamers[j] = blamers[j], blamers[i]
			}
		}
	}

	return blamers
}
