package runner

import (
	"bufio"
)

// Return a bufio.SplitFunc that tries to split on newlines while giving as many bytes that are <= charLimit each time
func bulkSplitter(charLimit int) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// We need to be prepared to collect tokens because ScanLines trims trailing CR's for us
		tokens := make([]byte, 0, charLimit)

		advance, token, err = bufio.ScanLines(data, atEOF)

		if err != nil || token == nil {
			// Didn't get a line
			return
		}

		if len(token) >= charLimit {
			// Got too much. Give charLimit bytes and finish the rest of the line next time.
			advance = charLimit
			token = token[:charLimit]
			return
		}

		tokens = append(tokens, token...)

		// Keep getting lines until we exceed charLimit
		for {
			newAdvance, token, err := bufio.ScanLines(data[advance:], atEOF)

			if err != nil || token == nil {
				// Failed to get a line
				break
			}

			if len(tokens)+len(token) > charLimit {
				// Too much. Give what we had.
				return advance, tokens, err
			}

			advance += newAdvance
			tokens = append(tokens, '\n')
			tokens = append(tokens, token...)
		}

		// Stopped getting lines but still hungry for bytes

		// Are we done?
		if atEOF {
			return advance, tokens, nil
		}

		// Need more data
		return 0, nil, nil
	}
}
