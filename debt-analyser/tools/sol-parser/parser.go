package main

import (
	"context"
	"fmt"
	"log"
	"os"

	sitter "github.com/khalidzahra/parser/go-tree-sitter"
	sol "github.com/khalidzahra/parser/go-tree-sitter/solidity"
)

func main() {
	parser := sitter.NewParser()
	lang := sol.GetLanguage()
	parser.SetLanguage(lang)

	fileName := "C:\\Users\\Khalid zahrah\\OneDrive\\Desktop\\studies\\UofM\\COMP 7880 - Software Quality\\project\\smart-contract-analysis\\debt-analyser\\versioned-smart-contracts\\mainnet\\GenesisNationalParkLand\\0x56d8c97c33eaeed96e876d1ed206f08d1c0ad14e\\0x64af817de5c83ba529d9067bb82aa83117dccdd9_GenesisNationalParkLand_V0.sol"
	sourceCode, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	queryPattern := `(
		(comment) @comment
		.
		(function_definition) @func-def
	)`

	n, _ := sitter.ParseCtx(context.Background(), sourceCode, lang)
	query, _ := sitter.NewQuery([]byte(queryPattern), lang)

	qc := sitter.NewQueryCursor()
	qc.Exec(query, n)

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}
		// Apply predicates filtering
		m = qc.FilterPredicates(m, sourceCode)
		comment := m.Captures[0].Node.Content(sourceCode)
		fBody := m.Captures[1].Node.Content(sourceCode)
		fmt.Println(comment)
		fmt.Println(fBody)
	}

	// child := n.NamedChild(0).NamedChild(1).NamedChild(2)
	// fmt.Println(child.Type())      // lexical_declaration
	// fmt.Println(child.StartByte()) // 0
	// fmt.Println(child.EndByte())   // 9
}
