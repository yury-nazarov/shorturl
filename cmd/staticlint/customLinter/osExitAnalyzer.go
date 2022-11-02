package customLinter

import (
	"fmt"
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osExitCheck",
	Doc:  "Check call os.Exit()\n",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Обходим каждую верширу AST дерева
	for _, file := range pass.Files {
		if strings.Contains(file.Name.Name, "main") && strings.Contains(pass.Pkg.String(), "main") {
			ast.Inspect(file, func(n ast.Node) bool {
				// проверяем, какой конкретный тип лежит в узле
				switch x := n.(type) {
				// ast.CallExpr представляет вызов функции или метода
				case *ast.CallExpr:
					stringFun := fmt.Sprintf("%s", x.Fun)
					// Тут будет что то вроде: &{fmt Println}&{os Exit}%
					// я так и не нашел способа добратся до объекта содержащего вызов os.Exit кроме как через split
					// и операции со строкой
					callMethodsList := strings.Split(stringFun, "&")
					// Если в file есть вызов os.Exit - сообщаем об этом
					for _, callMethod := range callMethodsList {
						if strings.Contains(callMethod, "os Exit") {
							msg := fmt.Sprintf("package %s function %s containt call os.Exit\n", file.Name, file.Name.Name)
							pass.Reportf(x.Fun.Pos(), msg)
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
