package runner

import (
	"bufio"
	"fmt"
)

var ellipsis = []byte("...")

// Return a bufio.SplitFunc that splits on as few newlines as possible
// while giving as many bytes that are <= charLimit each time.
// Note: charLimit must be <= the buffer underlying the bufio.Scanner
func bulkSplitter(charLimit int) (bufio.SplitFunc, error) {
	if charLimit <= len(ellipsis) {
		return nil, fmt.Errorf("charLimit must be > %d", len(ellipsis))
	}
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

				if len(data) < charLimit {
					// We need more data
					return 0, nil, nil
				} else {
					// We have enough data, but bufio.ScanLines couldn't see a newline in what's left
					// If we handle this then we can assure the bufio.Scanner will never give bufio.ErrTooLong
					if len(token) == 0 {
						// Even just the first line is too much
						// Truncate and give it
						token = append(token, data[:charLimit-len(ellipsis)]...)
						token = append(token, ellipsis...)
						advance = charLimit - len(ellipsis)
						return advance, token, nil
					} else {
						// Give what we had
						break
					}
				}
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
	}, nil
}

// Return a bufio.SplitFunc that splits on all newlines
// while giving as many bytes that are <= charLimit each time.
// Note: charLimit must be <= the buffer underlying the bufio.Scanner
func lineLengthSplitter(charLimit int) (bufio.SplitFunc, error) {
	if charLimit <= len(ellipsis) {
		return nil, fmt.Errorf("charLimit must be > %d", len(ellipsis))
	}
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		var line []byte

		// Get a line
		advance, line, err = bufio.ScanLines(data[advance:], atEOF)

		if !atEOF && (err != nil || line == nil) {
			if len(data) < charLimit {
				// We need more data
				return 0, nil, nil
			} else {
				// We have enough data, but bufio.ScanLines couldn't see a newline in it
				// If we handle this then we can assure the bufio.Scanner will never give bufio.ErrTooLong
				token = append(token, data[:charLimit-len(ellipsis)]...)
				token = append(token, ellipsis...)
				advance = charLimit - len(ellipsis)
				return advance, token, nil
			}
		}

		if len(line) > charLimit {
			// Got too much
			token = append(token, line[:charLimit-len(ellipsis)]...)
			token = append(token, ellipsis...)
			advance = charLimit - len(ellipsis)
			return
		}

		return advance, line, err
	}, nil
}
