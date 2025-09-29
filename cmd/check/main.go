package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"time"
	"unsafe"

	"github.com/aaa2ppp/multgen/internal/checker"
)

var (
	help       = flag.Bool("help", false, "show usage help")
	minX       = flag.Float64("min", 1, "min sequence value, must be >= 1.0")
	maxX       = flag.Float64("max", 10000, "max sequence value, must be >= 1.0")
	one        = flag.Bool("1", false, "if this flag is set, then payment = 1")
	multiply   = flag.Bool("m", false, "if this flag is set, then transform = x * m, otherwise x")
	playersNum = flag.Int("n", 1, "number of playes")
	verbose    = flag.Bool("v", false, "output human-readable results in stderr")
)

func validateFlags() error {
	var errs []error

	if !(*minX >= 1 && *maxX >= 1) {
		errs = append(errs, errors.New("mix and max sequence value must be >= 1"))
	}

	if !(*minX <= *maxX) {
		errs = append(errs, errors.New("min sequence value must be <= max sequence value"))
	}

	if !(*playersNum >= 1) {
		errs = append(errs, errors.New("number of playesr must be >= 1"))
	}

	return errors.Join(errs...)
}

func main() {
	flag.Parse()

	if *help {
		fmt.Fprintln(os.Stderr, "Usee: check [options]\nOptions:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if err := validateFlags(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.PrintDefaults()
		os.Exit(1)
	}

	sc := bufio.NewScanner(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	type player struct {
		totalPayment float64
		totalProfit  float64
	}

	if *minX == *maxX {
		*playersNum = 1
	}

	log.Printf("players=%d", *playersNum)
	players := make([]player, *playersNum)

	var (
		m             float64 // мультипликатор
		x             float64 // значение последовательности
		p             = 1.0   // платеж
		t             float64 // значение после трансформации
		d             = *maxX - *minX
		count         int
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

		for i := range players {
			// get an element of a "random" sequence
			x = *minX
			if d > 0 {
				x += rand.Float64() * d
			}

			if !*one {
				p = x
			}

			// transformate t = F(m, x)
			if m <= x {
				t = 0
			} else if !*multiply {
				t = x
			} else {
				t = x * m
			}

			// count player aggregates
			players[i].totalPayment += p
			players[i].totalProfit += t
		}

		// count common aggregates
		count++
		maxMultiplier = max(maxMultiplier, m)
	}

	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	// backward compatibility
	if len(players) == 1 {
		totalPayment := players[0].totalPayment
		totalProfit := players[0].totalProfit

		rtp := totalProfit / totalPayment

		if *verbose {
			log.Printf("count=%d elapsed=%v payment=%0.3f profit=%0.3f max_multiplier=%g",
				count, time.Since(start), totalPayment, totalProfit, maxMultiplier)
		}

		w.WriteString(strconv.FormatFloat(rtp, 'g', -1, 64))
		w.WriteByte('\n')
		return
	}

	rtps := make([]float64, len(players))
	for i := range players {
		rtps[i] = players[i].totalProfit / players[i].totalPayment
	}

	if *verbose {
		log.Printf("count=%d elapsed=%v max_multiplier=%g",
			count, time.Since(start), maxMultiplier)
	}

	for _, cl := range []float64{0.90, 0.95, 0.99} {
		rtp, rtpLo, rtpHi, err := checker.ConfidenceInterval(rtps, cl)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%g %g %g %g\n", rtp, rtpLo, rtpHi, cl)

		if *verbose {
			log.Printf("%s %g%%\n", formatRangePretty(rtpLo, rtpHi), cl*100)
		}
	}
}

func formatRangePretty(lo, hi float64) string {
	if lo > hi {
		lo, hi = hi, lo
	}

	mean := (lo + hi) / 2

	p := 0.0001
	d := max((hi-lo)/2, p)
	n := 4
	for p*10 <= d {
		p *= 10
		n--
	}
	if n < 0 {
		n = 0
	}

	format := fmt.Sprintf("%%0.%df ±%%0.%df", n, n)
	return fmt.Sprintf(format, math.Round(mean/p)*p, math.Round(d/p)*p)
}

func unsafeString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
