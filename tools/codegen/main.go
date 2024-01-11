package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

var buf bytes.Buffer

func main() {
	fileName := `D:\ProgramData\Go\src\mymicro\app\pkg\code\user.go`
	f, err := parser.ParseFile(token.NewFileSet(), fileName, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(&buf, "// Code generated by \"codegen %s\"; DO NOT EDIT.\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintf(&buf, "\n")
	fmt.Fprintf(&buf, "package %s\n", "code")
	fmt.Fprintf(&buf, "\n")
	fmt.Fprintf(&buf, "func init() {\n")
	genDecl(f.Decls[0].(*ast.GenDecl))
	fmt.Fprintf(&buf, "}\n")

	src := []byte(buf.String())
	err = os.WriteFile(`D:\ProgramData\Go\src\mymicro\app\pkg\code\code_generated.go`, src, 0o600)
	if err != nil {
		panic(err)
	}
}

func genDecl(decl *ast.GenDecl) {
	for _, s := range decl.Specs {
		v := s.(*ast.ValueSpec)
		for _, name := range v.Names {
			var comment string
			if v.Doc != nil && v.Doc.Text() != "" {
				comment = v.Doc.Text()
			} else if c := v.Comment; c != nil && len(c.List) == 1 {
				comment = c.Text()
			}
			httpCode, desc := ParseComment(comment)
			fmt.Fprintf(&buf, "\tregister(%s, %s, \"%s\")\n", name.Name, httpCode, desc)
			//fmt.Println(name.Name, httpCode, desc)
		}
	}
}

func ParseComment(comment string) (httpCode string, desc string) {
	reg := regexp.MustCompile(`\w\s*-\s*(\d{3})\s*:\s*([A-Z].*)\s*\.\n*`)
	if !reg.MatchString(comment) {
		return "500", "Internal server error"
	}
	groups := reg.FindStringSubmatch(comment)
	if len(groups) != 3 {
		return "500", "Internal server error"
	}
	return groups[1], groups[2]
}