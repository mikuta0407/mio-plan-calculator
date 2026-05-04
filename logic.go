package main

// Plan holds the GB capacity and monthly price for a single SIM plan.
type Plan struct {
	GB    int
	Price int
}

const (
	PlanTypeVoice    = 0 // 音声SIM
	PlanTypeSMS      = 1 // SMS SIM
	PlanTypeDataESIM = 2 // データeSIM
	PlanTypeData     = 3 // データSIM
	NumPlanTypes     = 4
)

var planTypeNames = [NumPlanTypes]string{"音声", "SMS", "データeSIM", "データ"}

// allPlans contains available plans per SIM type, sorted by GB (and price) ascending.
var allPlans = [NumPlanTypes][]Plan{
	{ // 音声SIM
		{2, 850}, {5, 950}, {10, 1400}, {15, 1600},
		{25, 2000}, {35, 2400}, {45, 3300}, {55, 3900},
	},
	{ // SMS SIM
		{2, 820}, {5, 930}, {10, 1370}, {15, 1580},
		{25, 1980}, {35, 2380}, {45, 3280}, {55, 3880},
	},
	{ // データeSIM
		{2, 440}, {5, 650}, {10, 1050}, {15, 1320},
		{25, 1650}, {35, 2240}, {45, 2940}, {55, 3540},
	},
	{ // データSIM
		{2, 740}, {5, 860}, {10, 1300}, {15, 1530},
		{25, 1950}, {35, 2340}, {45, 3240}, {55, 3840},
	},
}

const (
	discountPerLine = 100
	maxVoiceLines   = 5
	maxTotalLines   = 10
)

// LineConstraint holds the min and max line count for one SIM type.
// Max == -1 means no explicit user limit; the system default applies.
type LineConstraint struct {
	Min int
	Max int
}

// TypeConstraints holds per-type line count constraints.
type TypeConstraints [NumPlanTypes]LineConstraint

// effectiveMax returns the clamped maximum for type t.
func (tc TypeConstraints) effectiveMax(t int) int {
	sysMax := maxTotalLines
	if t == PlanTypeVoice {
		sysMax = maxVoiceLines
	}
	m := tc[t].Max
	if m < 0 || m > sysMax {
		return sysMax
	}
	return m
}

// MultiCombo stores the selected plans for each SIM type in one combination.
type MultiCombo [NumPlanTypes][]Plan

// TotalGB returns the total GB across all types.
func (mc MultiCombo) TotalGB() int {
	total := 0
	for _, plans := range mc {
		for _, p := range plans {
			total += p.GB
		}
	}
	return total
}

// ComboResult is the output of a cheapest-combination search.
type ComboResult struct {
	Found     bool
	Message   string
	Lines     int
	BestCost  int
	Discount  int
	FinalCost int
	Combos    []MultiCombo
}

// findCheapestMultiType finds the cheapest plan combination satisfying the
// per-type line constraints and the total-GB range [minGB, maxGB].
func findCheapestMultiType(tc TypeConstraints, minGB, maxGB int) ComboResult {
	bestFinalCost := -1
	var bestCombos []MultiCombo

	maxV := tc.effectiveMax(PlanTypeVoice)
	maxS := tc.effectiveMax(PlanTypeSMS)
	maxDE := tc.effectiveMax(PlanTypeDataESIM)
	maxD := tc.effectiveMax(PlanTypeData)

	for nV := tc[PlanTypeVoice].Min; nV <= maxV; nV++ {
		for nS := tc[PlanTypeSMS].Min; nS <= maxS && nV+nS <= maxTotalLines; nS++ {
			for nDE := tc[PlanTypeDataESIM].Min; nDE <= maxDE && nV+nS+nDE <= maxTotalLines; nDE++ {
				for nD := tc[PlanTypeData].Min; nD <= maxD && nV+nS+nDE+nD <= maxTotalLines; nD++ {
					total := nV + nS + nDE + nD
					if total == 0 {
						continue
					}
					discount := total * discountPerLine
					upperBound := -1
					if bestFinalCost != -1 {
						upperBound = bestFinalCost + discount
					}
					counts := [NumPlanTypes]int{nV, nS, nDE, nD}
					cost, combos := searchPlans(counts, minGB, maxGB, upperBound)
					if cost == -1 {
						continue
					}
					finalCost := cost - discount
					if bestFinalCost == -1 || finalCost < bestFinalCost {
						bestFinalCost = finalCost
						bestCombos = combos
					} else if finalCost == bestFinalCost {
						bestCombos = append(bestCombos, combos...)
					}
				}
			}
		}
	}

	if bestFinalCost == -1 {
		return ComboResult{Found: false, Message: "該当する組み合わせがありません"}
	}

	totalLines := 0
	for _, plans := range bestCombos[0] {
		totalLines += len(plans)
	}
	return ComboResult{
		Found:     true,
		Lines:     totalLines,
		BestCost:  bestFinalCost + totalLines*discountPerLine,
		Discount:  totalLines * discountPerLine,
		FinalCost: bestFinalCost,
		Combos:    bestCombos,
	}
}

