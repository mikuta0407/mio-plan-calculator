package main

type Plan struct {
	GB    int
	Price int
}

var plans = []Plan{
	{2, 850},
	{5, 950},
	{10, 1400},
	{15, 1600},
	{25, 2000},
	{35, 2400},
	{45, 3300},
	{55, 3900},
}

const discountPerLine = 100

type ComboResult struct {
	Lines     int
	BestCost  int
	Discount  int
	FinalCost int
	Combos    [][]Plan
}

func findCheapestCombos(lines, minGB, maxGB int) ComboResult {
	bestCost := -1
	var bestCombos [][]Plan

	var search func(startIdx, remaining int, current []Plan, currentCost, currentGB int)
	search = func(startIdx, remaining int, current []Plan, currentCost, currentGB int) {
		if remaining == 0 {
			if currentGB >= minGB && currentGB <= maxGB {
				if bestCost == -1 || currentCost < bestCost {
					bestCost = currentCost
					bestCombos = [][]Plan{append([]Plan(nil), current...)}
				} else if currentCost == bestCost {
					bestCombos = append(bestCombos, append([]Plan(nil), current...))
				}
			}
			return
		}
		for i := startIdx; i < len(plans); i++ {
			search(i, remaining-1, append(current, plans[i]), currentCost+plans[i].Price, currentGB+plans[i].GB)
		}
	}
	search(0, lines, nil, 0, 0)

	discount := lines * discountPerLine
	finalCost := -1
	if bestCost != -1 {
		finalCost = bestCost - discount
	}
	return ComboResult{
		Lines:     lines,
		BestCost:  bestCost,
		Discount:  discount,
		FinalCost: finalCost,
		Combos:    bestCombos,
	}
}

// findCheapestAnyLines finds the cheapest combination for the given GB range
// without fixing the number of lines (tries 1..maxLines).
func findCheapestAnyLines(minGB, maxGB, maxLines int) ComboResult {
	best := ComboResult{BestCost: -1}
	for lines := 1; lines <= maxLines; lines++ {
		r := findCheapestCombos(lines, minGB, maxGB)
		if r.BestCost == -1 {
			continue
		}
		if best.BestCost == -1 || r.FinalCost < best.FinalCost {
			best = r
		}
	}
	return best
}
