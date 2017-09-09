package limiter

import (
	"errors"
	"io"
	"math"
	"time"
)

//New Creates a new TokenBucket which will limit the rate to the value specified in the bytesPerSecond argument
func New(bytesPerSecond int, wrapped io.Writer) TokenBucket {

	timeResolution := time.Nanosecond * 1000
	maxThrouput := float64(bytesPerSecond)
	tokenSize := maxThrouput / float64(timeResolution.Nanoseconds())
	burstSize := int(math.Ceil(maxThrouput / float64(tokenSize)))

	tb := TokenBucket{
		ticker:          time.NewTicker(timeResolution),
		availableTokens: make(chan float64, burstSize),
		tokenSize:       tokenSize}
	tb.init(tokenSize)
	return tb
}

//TokenBucket implementing the TokenBucket algorithim (see https://en.wikipedia.org/wiki/Token_bucket)
type TokenBucket struct {
	ticker          *time.Ticker
	availableTokens chan float64
	tokenSize       float64
	wrapped         io.Writer
}

func (bucket TokenBucket) Write(p []byte) (n int, err error) {

	neededBuffer := float64(len(p)) / bucket.tokenSize
	acquiredBuffer := 0.00

	keepGoing := true
	for keepGoing {

		select {
		case addition := <-bucket.availableTokens:
			acquiredBuffer += addition
		default:
			keepGoing = false
		}
	}

	if acquiredBuffer >= neededBuffer {
		return bucket.wrapped.Write(p)
	}

	return 0, errors.New("throttled")
}

func (bucket *TokenBucket) init(additionSize float64) {

	go bucket.startPumpingTokens(additionSize)
}

func (bucket *TokenBucket) startPumpingTokens(additionSize float64) {

	for {
		for range bucket.ticker.C {
			select {
			case bucket.availableTokens <- additionSize:
			default:
			}
		}
	}
}
