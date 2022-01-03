// GoQuiche - Bake a go program into a gödel number quiche!
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
)

//
// Global Data
//

var p *big.Int = big.NewInt(3) // prime to use for current symbol
var v *big.Int = big.NewInt(2) // integer for the current symbol
var symt symTable = symTable{  // global symbol tables
	ValueTable: map[string]*big.Int{},
	PrimeTable: map[string]*big.Int{},
	GodelTable: map[string]*big.Int{},
	mutex:      sync.RWMutex{},
}

//
// Program setup and entry
//

func main() {
	tree, err := setup()
	if err != nil {
		log.Println(err)
	}

	var g *big.Int = big.NewInt(0) // gödel number of the input

	for _, decl := range tree.Decls {
		ast.Inspect(decl, func(n ast.Node) bool {
			if n == nil {
				return true // TODO figure out why this is even happening
			}
			nextg := godeln(n)
			if nextg == nil { // if a complete node has already been encountered it will not be added
				return true
			}
			if g.Cmp(big.NewInt(0)) == 0 {
				g = nextg
			}
			g.Mul(g, nextg)
			return err == nil
		})
	}

	log.Println("gödel number for program:", g.String())
	tb := make(map[string][2]string, len(symt.ValueTable))

	for k, p := range symt.PrimeTable {
		v := symt.ValueTable[k]
		a := [...]string{p.String(), v.String()}
		tb[k] = a
	}
	for _, v := range tb {
		fmt.Println(v)
	}
}

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

//
// Functions
//

// get the gödel number of the node
func godeln(n ast.Node) *big.Int {
	switch n.(type) {
	case *ast.GenDecl:
		return gdecl(n.(*ast.GenDecl))
	case *ast.FuncDecl:
		return gfuncdecl(n.(*ast.FuncDecl))
	case *ast.BinaryExpr:
		return gbinexpr(n.(*ast.BinaryExpr))
	case *ast.CallExpr:
		return gcallexpr(n.(*ast.CallExpr))
	case *ast.ExprStmt:
		return gexprstmt(n.(*ast.ExprStmt))
	case *ast.IncDecStmt:
		return gincdecstmt(n.(*ast.IncDecStmt))
	case *ast.ReturnStmt:
		return gretstmt(n.(*ast.ReturnStmt))
	case *ast.BasicLit:
		return glit(n.(*ast.BasicLit))
	case *ast.Ident:
		return gident(n.(*ast.Ident))
	case *ast.FuncType:
		return gfunctype(n.(*ast.FuncType))
	case *ast.ValueSpec:
		return nil
	case *ast.FieldList:
		return gfieldlist(n.(*ast.FieldList))
	case *ast.Field:
		return gfield(n.(*ast.Field))
	default:
		fmt.Println("type not implemented")
	}

	return nil
}

func gfieldlist(l *ast.FieldList) *big.Int {
	return nil
}

func gfield(f *ast.Field) *big.Int {
	tgn := godeln(f.Type)
	if len(f.Names) == 0 {
		return tgn
	}

	gns := make([]*big.Int, 0)
	for _, name := range f.Names {
		gns = append(gns, godeln(name))
	}
	if len(gns) != 0 {
		if len(gns) > 1 {
			for i := 0; i < len(gns)-1; i++ {
				tgn.Mul(gns[i], gns[i+1])
			}
		} else if len(gns) == 1 {
			tgn.Mul(tgn, gns[0])
		}
	}
	return tgn
}

func gfunctype(f *ast.FuncType) *big.Int {
	gns := make([]*big.Int, 0)
	if len(f.Params.List) != 0 {
		paramList := f.Params.List[0].Names

		for _, param := range paramList {
			gns = append(gns, godeln(param))
		}
	}

	if f.Results != nil {
		resl := f.Results.List

		for _, res := range resl {
			gns = append(gns, godeln(res))
		}
	}

	gn := big.NewInt(0)
	if len(gns) != 0 {
		if len(gns) > 1 {
			for i := 0; i < len(gns)-1; i++ {
				gn.Mul(gns[i], gns[i+1])
			}
		} else if len(gns) == 1 {
			gn.Mul(gn, gns[0])
		}

	}
	return nil
}

