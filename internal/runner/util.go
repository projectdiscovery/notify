package runner

import (
	"bufio"
	"math"
)

var ellipsis = []byte("...")

// Return a bufio.SplitFunc that tries to split on newlines while giving as many bytes that are <= charLimit each time
func bulkSplitter(charLimit int) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		var lineAdvance int
		var line []byte

		// Keep getting lines until we exceed charLimit
		for {
			lineAdvance, line, err = bufio.ScanLines(data[advance:], atEOF)

			if err != nil || line == nil {
				// Failed to get a line
				if atEOF {
					// We're done
					break
				}

				// We need more data
				return
			}

			if len(token)+len(line) > charLimit {
				// What we had and what we got is too much
				if len(token) == 0 {
					// Even just the first line is too much
					// Truncate and give it
					token = append(token, line[:charLimit-len(ellipsis)]...)
					token = append(token, ellipsis...)
					advance = charLimit - len(ellipsis)
					return
				} else {
					// Give what we had
					break
				}
			}

			advance += lineAdvance
			token = append(token, line...)
			token = append(token, '\n')
		}

		if len(token) > 0 {
			// We have something. It'll have a trailing newline
			// Remove it
			token = token[:len(token)-1]
		}
		return
	}
}

// SplitInChunks splits a string into chunks of size charLimit
func SplitInChunks(data string, charLimit int) []string {
	length := len(data)
	noOfChunks := int(math.Ceil(float64(length) / float64(charLimit)))
	chunks := make([]string, noOfChunks)
	var start, stop int

	for i := 0; i < noOfChunks; i += 1 {
		start = i * charLimit
		stop = start + charLimit
		if stop > length {
			stop = length
		}
		chunks[i] = data[start:stop]
	}
	return chunks
}
