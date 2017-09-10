package limiter

import (
	"os"
	"testing"
)

func TestThrottling(t *testing.T) {

	limiter := NewLimiter(os.Stdout, 5)
	bytes := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	_, err := limiter.Write(bytes)

	if err == nil {
		t.Fail()
	}

}

func TestHappyPath(t *testing.T) {

	limiter := NewLimiter(os.Stdout, 10)
	bytes := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	written, _ := limiter.Write(bytes)
	if written != len(bytes) {
		t.Fail()
	}

}
