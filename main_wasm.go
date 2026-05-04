//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

// findCheapestJS is exposed to JavaScript as findCheapest.
// Args: voiceMin, voiceMax, smsMin, smsMax, esimMin, esimMax, dataMin, dataMax, minGB, maxGB
// Pass -1 for a type's max to mean "no explicit limit".
func findCheapestJS(this js.Value, args []js.Value) any {
	if len(args) < 10 {
		return map[string]any{"found": false, "message": "引数が不足しています"}
	}
	voiceMin := args[0].Int()
	voiceMax := args[1].Int()
	smsMin := args[2].Int()
	smsMax := args[3].Int()
	esimMin := args[4].Int()
	esimMax := args[5].Int()
	dataMin := args[6].Int()
	dataMax := args[7].Int()
	minGB := args[8].Int()
	maxGB := args[9].Int()

	if minGB < 1 || maxGB < minGB {
		return map[string]any{"found": false, "message": "入力値が不正です"}
	}

	tc := TypeConstraints{
		{voiceMin, voiceMax},
		{smsMin, smsMax},
		{esimMin, esimMax},
		{dataMin, dataMax},
	}
	result := findCheapestMultiType(tc, minGB, maxGB)
	return comboResultToJS(result)
}

func comboResultToJS(result ComboResult) any {
	if !result.Found {
		return map[string]any{"found": false, "message": result.Message}
	}

	combosJS := make([]any, len(result.Combos))
	for i, combo := range result.Combos {
		totalGB := combo.TotalGB()

		var typesJS []any
		for t, plans := range combo {
			if len(plans) == 0 {
				continue
			}
			label := ""
			typeGB := 0
			for j, p := range plans {
				if j > 0 {
					label += " + "
				}
				label += fmt.Sprintf("%dGB(¥%d)", p.GB, p.Price)
				typeGB += p.GB
			}
			typesJS = append(typesJS, map[string]any{
				"name":    planTypeNames[t],
				"label":   label,
				"totalGB": typeGB,
				"lines":   len(plans),
			})
		}

		combosJS[i] = map[string]any{
			"totalGB":    totalGB,
			"pricePerGB": float64(result.FinalCost) / float64(totalGB),
			"types":      typesJS,
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
	select {}
}
