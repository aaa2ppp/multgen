package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"time"
	"unsafe"
)

var (
	help     = flag.Bool("help", false, "show usage help")
	minX     = flag.Float64("min", 1, "min sequence value, must be >= 1.0")
	maxX     = flag.Float64("max", 10000, "max sequence value, must be >= 1.0")
	multiply = flag.Bool("m", false, "if this flag is set, then transform = x * m, otherwise x")
)

func main() {
	flag.Parse()

	if *help {
		fmt.Fprintln(os.Stderr, "Usee: check [-min=<value>] [-max=<value>] [-m]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !(*minX >= 1 && *maxX >= 1) {
		fmt.Fprintf(os.Stderr, "mix and max sequence value must be >= 1")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if !(*minX <= *maxX) {
		fmt.Fprintf(os.Stderr, "min sequence value must be <= max sequence value")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sc := bufio.NewScanner(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	var (
		m             float64 // мультипликатор
		x             float64 // значение последовательности
		t             float64 // значение после трансформации
		d             = *maxX - *minX
		count         int
		totalPayment  float64
		totalProfit   float64
		maxMultiplier float64
		err           error
	)

	start := time.Now()
	for sc.Scan() {
		// read the multiplier
		// NOTE: We use `unsafeString` to performance. It's safe here because we don't save the returned string anywhere.
		m, err = strconv.ParseFloat(unsafeString(sc.Bytes()), 64)
		if err != nil {
			log.Fatalf("unexpected input: %q: %v", sc.Text(), err)
		}

		// get an element of a "random" sequence
		x = *minX
		if d > 0 {
			x += rand.Float64() * d
		}

		// transformate t = F(m, x)
		if m <= x {
			t = 0
		} else if !*multiply {
			t = x
		} else {
			t = x * m
		}

		// count aggregates
		count++
		totalPayment += x
		totalProfit += t
		maxMultiplier = max(maxMultiplier, m)
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	rtp := totalProfit / totalPayment

	log.Printf("count=%d elapsed=%v payment=%0.3f profit=%0.3f max_multiplier=%0.3f",
		count, time.Since(start), totalPayment, totalProfit, maxMultiplier)

	w.WriteString(strconv.FormatFloat(rtp, 'g', -1, 64))
	w.WriteByte('\n')
}

func unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
