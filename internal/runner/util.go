package runner

import (
	"bufio"
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
					if len(token) > 0 {
						// Remove the trailing newline
						token = token[:len(token)-1]
					}
					return
				}

				// We need more data
				return 0, nil, nil
			}

			if len(token)+len(line) > charLimit {
				// What we had and what we got is too much
				if len(token) == 0 {
					// Even just the first line is too much
					// Truncate and give it
					token = append(token, line[:charLimit-len(ellipsis)]...)
					token = append(token, ellipsis...)
					advance = charLimit - len(ellipsis)
				}
				return
			}

			advance += lineAdvance
			token = append(token, line...)
			token = append(token, '\n')
		}
	}
}
