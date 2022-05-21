package runner

import (
	"bufio"
	"math"
)

// SplitText tries to split a string by line while keeping the chunk size as close to maxChunkSize as possible (equal or less than maxChunkSize)
func SplitText(in string, maxChunkSize, searchLimit int) (chunks []string) {
	runes := []rune(in)
	totalSize := len(runes)
	minChunkSize := 1
	chunkOffset := 0

	if maxChunkSize > searchLimit {
		minChunkSize = maxChunkSize - searchLimit
	}
	maxPossibleChunks := int(math.Ceil(float64(totalSize) / float64(minChunkSize)))

	for i := 0; i <= maxPossibleChunks; i++ {

		chunkEnd := chunkOffset + maxChunkSize
		nextChunkStart := chunkEnd

		// Check if it is the last chunk (chunkEnd is greater or equal to total size)
		if chunkEnd >= totalSize {
			chunkEnd = totalSize
			nextChunkStart = totalSize
		} else {

			//Check for a line break
			for j := 0; j < searchLimit; j++ {

				sp := chunkEnd - j

				if sp < 0 {
					break
				}
				// Check if sp is the suitable split point
				if runes[sp] == '\n' {

					chunkEnd = sp
					nextChunkStart = chunkEnd + 1

					break
				}
			}

		}

		chunks = append(chunks, string(runes[chunkOffset:chunkEnd]))

		chunkOffset = nextChunkStart
		if chunkOffset >= totalSize {
			break
		}
	}

	return chunks
}

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
