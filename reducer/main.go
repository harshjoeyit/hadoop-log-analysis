package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// // default 64KB buffer too small for 1.8M record input
	// const maxCapacity = 1024 * 1024
	// buf := make([]byte, maxCapacity)
	// scanner.Buffer(buf, maxCapacity)

	counts := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		if len(parts) != 2 {
			continue
		}

		url := parts[0]
		cnt, err := strconv.Atoi(parts[1])
		if err == nil {
			counts[url] += cnt
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "scanner error: %v\n", err)
	}

	// fmt.Printf is unbuffered per-call but on large output
	// the OS pipe buffer can be cut off. Use explicit buffered writer:
	writer := bufio.NewWriter(os.Stdout)

	for url, count := range counts {
		fmt.Fprintf(writer, "%s\t%d\n", url, count)
	}

	writer.Flush()
}
