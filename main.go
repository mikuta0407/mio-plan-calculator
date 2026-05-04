//go:build !js

package main

import (
	"flag"
	"fmt"
	"strings"
)

func comboLabel(combo MultiCombo) string {
	var parts []string
	for t, plans := range combo {
		if len(plans) == 0 {
			continue
		}
		s := planTypeNames[t] + ": "
		for i, p := range plans {
			if i > 0 {
				s += "+"
			}
			s += fmt.Sprintf("%dGB(¥%d)", p.GB, p.Price)
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, " / ")
}

func printResult(result ComboResult) {
	fmt.Printf("回線数: %d\n", result.Lines)
	fmt.Printf("割引: ¥%d/月 (%d回線 × ¥%d)\n", result.Discount, result.Lines, discountPerLine)
	fmt.Printf("最安値(割引後): ¥%d/月\n\n", result.FinalCost)
	for _, combo := range result.Combos {
		totalGB := combo.TotalGB()
		fmt.Printf("合計 %dGB (¥%.1f/GB): %s\n",
			totalGB,
			float64(result.FinalCost)/float64(totalGB),
			comboLabel(combo),
		)
	}
}

func printAllTable(results []ComboResult, minGB int) {
	fmt.Printf("%-8s %-6s %-14s %-12s %s\n", "合計GB", "回線数", "最安値(割引後)", "単価(/GB)", "組み合わせ")
	fmt.Println("--------------------------------------------------------------------------------")
	for i, result := range results {
		gb := minGB + i
		if !result.Found {
			fmt.Printf("%-8s %-6s %s\n", fmt.Sprintf("%dGB", gb), "-", "組み合わせなし")
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
	minGB := flag.Int("min", 25, "最低合計容量 (GB)")
	maxGB := flag.Int("max", 45, "最高合計容量 (GB)")
	all := flag.Bool("all", false, "min〜maxを1GB単位で一覧表示")

	voiceMin := flag.Int("voice-min", 0, "音声SIMの最小枚数 (上限5)")
	voiceMax := flag.Int("voice-max", -1, "音声SIMの最大枚数 (-1=制限なし、上限5)")
	smsMin := flag.Int("sms-min", 0, "SMS SIMの最小枚数")
	smsMax := flag.Int("sms-max", -1, "SMS SIMの最大枚数 (-1=制限なし)")
	esimMin := flag.Int("esim-min", 0, "データeSIMの最小枚数")
	esimMax := flag.Int("esim-max", -1, "データeSIMの最大枚数 (-1=制限なし)")
	dataMin := flag.Int("data-min", 0, "データSIMの最小枚数")
	dataMax := flag.Int("data-max", -1, "データSIMの最大枚数 (-1=制限なし)")

	flag.Parse()

	tc := TypeConstraints{
		{*voiceMin, *voiceMax},
		{*smsMin, *smsMax},
		{*esimMin, *esimMax},
		{*dataMin, *dataMax},
	}

	if *all {
		results := make([]ComboResult, *maxGB-*minGB+1)
		for i, gb := range makeRange(*minGB, *maxGB) {
			results[i] = findCheapestMultiType(tc, gb, gb)
		}
		printAllTable(results, *minGB)
		return
	}

	result := findCheapestMultiType(tc, *minGB, *maxGB)
	if !result.Found {
		fmt.Println(result.Message)
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
