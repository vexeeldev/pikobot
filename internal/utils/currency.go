package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func FormatRupiah(amount int64) string {
	isNegative := amount < 0
	if isNegative {
		amount = -amount
	}

	str := strconv.FormatInt(amount, 10)
	n := len(str)
	if n <= 3 {
		if isNegative {
			return "-Rp " + str
		}
		return "Rp " + str
	}

	var result strings.Builder
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(str[:remainder])
		if n > remainder {
			result.WriteString(".")
		}
	}

	for i := remainder; i < n; i += 3 {
		result.WriteString(str[i : i+3])
		if i+3 < n {
			result.WriteString(".")
		}
	}

	if isNegative {
		return "-Rp " + result.String()
	}
	return "Rp " + result.String()
}

func ParseRupiah(input string) (int64, error) {
	clean := strings.TrimSpace(strings.ToLower(input))
	clean = strings.ReplaceAll(clean, "rp", "")
	clean = strings.ReplaceAll(clean, " ", "")

	if strings.HasSuffix(clean, "k") || strings.HasSuffix(clean, "rb") || strings.HasSuffix(clean, "ribu") {
		clean = strings.TrimSuffix(clean, "ribu")
		clean = strings.TrimSuffix(clean, "rb")
		clean = strings.TrimSuffix(clean, "k")
		clean = strings.ReplaceAll(clean, ",", ".")
		val, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return 0, fmt.Errorf("nominal tidak valid: %s", input)
		}
		return int64(val * 1000), nil
	}

	if strings.HasSuffix(clean, "jt") || strings.HasSuffix(clean, "juta") {
		clean = strings.TrimSuffix(clean, "juta")
		clean = strings.TrimSuffix(clean, "jt")
		clean = strings.ReplaceAll(clean, ",", ".")
		val, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return 0, fmt.Errorf("nominal tidak valid: %s", input)
		}
		return int64(val * 1000000), nil
	}

	clean = strings.ReplaceAll(clean, ".", "")
	clean = strings.ReplaceAll(clean, ",", "")

	re := regexp.MustCompile(`^-?\d+$`)
	if !re.MatchString(clean) {
		return 0, fmt.Errorf("format nominal tidak dikenali: %s", input)
	}

	val, err := strconv.ParseInt(clean, 10, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}
