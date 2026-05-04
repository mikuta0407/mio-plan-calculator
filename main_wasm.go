//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func findCheapestJS(this js.Value, args []js.Value) any {
	lines := args[0].Int()
	minGB := args[1].Int()
	maxGB := args[2].Int()

	if lines < 1 || minGB < 1 || maxGB < minGB {
		return map[string]any{
			"found":   false,
			"message": "入力値が不正です",
		}
	}

	result := findCheapestCombos(lines, minGB, maxGB)
	return comboResultToJS(result)
}

func findCheapestAnyLinesJS(this js.Value, args []js.Value) any {
	minGB := args[0].Int()
	maxGB := args[1].Int()
	maxLines := args[2].Int()

	if minGB < 1 || maxGB < minGB || maxLines < 1 {
		return map[string]any{
			"found":   false,
			"message": "入力値が不正です",
		}
	}

	result := findCheapestAnyLines(minGB, maxGB, maxLines)
	return comboResultToJS(result)
}

func comboResultToJS(result ComboResult) any {
	if result.BestCost == -1 {
		return map[string]any{
			"found":   false,
			"message": "該当する組み合わせがありません",
		}
	}

	combosJS := make([]any, len(result.Combos))
	for i, combo := range result.Combos {
		totalGB := 0
		for _, p := range combo {
			totalGB += p.GB
		}
		label := ""
		for idx, p := range combo {
			if idx > 0 {
				label += " + "
			}
			label += fmt.Sprintf("%dGB(¥%d)", p.GB, p.Price)
		}
		combosJS[i] = map[string]any{
			"totalGB":    totalGB,
			"label":      label,
			"pricePerGB": float64(result.FinalCost) / float64(totalGB),
		}
	}

	return map[string]any{
		"found":     true,
		"lines":     result.Lines,
		"bestCost":  result.BestCost,
		"discount":  result.Discount,
		"finalCost": result.FinalCost,
		"combos":    combosJS,
	}
}

func main() {
	js.Global().Set("findCheapest", js.FuncOf(findCheapestJS))
	js.Global().Set("findCheapestAnyLines", js.FuncOf(findCheapestAnyLinesJS))
	select {}
}