func gincdecstmt(n *ast.IncDecStmt) *big.Int {
	sym := "++"
	var expr string
	switch n.X.(type) {
	case *ast.Ident:
		expr = n.X.(*ast.Ident).Name
	default:
		log.Println("type not implemented: ")
		return nil
	}
	_, _, symGn := symt.add(sym)
	_, _, exprGn := symt.add(expr)
	stmt := expr + sym
	stmtGn := symGn.Mul(symGn, exprGn)
	var gns = [...]*big.Int{symGn, exprGn}
	symt.addCompound(stmt, gns[:])
	return stmtGn

}

func gretstmt(n *ast.ReturnStmt) *big.Int {
	symRet := "return"
	_, _, symGn := symt.add(symRet)

	gns := make([]*big.Int, 0)
	for _, res := range n.Results {
		gns = append(gns, godeln(res))
	}

	if len(gns) > 1 {
		for i := 0; i < len(gns)-1; i++ {
			symGn.Mul(gns[i], gns[i+1])
		}
	} else if len(gns) == 1 {
		symGn.Mul(symGn, gns[0])
	}

	return symGn
}

func gident(i *ast.Ident) *big.Int {
	symVal, symPrime, symGn := symt.add(i.Name)
	symGn.Exp(symPrime, symVal, nil) // gödel number of ident (should already exist?)
	return symGn
}

// get the gödel number of a GenDecl
func gdecl(d *ast.GenDecl) *big.Int {
	switch d.Tok {
	case token.CONST:
		return gvardec(d)
	case token.VAR:
		return gvardec(d)
	}
	return nil
}

func glit(e *ast.BasicLit) *big.Int {
	sym := e.Value
	symVal, symPrime, gn := symt.add(sym)
	gn.Exp(symPrime, symVal, nil) // gödel number of literal
	return gn
}

func gbinexpr(e *ast.BinaryExpr) *big.Int {
	var opSym string
	switch e.Op {
	case token.ADD:
		opSym = "+"
	case token.SUB:
		opSym = "-"
	case token.MUL:
		opSym = "*"
	case token.QUO:
		opSym = "/"
	case token.REM:
		opSym = "%"
	case token.EQL:
		opSym = "=="
	default:
		log.Panicln("encountered unknown binaryexpr token:", e.Op)
	}
	_, _, opGn := symt.add(opSym)

	gns := make([]*big.Int, 0)
	gns = append(gns, opGn)
	gns = append(gns, godeln(e.X))
	gns = append(gns, godeln(e.Y))
	gn := big.NewInt(0)

	for i := 0; i < len(gns)-1; i++ {
		gn.Mul(gns[i], gns[i+1])
	}

	return gn
}

func gcallexpr(e *ast.CallExpr) *big.Int {
	nameSym := e.Fun.(*ast.Ident).Name
	if nameSym == "" {
		log.Panicln("cannot yet handle anonymous functions")
	}

	_, _, gn := symt.add(nameSym)
	var gns = make([]*big.Int, 0)
	gns = append(gns, gn)

	if len(e.Args) != 0 {
		for _, arg := range e.Args {
			gns = append(gns, godeln(arg))
		}
	}

	for i := 0; i < len(gns)-1; i++ {
		gn.Mul(gns[i], gns[i+1])
	}

	return gn
}

func gexprstmt(n *ast.ExprStmt) *big.Int {
	switch n.X.(type) {
	case *ast.CallExpr:
		return gcallexpr(n.X.(*ast.CallExpr))
	}
	return nil
}

func gfuncdecl(e *ast.FuncDecl) *big.Int {
	funcSym := "func"
	_, _, gn := symt.add(funcSym)

	gns := make([]*big.Int, 0)
	if len(e.Type.Params.List) != 0 {
		paramList := e.Type.Params.List[0].Names

		for _, param := range paramList {
			gns = append(gns, godeln(param))
		}
	}

	if len(e.Body.List) != 0 {
		for _, item := range e.Body.List {
			gns = append(gns, godeln(item))
		}
	}

	nameSym := e.Name.Name
	if nameSym == "" {
		log.Panicln("cannot yet handle anonymous functions")
	}
	_, _, nameGn := symt.add(nameSym)
	gn.Mul(gn, nameGn)

	for i := 0; i < len(gns)-1; i++ {
		gn.Mul(gns[i], gns[i+1])
	}

	// //factors := make([]uint64, 0)
	// if e.Recv != nil {
	// 	//append(factors, )
	// 	return nil
	// }
	return gn
}

