package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/patrikeh/go-deep/training"
)

func loadAndProcessData(filename string) training.Examples {
	file, _ := os.Open(filename)
	defer file.Close()

	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()

	var examples training.Examples

	for i, row := range records {
		if i == 0 {
			continue
		}

		age, _ := strconv.ParseFloat(row[0], 64)
		exp, _ := strconv.ParseFloat(row[1], 64)
		score, _ := strconv.ParseFloat(row[4], 64)

		eduOH := educationOneHot(row[2])
		depOH := departmentOneHot(row[3])

		input := append([]float64{
			normalize(age, 20, 60),
			normalize(exp, 0, 40),
		}, append(eduOH, depOH...)...)

		examples = append(examples, training.Example{
			Input:    input,
			Response: []float64{score / 10.0},
		})
	}

	fmt.Printf("✅ %d empleados cargados\n", len(examples))
	return examples
}

func educationOneHot(edu string) []float64 {
	edu = strings.ToLower(strings.TrimSpace(edu))
	switch edu {
	case "high school": return []float64{1, 0, 0, 0}
	case "bachelor":    return []float64{0, 1, 0, 0}
	case "master":      return []float64{0, 0, 1, 0}
	case "phd":         return []float64{0, 0, 0, 1}
	default:            return []float64{0, 0, 0, 0}
	}
}

func departmentOneHot(dep string) []float64 {
	dep = strings.ToLower(strings.TrimSpace(dep))
	switch dep {
	case "finance": return []float64{1, 0, 0, 0}
	case "hr":      return []float64{0, 1, 0, 0}
	case "tech":    return []float64{0, 0, 1, 0}
	case "sales":   return []float64{0, 0, 0, 1}
	default:        return []float64{0, 0, 0, 0}
	}
}

func normalize(x, min, max float64) float64 {
	return (x - min) / (max - min)
}

func readCandidateInput() (float64, float64, string, string) {
	var age, exp float64
	var edu, dep string

	fmt.Print("Edad: ")
	fmt.Scanln(&age)
	fmt.Print("Años de experiencia: ")
	fmt.Scanln(&exp)
	fmt.Print("Educación (High School / Bachelor / Master / PhD): ")
	fmt.Scanln(&edu)
	fmt.Print("Departamento (Finance / HR / Tech / Sales): ")
	fmt.Scanln(&dep)

	return age, exp, edu, dep
}

func normalizeInput(age, exp float64, edu, dep string) []float64 {
	ageN := normalize(age, 20, 60)
	expN := normalize(exp, 0, 40)
	eduOH := educationOneHot(edu)
	depOH := departmentOneHot(dep)
	return append([]float64{ageN, expN}, append(eduOH, depOH...)...)
}
