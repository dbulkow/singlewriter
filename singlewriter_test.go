package singlewriter_test

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	. "yin.mno.stratus.com/gogs/dbulkow/singlewriter"
)

func TestWriter(t *testing.T) {
	buf := NewSingleWriter()

	data := "this is a test of the emergency broadcast system"

	n, err := buf.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal("length")
	}

	if err := buf.Close(); err != nil {
		t.Fatal("close")
	}
}

func TestReadWrite(t *testing.T) {
	buf := NewSingleWriter()

	go func(b *SingleWriter) {
		time.Sleep(100 * time.Millisecond)

		data := []string{
			"this is a test of the emergency broadcast system",
			"now is the time for all good men to come to the aid of their country",
			"testing, testing, 1, 2, 3",
		}

		defer func() {
			if err := b.Close(); err != nil {
				t.Fatal("close")
			}
		}()

		for _, d := range data {
			n, err := b.Write([]byte(d))
			if err != nil {
				t.Fatal(err)
			}
			if n != len(d) {
				t.Fatal("length")
			}
			time.Sleep(10 * time.Millisecond)
		}
	}(buf)

	rdr, err := buf.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer rdr.Close()

	x := make([]byte, 512)

	for {
		n, err := rdr.Read(x)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}

		fmt.Println(string(x[:n]))
	}
}

func TestWriteReadReadRead(t *testing.T) {
	buf := NewSingleWriter()

	time.Sleep(100 * time.Millisecond)

	data := []string{
		"this is a test of the emergency broadcast system\n",
		"now is the time for all good men to come to the aid of their country\n",
		"testing, testing, 1, 2, 3\n",
	}

	for _, d := range data {
		n, err := buf.Write([]byte(d))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(d) {
			t.Fatal("length")
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := buf.Close(); err != nil {
		t.Fatal("close")
	}

	var wg sync.WaitGroup

	reader := func(b *SingleWriter) {
		defer wg.Done()

		rdr, err := buf.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer rdr.Close()

		x := make([]byte, 512)

		for {
			n, err := rdr.Read(x)
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}

			fmt.Println(string(x[:n]))
		}
	}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go reader(buf)
	}

	wg.Wait()
}
