package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func extractBookmarks(r io.Reader) ([]string, error) {
	var bookmarks []string
	inDL := false
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<DL>") {
			inDL = true
			continue
		}
		if strings.Contains(line, "</DL>") {
			inDL = false
			continue
		}
		if inDL && strings.Contains(line, "<DT>") {
			bookmarks = append(bookmarks, line)
		}
	}
	return bookmarks, scanner.Err()
}

func getHeaderAndFooter(r io.Reader) (header, footer string, err error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	inDL := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "<DL>") {
			inDL = true
			lines = append(lines, line)
			break
		}
		lines = append(lines, line)
	}
	header = strings.Join(lines, "\n")
	// Footer is just the closing </DL> and anything after
	var footerLines []string
	foundEnd := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "</DL>") {
			foundEnd = true
		}
		if foundEnd {
			footerLines = append(footerLines, line)
		}
	}
	footer = strings.Join(footerLines, "\n")
	return header, footer, scanner.Err()
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run merge_bookmarks.go <file1.html> <file2.html> <output.html>")
		os.Exit(1)
	}
	file1, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file1.Close()

	file2, err := os.Open(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer file2.Close()

	// For header and footer, use the first file as template
	file1ForHeader, _ := os.Open(os.Args[1])
	defer file1ForHeader.Close()
	header, footer, err := getHeaderAndFooter(file1ForHeader)
	if err != nil {
		panic(err)
	}

	bookmarks1, err := extractBookmarks(file1)
	if err != nil {
		panic(err)
	}
	bookmarks2, err := extractBookmarks(file2)
	if err != nil {
		panic(err)
	}

	out, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Write header
	fmt.Fprintln(out, header)
	// Write bookmarks
	for _, bm := range bookmarks1 {
		fmt.Fprintln(out, bm)
	}
	for _, bm := range bookmarks2 {
		fmt.Fprintln(out, bm)
	}
	// Write footer
	fmt.Fprintln(out, footer)

	fmt.Println("Merged bookmarks written to", os.Args[3])
}
