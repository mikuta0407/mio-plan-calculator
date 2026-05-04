//go:build !js

package main

import (
	"flag"
	"fmt"
)

func comboLabel(combo []Plan) string {
	label := ""
	for idx, p := range combo {
		if idx > 0 {
			label += " + "
		}
		label += fmt.Sprintf("%dGB(¥%d)", p.GB, p.Price)
	}
	return label
}

func printResult(result ComboResult) {
	fmt.Printf("割引: ¥%d/月 (%d回線 × ¥%d)\n", result.Discount, result.Lines, discountPerLine)
	fmt.Printf("最安値(割引後): ¥%d/月\n\n", result.FinalCost)
	for _, combo := range result.Combos {
		totalGB := 0
		for _, p := range combo {
			totalGB += p.GB
		}
		fmt.Printf("合計 %dGB (¥%.1f/GB): %s\n",
			totalGB,
			float64(result.FinalCost)/float64(totalGB),
			comboLabel(combo),
		)
	}
}

func printAllTable(results []ComboResult, minGB int) {
	fmt.Printf("%-8s %-6s %-14s %-12s %s\n", "合計GB", "回線数", "最安値(割引後)", "単価(/GB)", "組み合わせ")
	fmt.Println("------------------------------------------------------------------------")
	for i, result := range results {
		gb := minGB + i
		if result.BestCost == -1 {
			fmt.Printf("%-8s %-6s %-14s\n", fmt.Sprintf("%dGB", gb), "-", "組み合わせなし")
			continue
		}
		perGB := float64(result.FinalCost) / float64(gb)
		for j, combo := range result.Combos {
			if j == 0 {
				fmt.Printf("%-8s %-6d ¥%-13d ¥%-11.1f %s\n",
					fmt.Sprintf("%dGB", gb), result.Lines, result.FinalCost, perGB, comboLabel(combo))
			} else {
				fmt.Printf("%-8s %-6s %-14s %-12s %s\n", "", "", "", "", comboLabel(combo))
			}
		}
	}
}

func main() {
	lines := flag.Int("lines", 4, "回線数（-freeモードでは無視）")
	minGB := flag.Int("min", 25, "最低容量(GB)")
	maxGB := flag.Int("max", 45, "最高容量(GB)")
	all := flag.Bool("all", false, "min〜maxを1GB単位で一覧表示")
	free := flag.Bool("free", false, "回線数を固定せず最安値を探す")
	maxLines := flag.Int("maxlines", 10, "-freeモード時に試す最大回線数")
	flag.Parse()

	if *free {
		if *all {
			results := make([]ComboResult, *maxGB-*minGB+1)
			for i, gb := range makeRange(*minGB, *maxGB) {
				results[i] = findCheapestAnyLines(gb, gb, *maxLines)
			}
			printAllTable(results, *minGB)
			return
		}
		result := findCheapestAnyLines(*minGB, *maxGB, *maxLines)
		if result.BestCost == -1 {
			fmt.Println("該当する組み合わせがありません")
			return
		}
		printResult(result)
		return
	}

	if *all {
		results := make([]ComboResult, *maxGB-*minGB+1)
		for i, gb := range makeRange(*minGB, *maxGB) {
			results[i] = findCheapestCombos(*lines, gb, gb)
		}
		printAllTable(results, *minGB)
		return
	}

	result := findCheapestCombos(*lines, *minGB, *maxGB)
	if result.BestCost == -1 {
		fmt.Println("該当する組み合わせがありません")
		return
	}
	printResult(result)
}

func makeRange(min, max int) []int {
	s := make([]int, max-min+1)
	for i := range s {
		s[i] = min + i
	}
	return s
}
