// Package staticlint содержит линтеры для статической проверки кода
// 	Собираем: go build staticlint.go
//  Копируем в корень проекта: cp staticlint ../.. && cd ../..
//	Запускаем: ./staticlint ./...
// Используемые линтеры:
//	- analysis/passes/
//	  - printf		- Проверяет согласованость строк и аргумендов для printf
//	  - structtag   - Проверяет что бы теги полей структур соответствовали reflect.StructTag.Get
//
//	- statickcheck.io:
//	  - Все линтеры класса SA - проверяет код на проблемы с производительностью
//	  - Анализаторы класса QF - содержат альтернативный взгляд на рефакторинг кода
//	  - Анализаторы класса S  - помогают сделать код проще
//	  - Анализаторы класса ST - проверяют оформление кода
// - Собственые:
//	 - osExitanalizer - анализатор запрещающий прямой вызов os.Exist в функции main пакета main
package main

import (
	"staticlint/customLinter"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	shortURLChecker()
}

func shortURLChecker() {
	// Список линтеров
	linters := map[string]bool{
		"QF1004": true, // `Use \'strings.ReplaceAll\' instead of \'strings.Replace\' with \'n == -1\'`
		"S1000":  true, // `Use plain channel send or receive instead of single-case select`,
		"QF1001": true, // Apply De Morgan's law
		"ST1003": true, // `Poorly chosen identifier`
	}

	// Создаем слайс и наполняем его нужными линтерами
	var myChecker []*analysis.Analyzer

	// Добавляем все анализаторы класса SA
	for _, v := range staticcheck.Analyzers {
		myChecker = append(myChecker, v.Analyzer)
	}

	// Добавляем анализаторы класса QF
	for _, v := range quickfix.Analyzers {
		if linters[v.Analyzer.Name] {
			myChecker = append(myChecker, v.Analyzer)
		}
	}

	// Добавляем анализаторы класса S
	for _, v := range simple.Analyzers {
		if linters[v.Analyzer.Name] {
			myChecker = append(myChecker, v.Analyzer)
		}
	}

	// Добавляем анализаторы класса  ST
	for _, v := range stylecheck.Analyzers {
		if linters[v.Analyzer.Name] {
			myChecker = append(myChecker, v.Analyzer)
		}
	}

	// Добавляем линтеры из go/analysis
	myChecker = append(myChecker, printf.Analyzer)
	myChecker = append(myChecker, structtag.Analyzer)

	// Добавляем собственный анализатор
	myChecker = append(myChecker, customLinter.OsExitAnalyzer)

	// Подключаем в мультичекер
	multichecker.Main(
		myChecker...,
	)
}
