# nntpReaderWriter

A high-performance, protocol-correct Go NNTP client reader/writer package. It supports sequenced command execution (via an internal sequencer), dot-encoded multi-line responses, and efficient buffer management. Designed for correctness and throughput.

**Note:** `NntpReaderWriter` is not safe for concurrent use by multiple goroutines. All access must be externally synchronized.

## Features

- Efficient buffer management for both reads and writes
- Dot-encoded multi-line handling (reading and writing)
- Custom error types for protocol and unexpected responses
- Minimal, low-level API for maximum flexibility
- Sequenced NNTP command execution with strict in-order guarantees (via an internal sequencer)
- Comprehensive test suite with protocol and concurrency mocks
- Benchmark suite for performance comparison

## Installation

```sh
go get github.com/Tensai75/nntpReaderWriter
```

## Usage Example

```go
package main

import (
    "fmt"
    "log"
    "net"
    "time"

    "github.com/Tensai75/nntpReaderWriter"
)

func main() {
    conn, err := net.DialTimeout("tcp", "news.example.com:119", 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    opts := nntpReaderWriter.NntpReaderWriterOptions{
        ReadTimeout:   10 * time.Second,
        WriteTimeout:  10 * time.Second,
    }
    nntpRW := nntpReaderWriter.NewNntpReaderWriter(conn, opts)

    // Start a sequence for a command
    seqID := nntpRW.StartSequence()

    // Write a command
    if err := nntpRW.WriteLine("MODE READER"); err != nil {
        if nntpReaderWriter.IsProtocolError(err) {
            log.Fatalf("protocol error: %v", err)
        }
        log.Fatalf("write error: %v", err)
    }

    // Read the response code and message
    code, msg, err := nntpRW.ReadCodeResponseLine()
    if err != nil {
        if nntpReaderWriter.IsProtocolError(err) {
            log.Fatalf("protocol error: %v", err)
        }
        log.Fatalf("read error: %v", err)
    }
    fmt.Printf("Response: %d %s\n", code, msg)

    // Write another command
    if err := nntpRW.WriteLine("LIST"); err != nil {
        log.Fatalf("write error: %v", err)
    }

    // Read the response code and message for the second command
    code, msg, err = nntpRW.ReadCodeResponseLine()
    if err != nil {
        if nntpReaderWriter.IsProtocolError(err) {
            log.Fatalf("protocol error: %v", err)
        }
        log.Fatalf("read error: %v", err)
    }
    fmt.Printf("Response: %d %s\n", code, msg)

    // Read the dot-encoded multi-line response for the second command
    lines, err := nntpRW.ReadDotLines()
    if err != nil {
        if nntpReaderWriter.IsProtocolError(err) {
            log.Fatalf("protocol error: %v", err)
        }
        log.Fatalf("error reading dot lines: %v", err)
    }
    for _, line := range lines {
        fmt.Println(line)
    }

    // End the sequence to allow the next command to proceed
    nntpRW.EndSequence(seqID)
}
```

## Advanced Features

- **Dot-line streaming reading/writing:** Use `ReadDotLinesReader` and `WriteDotLinesReader` for streaming multi-line data reading or writing.
- **Dot-line streaming reading with callback:** Use `ReadDotLinesCallback` or `ReadDotLinesCallbackBytes` for streaming line-by-line processing of multi-line responses either as strings or bytes.
- **Custom error types:** Check for protocol errors, invalid lines, and unexpected response codes using exported error helpers.
- **Mocking and testing:** Use the `mockScriptedConn` package for protocol and concurrency tests.

## Testing & Benchmarks

- Run all tests: `go test ./...`
- Run benchmarks: `go test -bench .`
- The package includes concurrency and protocol correctness tests.

## Performance

Benchmarks results of reading and writing 100'000 lines with a lenght of 1024 bytes each:

```
goos: windows
goarch: amd64
pkg: github.com/Tensai75/nntpReaderWriter
cpu: AMD Ryzen 9 5950X 16-Core Processor
BenchmarkReadLinesReader-32                  100          11910156 ns/op        8597.70 MB/s        4479 B/op          3 allocs/op
BenchmarkReadLinesStrings-32                  51          25166108 ns/op        4068.96 MB/s    111324784 B/op    100029 allocs/op
BenchmarkReadLinesCallback-32                 52          24375152 ns/op        4201.00 MB/s    102401280 B/op    100001 allocs/op
BenchmarkReadLinesCallbackBytes-32           127           9378006 ns/op        10919.17 MB/s       1280 B/op          1 allocs/op
BenchmarkWriteLinesReader-32                  38          30077329 ns/op        3404.56 MB/s        1152 B/op          1 allocs/op
BenchmarkWriteLinesStrings-32                 50          22232012 ns/op        4605.97 MB/s           0 B/op          0 allocs/op
BenchmarkWriteLinesChannel-32                 48          24581540 ns/op        4165.73 MB/s           0 B/op          0 allocs/op
```

For highest performance use `ReadDotLinesCallbackBytes` for reading multi-line responses as bytes, and `WriteDotLines` or `WriteDotLinesChannel` for writing multi-line requests from strings.

## Full docs

See the [GoDoc](https://pkg.go.dev/github.com/Tensai75/nntpReaderWriter) for full API documentation.

## License

MIT
