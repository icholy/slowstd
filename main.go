package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

type ByteThrottler struct {
	Bytes  int
	Period time.Duration
}

func (t ByteThrottler) SleepDuration(bytes int) time.Duration {
	if t.Bytes == 0 {
		return 0
	}
	durationPerByte := t.Period / time.Duration(t.Bytes)
	return durationPerByte * time.Duration(bytes)
}

func (t ByteThrottler) Sleep(bytes int) {
	time.Sleep(t.SleepDuration(bytes))
}

type ThrottledReader struct {
	R io.Reader
	T ByteThrottler
}

func (r ThrottledReader) Read(data []byte) (int, error) {
	n, err := r.R.Read(data)
	r.T.Sleep(n)
	return n, err
}

type ThrottledWriter struct {
	W io.Writer
	T ByteThrottler
}

func (w ThrottledWriter) Write(data []byte) (int, error) {
	w.T.Sleep(len(data))
	return w.W.Write(data)
}

func main() {
	var rT, wT ByteThrottler
	flag.IntVar(&rT.Bytes, "rb", 1000, "read bytes per time period")
	flag.IntVar(&wT.Bytes, "wb", 1000, "write bytes per time period")
	flag.DurationVar(&rT.Period, "rp", time.Second, "read time period")
	flag.DurationVar(&wT.Period, "wp", time.Second, "write time period")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("no command")
	}
	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	cmd.Stdin = ThrottledReader{R: os.Stdin, T: rT}
	cmd.Stdout = ThrottledWriter{W: os.Stdout, T: wT}
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