// searchPlans returns the minimum-cost plans and all equal-cost MultiCombos for
// the given per-type counts with total GB in [minGB, maxGB].
// upperBound prunes branches whose cost already exceeds it (−1 = no limit).
func searchPlans(counts [NumPlanTypes]int, minGB, maxGB, upperBound int) (int, []MultiCombo) {
	// Precompute suffix min/max GB for GB-range pruning.
	var suffixMinGB [NumPlanTypes + 1]int
	var suffixMaxGB [NumPlanTypes + 1]int
	for t := NumPlanTypes - 1; t >= 0; t-- {
		p := allPlans[t]
		suffixMinGB[t] = suffixMinGB[t+1] + counts[t]*p[0].GB
		suffixMaxGB[t] = suffixMaxGB[t+1] + counts[t]*p[len(p)-1].GB
	}

	firstType := -1
	for t := 0; t < NumPlanTypes; t++ {
		if counts[t] > 0 {
			firstType = t
			break
		}
	}
	if firstType == -1 {
		return -1, nil
	}

	bestCost := -1
	var bestCombos []MultiCombo
	var current MultiCombo
	for t, n := range counts {
		if n > 0 {
			current[t] = make([]Plan, 0, n)
		}
	}

	var fill func(typeIdx, planStart, remaining, cost, gb int)
	fill = func(typeIdx, planStart, remaining, cost, gb int) {
		if remaining == 0 {
			nextType := typeIdx + 1
			for nextType < NumPlanTypes && counts[nextType] == 0 {
				nextType++
			}
			if nextType >= NumPlanTypes {
				if gb >= minGB && gb <= maxGB {
					if bestCost == -1 || cost < bestCost {
						bestCost = cost
						bestCombos = []MultiCombo{copyMultiCombo(current)}
					} else if cost == bestCost {
						bestCombos = append(bestCombos, copyMultiCombo(current))
					}
				}
				return
			}
			fill(nextType, 0, counts[nextType], cost, gb)
			return
		}

		plans := allPlans[typeIdx]
		maxGBPerPlan := plans[len(plans)-1].GB
		// Maximum additional GB after picking one plan here: (remaining−1) future picks
		// in this type (at most maxGBPerPlan each) + all subsequent types.
		maxFutureGB := (remaining-1)*maxGBPerPlan + suffixMaxGB[typeIdx+1]

		for i := planStart; i < len(plans); i++ {
			p := plans[i]
			newCost := cost + p.Price
			newGB := gb + p.GB

			// Cost pruning: plans are sorted by price asc; once over limit, stop.
			if upperBound >= 0 && newCost > upperBound {
				break
			}

			// GB pruning (non-decreasing order means future picks are >= p.GB).
			minFutureGB := (remaining-1)*p.GB + suffixMinGB[typeIdx+1]
			if newGB+maxFutureGB < minGB {
				// Even maximum future GB can't reach minGB; try a larger plan.
				continue
			}
			if newGB+minFutureGB > maxGB {
				// Even minimum future GB already exceeds maxGB; all larger plans worse.
				break
			}

			current[typeIdx] = append(current[typeIdx], p)
			fill(typeIdx, i, remaining-1, newCost, newGB)
			current[typeIdx] = current[typeIdx][:len(current[typeIdx])-1]
		}
	}

	fill(firstType, 0, counts[firstType], 0, 0)
	return bestCost, bestCombos
}

func copyMultiCombo(src MultiCombo) MultiCombo {
	var dst MultiCombo
	for t := range src {
		if len(src[t]) > 0 {
			dst[t] = make([]Plan, len(src[t]))
			copy(dst[t], src[t])
		}
	}
	return dst
}