func gvardec(d *ast.GenDecl) *big.Int {
	var varSym string

	switch d.Tok {
	case token.CONST:
		varSym = "const"
	case token.VAR:
		varSym = "var"
	}

	_, _, varGn := symt.add(varSym)

	// check if the symbol, value, and prime for the variable
	// name are in the global symbol tables
	nameSym := d.Specs[0].(*ast.ValueSpec).Names[0].Name
	_, _, nameGn := symt.add(nameSym)

	// check if the symbol, value, and prime for the variable's value
	// are in the global symbol tables
	valueSym := d.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
	_, _, valueGn := symt.add(valueSym)

	gn := big.NewInt(0).Mul(varGn, nameGn)
	gn.Mul(gn, valueGn)

	return gn
}

//
// Type Definitions and Methods
//

type symTable struct {
	ValueTable map[string]*big.Int // mapping of source code symbols to their generated values. No
	PrimeTable map[string]*big.Int // mapping of symbols to their primes
	GodelTable map[string]*big.Int // mapping of code strings to godel numbers
	mutex      sync.RWMutex        // mutex to synchronize writes to symtable
}

func (t *symTable) add(sym string) (*big.Int, *big.Int, *big.Int) {
	symVal, exists := t.ValueTable[sym]
	symPrime := t.PrimeTable[sym]
	gn := t.GodelTable[sym]
	// add the symbol to the tables if not exist
	if !exists {
		symVal = nextValueEntry()
		symPrime = nextPrime()
		gn = big.NewInt(0).Exp(symPrime, symVal, big.NewInt(0))
		t.addSym(sym, symVal)
		t.addPrime(sym, symPrime)
		t.addGodeln(sym, gn)
		fmt.Printf("symbol '%s' is %s to the %s is %s\n",
			sym, symPrime.String(), symVal.String(), gn)
	}
	return symVal, symPrime, gn
}

func (t *symTable) addCompound(sym string, gns []*big.Int) *big.Int {
	if len(gns) == 1 {
		return nil
	}

	symVal, exists := t.ValueTable[sym]
	symPrime := t.PrimeTable[sym]
	gn := big.NewInt(0)
	// add the symbol to the tables if not exist
	if !exists {
		symVal = nextValueEntry()
		symPrime = nextPrime()

		curGn := gns[0]
		curPrime := nextPrime()
		nextGn := gns[1]
		nextPrime := nextPrime()

		newGn1 := big.NewInt(0).Exp(curPrime, curGn, big.NewInt(0))
		newGn2 := big.NewInt(0).Exp(nextPrime, nextGn, big.NewInt(0))
		gn := big.NewInt(0).Mul(newGn1, newGn2)

		t.addSym(sym, symVal)
		t.addPrime(sym, symPrime)
		t.addGodeln(sym, gn)
		fmt.Printf("symbol '%s' is %s to the %s is %s...(%d digits ommitted)\n",
			sym, symPrime.String(), symVal.String(), gn.String()[0:10], len(gn.String())-10)
	}
	return gn
}

func (t *symTable) exists(sym string) bool {
	_, exists := t.ValueTable[sym]
	return exists
}

func (t *symTable) addSym(sym string, val *big.Int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.ValueTable[sym] = val
}

func (t *symTable) addPrime(sym string, val *big.Int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.PrimeTable[sym] = val
}

func (t *symTable) addGodeln(sym string, val *big.Int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.GodelTable[sym] = val
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

// // Get all prime factors of a given number n
// func primefactors(n *big.Int) []*big.Int {
// 	fx := make([]*big.Int, 0)
// 	// Get the number of 2s that divide n
// 	twos := big.NewInt(0)
// 	for {
// 		twos = n.Mod(n, big.NewInt(2))
// 		if twos.Cmp(big.NewInt(0)) != 0 {
// 			break
// 		}
// 		fx = append(fx, big.NewInt(2))
// 		n.Div(n, big.NewInt(2))
// 	}

// 	// n must be odd at this point. so we can skip one element
// 	// (note i = i + 2)
// 	for i := 3; i*i <= n; i = i + 2 {
// 		// while i divides n, append i and divide n
// 		for n%i == 0 {
// 			pfs = append(pfs, i)
// 			n = n / i
// 		}
// 	}

// 	// This condition is to handle the case when n is a prime number
// 	// greater than 2
// 	if n > 2 {
// 		pfs = append(pfs, n)
// 	}

// 	return
// }
