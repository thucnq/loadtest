package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"net/http"

	"github.com/panjf2000/ants"
)

// var ch = make(chan byte)
var counter int64

var url string = "http://172.16.0.30:8080/v3.0/media"
var contentType string = "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW"

// var url string = "http://127.0.0.1:8080/v3.0/media"
var AntsSize int = 2000
var n int = 1

// var ch = make(chan byte)
func postman() {
	url := "http://172.16.0.30:8080/v3.0/media"

	payload := strings.NewReader("------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"file\"; filename=\"image.jpg\"\r\nContent-Type: image/jpeg\r\n\r\n\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"source\"\r\n\r\naws\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"category\"\r\n\r\ndefault\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW\r\nContent-Disposition: form-data; name=\"format\"\r\n\r\njpg\r\n------WebKitFormBoundary7MA4YWxkTrZu0gW--")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW")
	// req.Header.Add("cache-control", "no-cache")
	// req.Header.Add("postman-token", "b6f97535-855f-2966-00ed-60b2f67b6e5f")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
}

func f(file *os.File) error {

	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true
	// http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = time.Millisecond
	// http.DefaultTransport.(*http.Transport).DialContext = (&net.Dialer{Timeout: time.Nanosecond}).DialContext
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
	// 	MaxVersion:         tls.VersionTLS11,
	// 	InsecureSkipVerify: true,
	// }
	defer http.DefaultTransport.(*http.Transport).CloseIdleConnections()

	// Buffer to store our request body as bytes
	var requestBody bytes.Buffer

	// Create a multipart writer
	multiPartWriter := multipart.NewWriter(&requestBody)

	// Initialize the file field
	fileWriter, err := multiPartWriter.CreateFormFile("file", "image.jpg")
	if err != nil {
		log.Fatalln(err)
		return err
	}

	// Copy the actual file content to the field field's writer
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	// Populate other fields
	fieldWriter, err := multiPartWriter.CreateFormField("source")
	if err != nil {
		log.Fatalln(err)
		return err
	}

	_, err = fieldWriter.Write([]byte("aws"))
	if err != nil {
		log.Fatalln(err)
		return err
	}

	// Populate other fields
	fieldWriter0, err := multiPartWriter.CreateFormField("category")
	if err != nil {
		log.Fatalln(err)
		return err
	}

	_, err = fieldWriter0.Write([]byte("default"))
	if err != nil {
		log.Fatalln(err)
		return err
	}

	// Populate other fields
	fieldWriter1, err := multiPartWriter.CreateFormField("format")
	if err != nil {
		log.Fatalln(err)
		return err
	}

	_, err = fieldWriter1.Write([]byte("jpg"))
	if err != nil {
		log.Fatalln(err)
		return err
	}

	// We completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	multiPartWriter.Close()

	// By now our original request body should have been populated, so let's just use it with our custom request
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	// We need to set the content type from the writer, it includes necessary boundary as well
	// fmt.Println(multiPartWriter.FormDataContentType())
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println(resp)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()
	return err
}

func main() {
	// postman()
	// return
	// Limit the number of spare OS threads to just 1
	runtime.GOMAXPROCS(6)

	// Make a copy of MemStats
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)

	t0 := time.Now().UnixNano()
	var wg sync.WaitGroup
	p, _ := ants.NewPool(AntsSize, ants.WithNonblocking(false), ants.WithPanicHandler(func(data interface{}) {
		fmt.Errorf("%s\n", data)
	}))
	defer p.Release()

	// Open the file
	file, err := os.Open("image.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	// Close the file later
	defer file.Close()

	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Submit(func() {
			err = f(file)
			if err != nil {
				log.Panic("err ", err)
			}
			wg.Done()
		})
	}
	wg.Wait()
	runtime.Gosched()
	t1 := time.Now().UnixNano()
	runtime.GC()

	// Make a copy of MemStats
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	fmt.Println("Adv Goroutine per second: %f", float64(n)/(float64(t1-t0)/float64(n)/10e3))
}
