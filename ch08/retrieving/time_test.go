package main

import (
	"net/http"
	"testing"
	"time"
)


func TestHeadTime(t *testing.T) {
    resp, err := http.Head("https://www.time.gov/")
    if err != nil {
	t.Fatal(err)
    }

    // Always close this without exception.
    // For details see next section in the book.
    _ = resp.Body.Close()

    now := time.Now().Round(time.Second)
    date := resp.Header.Get("Date")
    if date == "" {
	t.Fatal("no Date header received from time.gov")
    }

    parsedTime, err := time.Parse(time.RFC1123, date)
    if err != nil {
	t.Fatal(err)
    }

    t.Logf("time.gov: %s (skew %s)", parsedTime, now.Sub(parsedTime))
}
