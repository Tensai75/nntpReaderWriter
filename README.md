# nntpReaderWriter

A high-performance, protocol-correct Go NNTP client reader/writer package. It supports pipelined command execution, dot-encoded multi-line responses, and efficient buffer management.
The package is designed for correctness and performance and can be used as a replacement for net/textproto in high-throughput NNTP applications.

**Note:** `NntpReaderWriter` is not safe for concurrent use by multiple goroutines. All access must be externally synchronized.

## Features

- Pipelined NNTP command execution
- Dot-encoded multi-line response handling
- Efficient buffer management
- Custom error types for protocol and unexpected responses

## Installation

```
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
    client := nntpReaderWriter.NewNntpReaderWriter(conn, opts)

    // Example: Issue a simple NNTP command
    code, msg, err := client.SingleResponseLineCmd("MODE READER")
    if err != nil {
        log.Fatalf("NNTP error: %v", err)
    }
    fmt.Printf("Response: %d %s\n", code, msg)

    // Example: Read multi-line response as strings
    code, msg, lines, err := client.DotLinesReadCmdAsStrings("LIST")
    if err != nil {
        log.Fatalf("LIST error: %v", err)
    }
    fmt.Printf("LIST: %d %s\nLines: %v\n", code, msg, lines[:3]) // print first 3 lines

    // Example: Write multi-line data
    // Line breaks and single dot end line are added by nntpReaderWriter
    postLines := []string{"From: test@example.com", "Subject: Test", "", "This is a test."}
    code, msg, err = client.DotLinesWriteCmdFromStrings("POST", postLines)
    if err != nil {
        log.Fatalf("POST error: %v", err)
    }
    fmt.Printf("POST: %d %s\n", code, msg)
}
```

See the [GoDoc](https://pkg.go.dev/github.com/Tensai75/nntpReaderWriter) for full API documentation.

## Error Handling

The package provides custom error types:

- `NntpError` for server error responses (4xx/5xx)
- `UnexpectedResponseCodeError` for syntactically valid but unexpected codes

You can check for these using:

```go
if nntpReaderWriter.IsNntpError(err) { /* ... */ }
if nntpReaderWriter.IsUnexpectedResponseCodeError(err) { /* ... */ }
```

## License

MIT
