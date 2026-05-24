package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/patrikeh/go-deep"
	"github.com/patrikeh/go-deep/training"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Cargando datos...")
	examples := loadAndProcessData("combined_employee_dataset_1.csv")

	fmt.Println("Entrenando modelo (esto toma ~15-25 segundos)...")

	nn := deep.NewNeural(&deep.Config{
		Inputs:     10,
		Layout:     []int{96, 48, 24, 1},
		Activation: deep.ActivationReLU,
		Mode:       deep.ModeRegression,
		Weight:     deep.NewNormal(0.08, 0.0),
		Bias:       true,
	})

	optimizer := training.NewSGD(0.018, 0.9, 1e-6, false)
	trainer := training.NewTrainer(optimizer, 80)
	train, val := examples.Split(0.85)
	trainer.Train(nn, train, val, 500)   // Reducido para que sea más rápido

	fmt.Println("\n¡Modelo listo!")

	fmt.Println("\n=== Predictor Inteligente de Empleados ===")

	for {
		fmt.Println("\n--- Nuevo Candidato ---")
		age, exp, edu, dep := readCandidateInput()

		input := normalizeInput(age, exp, edu, dep)
		nnPred := nn.Predict(input)[0] * 10

		finalScore := applyLogicAdjustment(nnPred, edu, exp, age, dep)
		finalScore = math.Max(1.0, math.Min(10.0, finalScore))

		fmt.Printf("\n🎯 Predicción final (3 meses): %.1f / 10\n", finalScore)

		if finalScore >= 8.0 {
			fmt.Println("✅ Excelente candidato - Muy recomendado")
		} else if finalScore >= 6.5 {
			fmt.Println("👍 Buen candidato")
		} else if finalScore >= 5.0 {
			fmt.Println("⚠️ Candidato promedio")
		} else {
			fmt.Println("❌ Riesgo alto - Evaluar con cuidado")
		}

		fmt.Print("\n¿Probar otro candidato? (s/n): ")
		var again string
		fmt.Scanln(&again)
		if again != "s" && again != "S" {
			fmt.Println("¡Gracias por usar el predictor!")
			break
		}
	}
}

func applyLogicAdjustment(score float64, edu string, exp, age float64, dep string) float64 {
	edu = strings.ToLower(strings.TrimSpace(edu))
	dep = strings.ToLower(strings.TrimSpace(dep))
	bonus := 0.0

	switch edu {
	case "phd":
		bonus += 1.8
	case "master":
		bonus += 1.1
	case "bachelor":
		bonus += 0.4
	}

	if exp >= 15 {
		bonus += 1.6
	} else if exp >= 7 {
		bonus += 0.9
	} else if exp >= 3 {
		bonus += 0.4
	}

	if age >= 32 && age <= 50 {
		bonus += 0.5
	}

	if dep == "tech" || dep == "finance" {
		bonus += 0.3
	}

	return score + bonus
}
