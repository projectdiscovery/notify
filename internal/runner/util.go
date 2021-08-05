package runner

import (
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
