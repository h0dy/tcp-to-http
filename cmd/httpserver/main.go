package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/h0dy/tcp-to-http/internal/headers"
	"github.com/h0dy/tcp-to-http/internal/request"
	"github.com/h0dy/tcp-to-http/internal/response"
	"github.com/h0dy/tcp-to-http/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		log.Fatalln("please make sure to setup PORT env")
	}

	// server.Serve starts an HTTP server
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

// handler routes the request to the appropriate response
func handler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		proxyHandler(w, req)
	}

	switch req.RequestLine.RequestTarget {
	case "/client-error":
		handler400(w, req)

	case "/internal-error":
		handler500(w, req)

	case "/video":
		videoHandler(w, req)

	case "/home":
		homeHandler(w, req)

	default:
		handler200(w, req)
	}
}

func homeHandler(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.ClientError)
	body := []byte(`<html>
<head>
<title>Welcome to Homepage</title>
</head>
<body>
<h1>Hello, World!</h1>
<p>Welcome to the homepage of this simple HTTP server.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.ClientError)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Client request error.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.ServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>The server was unable to complete your request. Please try again later.</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.Successful)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Check out the homepage!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!</p>
</body>
</html>
`)
	h := response.GetDefaultHeaders(len(body))
	h.Update("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

// proxyHandler forwards the incoming request to httpbin.org and streams
// the response back to the client using HTTP chunked transfer encoding
func proxyHandler(w *response.Writer, req *request.Request) {
	// Remove the "/httpbin/" prefix to get  the target path
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	fullUrl := "https://httpbin.org/" + target
	fmt.Println("Proxying to", fullUrl)

	resp, err := http.Get(fullUrl)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.Successful)
	h := response.GetDefaultHeaders(0)
	h.Update("Transfer-Encoding", "chunked")
	h.Update("Trailer", "X-Content-SHA256, X-Content-Length")
	h.Remove("Content-Length")
	w.WriteHeaders(h)

	fullBody := make([]byte, 0)
	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)

	// Read the response body in chunks and forward them
	for {
		n, err := resp.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")

		if n > 0 {
			// Write the chunk to the client
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			fullBody = append(fullBody, buffer[:n]...)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}

	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}

	trailers := headers.NewHeaders()

	// compute SHA-256 hash of the full response body
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Update("X-Content-SHA256", sha256)
	trailers.Update("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))

	// write the trailer to the client
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")
}

// videoHandler streams a video from assets folder
func videoHandler(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.Successful)
	filePath := os.Getenv("VIDEO_PATH")
	if filePath == "" {
		log.Fatalln("please make sure to setup VIDEO_PATH env")
	}
	videoBytes, err := os.ReadFile(filePath)
	if err != nil {
		handler500(w, nil)
		return
	}

	h := response.GetDefaultHeaders(len(videoBytes))
	h.Update("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(videoBytes)
}
