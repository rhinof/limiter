package limiter

import (
	"errors"
	"io"
	"sync"
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

	neededBuffer := len(p)
	acquiredBuffer := 0

	keepGoing := true
	for keepGoing {

		select {
		case addition := <-bucket.availableTokens:
			acquiredBuffer += addition
			if acquiredBuffer >= neededBuffer {
				keepGoing = false
			}
		default:
			keepGoing = false
		}
	}

	if acquiredBuffer >= neededBuffer {

		return bucket.wrapped.Write(p)
	}

	return 0, errors.New("throttled")
}

func (bucket *TokenBucket) init(bytesPerSecond int) {

	var wg = sync.WaitGroup{}
	wg.Add(1)

	go bucket.startPumpingTokens(bytesPerSecond, &wg)
	wg.Wait()
}

func (bucket *TokenBucket) startPumpingTokens(bytesPerSecond int, wg *sync.WaitGroup) {

	var once sync.Once
	onceBody := func() {
		wg.Done()
	}

	for {
		for range bucket.ticker.C {
			for i := 0; i < bytesPerSecond; i++ {
				select {
				case bucket.availableTokens <- 1:
				default:
				}
			}
			once.Do(onceBody)
		}
	}
}
