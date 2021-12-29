// GoQuiche - Bake a go program into a gödel number quiche!
//
// Algorithm:
// 1. Parse a given go source into an AST
// 2. For each token in a declaration
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"

	"github.com/joshfinley/goquiche/prime"
)

var p prime.Prime = 1

type GodelNum struct {
	Num []uint64
	idx int
}

// Write a slice of uint64, extending if necessary
func (n *GodelNum) Write(val uint64) error {
	if cap(n.Num) <= n.idx {
		news := make([]uint64, cap(n.Num)+1024) // extend by 1024 entries
		ncop := copy(news, n.Num)
		if ncop != cap(n.Num) {
			return fmt.Errorf("failed to extend num buffer")
		}
		n.Num = news
	}
	n.Num[n.idx] = val
	return nil
}

// get the gödel number of the node
func godeln(n ast.Node) uint64 {
	switch n.(type) {
	case *ast.GenDecl:
		return gdecl(n)
	}
	return 0
}

func gdecl(n ast.Node) uint64 {
	switch n.(*ast.GenDecl).Tok {
	case token.CONST:
		return uint64(math.Pow(float64(p.Next()), 6))
	}
	return 0
}

func main() {
	fi, err := os.Stat("quine_template")
	if err != nil {
		os.Exit(1)
	}
	set := token.NewFileSet()
	set.AddFile("quine_template", 1, int(fi.Size()))
	file, err := parser.ParseFile(set, "quine_template", nil, 0)
	if err != nil {
		os.Exit(1)
	}

	gn := GodelNum{
		Num: []uint64{},
		idx: 0,
	}

	ast.Inspect(file, func(n ast.Node) bool {
		err := gn.Write(godeln(n))
		return err == nil
	})

	for _, decl := range file.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			fmt.Println(decl)
		}
	}
	// for _, decl := range file.Decls {
	// 	fmt.Println(decl.Pos())
	// 	fmt.Println(decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0].Name)
	// 	godeln(decl.(*ast.GenDecl))
	// }
}
