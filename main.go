package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"runtime"
	"sync"
	"time"

	"net/http"

	"github.com/panjf2000/ants"
)

// var ch = make(chan byte)
var counter int64

var url string = "http://172.16.0.30:8080/v3.0/media"

// var url string = "http://127.0.0.1:8080/v3.0/media"
var AntsSize int = 500
var n int = 100000

// var ch = make(chan byte)

func f(file *os.File) error {

	http.DefaultTransport.(*http.Transport).DisableKeepAlives = true

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
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	// Do the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}

func main() {

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
