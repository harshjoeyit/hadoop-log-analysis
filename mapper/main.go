package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	// Regular expression to extract the URL path from the log line
	re := regexp.MustCompile(`"GET (.+?) HTTP`)

	// fmt.Printf is unbuffered per-call but on large output
	// the OS pipe buffer can be cut off. Use explicit buffered writer:
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			url := matches[1]
			// Emit the URL path with a count of 1
			fmt.Fprintf(writer, "%s\t%d\n", url, 1)
		}
	}

	writer.Flush()
}
