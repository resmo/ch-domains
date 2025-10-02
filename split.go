package main

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "sort"
    "strings"
    "sync"
)

type FileData struct {
    dirName  string
    fileName string
    lines    map[string]bool // Using map for automatic deduplication
    parentDir string // Add parent directory field
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run split.go <input_file.txt>")
        os.Exit(1)
    }

    inputFile := os.Args[1]

    // Check if input file exists
    if _, err := os.Stat(inputFile); os.IsNotExist(err) {
        fmt.Printf("Error: File '%s' not found!\n", inputFile)
        os.Exit(1)
    }

    // Get parent directory name from input file (without .txt extension)
    baseName := filepath.Base(inputFile)
    parentDirName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
    fmt.Printf("Creating parent directory: %s\n", parentDirName)

    // Open input file
    file, err := os.Open(inputFile)
    if err != nil {
        fmt.Printf("Error opening file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()

    // Map to collect lines by directory and file
    fileMap := make(map[string]*FileData)

    // Read file line by line
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // Skip empty lines
        if line == "" {
            continue
        }

        // Get first character (lowercase)
        firstChar := strings.ToLower(string(line[0]))

        // Get first two characters (lowercase)
        firstTwoChars := ""
        if len(line) >= 2 {
            firstTwoChars = strings.ToLower(line[:2])
        } else {
            firstTwoChars = strings.ToLower(line)
        }

        // Create key for the file
        key := parentDirName + "/" + firstChar + "/" + firstTwoChars + ".txt"

        // Initialize FileData if not exists
        if fileMap[key] == nil {
            fileMap[key] = &FileData{
                parentDir: parentDirName,
                dirName:  firstChar,
                fileName: firstTwoChars + ".txt",
                lines:    make(map[string]bool),
            }
        }

        // Add line (map automatically handles duplicates)
        fileMap[key].lines[line] = true
    }

    if err := scanner.Err(); err != nil {
        fmt.Printf("Error reading file: %v\n", err)
        os.Exit(1)
    }

    // Use goroutines for parallel processing
    numWorkers := runtime.NumCPU()
    fmt.Printf("Using %d CPU cores for processing...\n", numWorkers)

    // Create channels for work distribution
    jobs := make(chan *FileData, len(fileMap))
    results := make(chan string, len(fileMap))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go worker(jobs, results, &wg)
    }

    // Send jobs
    go func() {
        for _, data := range fileMap {
            jobs <- data
        }
        close(jobs)
    }()

    // Close results channel when all workers are done
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    var processedFiles []string
    for result := range results {
        processedFiles = append(processedFiles, result)
    }

    // Print results
    for _, result := range processedFiles {
        fmt.Println(result)
    }

    fmt.Println("Sorting completed!")
    fmt.Println("Lines have been sorted into directories by first letter and files by first two letters.")
    fmt.Println("Duplicates removed and entries sorted alphabetically.")
}

func worker(jobs <-chan *FileData, results chan<- string, wg *sync.WaitGroup) {
    defer wg.Done()

    for data := range jobs {
        // Create full directory path (parent/firstchar)
        fullDirPath := filepath.Join(data.parentDir, data.dirName)
        err := os.MkdirAll(fullDirPath, 0755)
        if err != nil {
            results <- fmt.Sprintf("Error creating directory %s: %v", fullDirPath, err)
            continue
        }

        // Convert map to sorted slice
        var sortedLines []string
        for line := range data.lines {
            sortedLines = append(sortedLines, line)
        }
        sort.Strings(sortedLines)

        // Write to file
        filePath := filepath.Join(fullDirPath, data.fileName)
        file, err := os.Create(filePath)
        if err != nil {
            results <- fmt.Sprintf("Error creating file %s: %v", filePath, err)
            continue
        }

        writer := bufio.NewWriter(file)
        for _, line := range sortedLines {
            writer.WriteString(line + "\n")
        }
        writer.Flush()
        file.Close()

        results <- fmt.Sprintf("Processed %s: %d unique entries", filePath, len(sortedLines))
    }
}
