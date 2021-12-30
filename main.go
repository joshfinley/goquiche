// GoQuiche - Bake a go program into a gödel number quiche!
//
// Algorithm:
// 1. Parse a given go source into an AST
// 2. For each token in a declaration
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"math/big"
	"os"
	"sync"
	"text/tabwriter"
)

// Global Data

var g *big.Int = big.NewInt(0) // Godel number of the input
var p *big.Int = big.NewInt(3) // prime type with method for getting next probable prime
var v *big.Int = big.NewInt(2) // integer generator for encountered symbols
var symt symTable = symTable{  // global symbol table
	ValueTable: map[string]*big.Int{},
	PrimeTable: map[string]*big.Int{},
	mutex:      sync.RWMutex{},
}

// Type Definitions and Methods

type symTable struct {
	ValueTable map[string]*big.Int // mapping of source code symbols to their generated values. No
	PrimeTable map[string]*big.Int // mapping of symbols to their primes
	mutex      sync.RWMutex        // mutex to synchronize writes to symtable
}

func (t *symTable) addSymValue(sym string, val *big.Int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.ValueTable[sym] = val
}

func (t *symTable) addPrimeValue(sym string, val *big.Int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.ValueTable[sym] = val
}

func nextValueEntry() *big.Int {
	v.Add(v, big.NewInt(1))
	val := big.NewInt(0)
	val.Add(val, v)
	return val
}

func nextPrime() *big.Int {
	// base case
	cmpr := p.Cmp(big.NewInt(1))
	if cmpr <= 0 { // if p <= 1
		return big.NewInt(2)
	}

	candidate := big.NewInt(0)
	candidate.Add(candidate, p)
	found := false
	for !found {
		candidate.Add(candidate, big.NewInt(1))
		if candidate.ProbablyPrime(0) {
			found = true
		}
	}

	p = candidate

	return candidate
}

// Program setup and entry

// initialize logging, parse args, get ast for traversal
func setup() (*ast.File, error) {
	// setup logging to console
	log.SetOutput(os.Stdout)

	// load file argument
	filearg := flag.String("file", "min_template", "path to file to bake")
	flag.Parse()

	fi, err := os.Stat(*filearg)
	if err != nil {
		return nil, err
	}

	// parse file into ast
	set := token.NewFileSet()
	set.AddFile(*filearg, 1, int(fi.Size()))
	tree, err := parser.ParseFile(set, *filearg, nil, 0)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func main() {
	tree, err := setup()
	if err != nil {
		log.Println(err)
	}

	visit := func(n ast.Node) bool {
		g.Add(g, godeln(n))
		return err == nil
	}
	//ast.Inspect(tree, visit)

	for _, decl := range tree.Decls {
		ast.Inspect(decl, visit)
	}
}

// get the gödel number of the node
func godeln(n ast.Node) *big.Int {
	switch n.(type) {
	case *ast.GenDecl:
		return gdecl(n)
	case *ast.FuncDecl:
		return gfunc(n)
	}

	return nil
}

// get the gödel number of a GenDecl
func gdecl(n ast.Node) *big.Int {
	d := n.(*ast.GenDecl)
	switch n.(*ast.GenDecl).Tok {
	case token.CONST:
		var constPrime *big.Int
		var namePrime *big.Int
		var valuePrime *big.Int

		// check if the symbol, value, and prime for "const" are
		// in the global symbol tables
		constSym := "const"
		constVal, exists := symt.ValueTable[constSym]
		if !exists {
			constVal = nextValueEntry()
			constVal.Add(constVal, v)
			constPrime = nextPrime()
			symt.addSymValue(constSym, constVal)
			symt.addPrimeValue(constSym, constPrime)
		}

		// check if the symbol, value, and prime for the variable
		// name are in the global symbol tables
		vspec := d.Specs[0].(*ast.ValueSpec)
		nameSym := vspec.Names[0].Name
		nameVal, exists := symt.ValueTable[nameSym]
		if !exists { // add name to sym table if not exists
			nameVal = nextValueEntry()
			namePrime = nextPrime()
			symt.addSymValue(nameSym, nameVal)
			symt.addPrimeValue(nameSym, namePrime)
		}

		// check if the symbol, value, and prime for the variable's value
		// are in the global symbol tables
		vspec = d.Specs[0].(*ast.ValueSpec)
		valueSym := vspec.Values[0].(*ast.BasicLit).Value
		valueVal, exists := symt.ValueTable[valueSym]
		if !exists {
			valueVal = nextValueEntry()
			valuePrime = nextPrime()
			symt.addSymValue(valueSym, valueVal)
			symt.addPrimeValue(valueSym, valuePrime)
		}

		constGodelN := big.NewInt(0).Exp(constPrime, constVal, nil) // gödel number of 'const'
		nameGodelN := big.NewInt(0).Exp(namePrime, nameVal, nil)    // gödel number of constant name
		valueGodelN := big.NewInt(0).Exp(valuePrime, valueVal, nil) // gödel number of value
		gn := big.NewInt(0).Mul(constGodelN, nameGodelN)
		gn.Mul(gn, valueGodelN)
		//gn.Exp(prime, tok, nil)

		log.Printf("gödel number of constant declaration: %s\n", gn.String())
		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintf(writer, "\tvalue for symbol \"const\":\t%d\tprime: %s\n", constVal, constPrime.String())
		fmt.Fprintf(writer, "\tvalue for symbol \"%s\":\t%d\tprime: %s\n", nameSym, nameVal, namePrime.String())
		fmt.Fprintf(writer, "\tvalue for variable value:\t%d\tprime: %s\n", valueVal, valuePrime.String())
		writer.Flush()
		return gn
	}
	return nil
}

//  get the gödel number of func declaration
func gfunc(n ast.Node) *big.Int {
	v := n.(*ast.FuncDecl)
	//factors := make([]uint64, 0)
	if v.Recv != nil {
		//append(factors, )
		return nil
	}
	return nil
}
