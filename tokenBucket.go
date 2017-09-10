package limiter

import (
	"errors"
	"io"
	"time"
)

//New Creates a new TokenBucket which will limit the rate to the value specified in the bytesPerSecond argument
func New(wrapped io.Writer, bytesPerSecond int) TokenBucket {

	timeResolution := time.Second

	tb := TokenBucket{
		wrapped:         wrapped,
		ticker:          time.NewTicker(timeResolution),
		availableTokens: make(chan int, bytesPerSecond),
	}
	tb.init(bytesPerSecond)

	return tb
}

//TokenBucket implementing the TokenBucket algorithim (see https://en.wikipedia.org/wiki/Token_bucket)
type TokenBucket struct {
	ticker          *time.Ticker
	availableTokens chan int
	wrapped         io.Writer
}

func (bucket TokenBucket) Write(p []byte) (n int, err error) {

	neededTokens := len(p)
	acquiredTokens := 0

	// pre check that there is enough tokens before draining the bucket
	if len(bucket.availableTokens) < neededTokens {
		return 0, errors.New("throttled")
	}

	// in order to address potential race conditions actually drain the bucket from the tokens
	keepGoing := true
	for keepGoing {

		select {
		case token := <-bucket.availableTokens:
			acquiredTokens += token
			if acquiredTokens >= neededTokens {
				keepGoing = false
			}
		default:
			// return tokens to the bucket
			bucket.addTokens(acquiredTokens)
			keepGoing = false
		}
	}

	if acquiredTokens >= neededTokens {

		return bucket.wrapped.Write(p)
	}

	return 0, errors.New("throttled")
}

func (bucket *TokenBucket) init(numberOfTokens int) {

	bucket.addTokens(numberOfTokens)
	go bucket.pumpTokens(numberOfTokens)

}

func (bucket *TokenBucket) pumpTokens(numberOfTokens int) {

	for {
		for range bucket.ticker.C {
			bucket.addTokens(numberOfTokens)
		}
	}
}

func (bucket *TokenBucket) addTokens(numberOfTokens int) {

	for i := 0; i < numberOfTokens; i++ {
		select {
		case bucket.availableTokens <- 1:
		default:
		}
	}
}
