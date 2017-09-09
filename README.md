# limiter

An implementation of the [Token Bucket](https://en.wikipedia.org/wiki/Token_bucket) rate limiting algorithm

## Usage

```go
     //max # of bytes/s to allow
	maxThroughput := 1000
	// wrapped io.Writer interface
	writer := getWriter()	
	limiter := limiter.New(writer, maxThroughput)
	buffer := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	n, err := limiter.Write(buffer)
	fmt.Printf("written %d bytes to wrapped stream - err: %v", n, err)
```