package main

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"strings"
)

func generatePrimes(n int) []int {
	primes := []int{2}
	num := 3
	for len(primes) < n {
		isPrime := true
		for _, prime := range primes {
			if num%prime == 0 {
				isPrime = false
				break
			}
			if prime*prime > num {
				break
			}
		}
		if isPrime {
			primes = append(primes, num)
		}
		num += 2
	}
	return primes
}

func recoverSymbols(result *big.Int, symbolPrimes map[string]int) []string {
	// Create a reverse mapping of primes to symbols
	primeToSymbol := make(map[int]string)
	for symbol, prime := range symbolPrimes {
		primeToSymbol[prime] = symbol
	}

	// Sort primes in descending order
	primes := make([]int, 0, len(symbolPrimes))
	for _, prime := range symbolPrimes {
		primes = append(primes, prime)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(primes)))

	recoveredSymbols := []string{}
	n := new(big.Int).Set(result)
	zero := big.NewInt(0)

	for n.Cmp(big.NewInt(1)) > 0 {
		for _, prime := range primes {
			primeBig := big.NewInt(int64(prime))
			exponent := 0
			for new(big.Int).Mod(n, primeBig).Cmp(zero) == 0 {
				n.Div(n, primeBig)
				exponent++
			}
			if exponent > 0 {
				symbol := primeToSymbol[prime]
				recoveredSymbols = append(recoveredSymbols, symbol)
				break
			}
		}
	}

	// Reverse the recovered symbols to get the correct order
	for i := 0; i < len(recoveredSymbols)/2; i++ {
		j := len(recoveredSymbols) - 1 - i
		recoveredSymbols[i], recoveredSymbols[j] = recoveredSymbols[j], recoveredSymbols[i]
	}

	return recoveredSymbols
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run program.go <input_file.go>")
		return
	}

	// Read the input file
	src, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Initialize the scanner
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile(os.Args[1], fset.Base(), len(src))
	s.Init(file, src, nil, scanner.ScanComments)

	// Initialize prime number generator
	primes := generatePrimes(1000) // Generate first 1000 primes

	// Map to store symbol-prime associations and symbol order
	symbolPrimes := make(map[string]int)
	symbolOrder := make([]string, 0)

	// Perform lexical analysis
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		symbol := tok.String()
		if lit != "" {
			symbol = lit
		}

		// Replace newline characters with \n
		symbol = strings.ReplaceAll(symbol, "\n", "\\n")

		if _, exists := symbolPrimes[symbol]; !exists {
			symbolPrimes[symbol] = primes[len(symbolOrder)]
			symbolOrder = append(symbolOrder, symbol)
		}

		fmt.Printf("%s\t%s\t%d\n", fset.Position(pos), symbol, symbolPrimes[symbol])
	}

	// Print the number line with symbols
	fmt.Println("\nNumber line with symbols:")
	for i, symbol := range symbolOrder {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%d:%s", i+1, symbol)
	}
	fmt.Println()

	// Print the number line with prime numbers
	fmt.Println("\nNumber line with prime numbers:")
	for i, symbol := range symbolOrder {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Printf("%d:%d", i+1, symbolPrimes[symbol])
	}
	fmt.Println()

	// Print the number line with exponentiated values
	fmt.Println("\nNumber line with exponentiated values:")
	result := big.NewInt(1)
	for i, symbol := range symbolOrder {
		if i > 0 {
			fmt.Print(" * ")
		}
		prime := big.NewInt(int64(symbolPrimes[symbol]))
		exponent := big.NewInt(int64(i + 1))
		value := new(big.Int).Exp(prime, exponent, nil)
		fmt.Printf("%s^%d", symbol, i+1)
		result.Mul(result, value)
	}
	fmt.Println()
	fmt.Printf("\nFinal result: %s\n", result.String())

	// Recover original symbols
	recoveredSymbols := recoverSymbols(result, symbolPrimes)
	fmt.Println("\nRecovered symbols:")
	for i, symbol := range recoveredSymbols {
		fmt.Printf("%d: %s\n", i+1, symbol)
	}
}
