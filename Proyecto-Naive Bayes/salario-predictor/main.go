package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Record struct {
	Edad            float64
	Promedio        float64
	Ingles          string
	Practicas       string
	Programacion    string
	Experiencia     string
	Certificaciones string
	Universidad     string
	Salario         string
}

type Stats struct {
	Mean float64
	Std  float64
}

type Model struct {
	ClassCounts  map[string]int
	Total        int
	NumericStats map[string]map[string]Stats
	Categorical  map[string]map[string]map[string]int
}

var clases = []string{"Bajo", "Medio", "Alto"}

var salarioRango = map[string]string{
	"Bajo":  "Bajo (8,000 - 12,000 MXN)",
	"Medio": "Medio (13,000 - 18,000 MXN)",
	"Alto":  "Alto (19,000 - 25,000+ MXN)",
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("===================================")
	fmt.Println("   PREDICTOR DE SALARIO - NAIVE BAYES")
	fmt.Println("===================================")

	model, err := loadModel("models/modelo.json")
	if err != nil {
		fmt.Println("No se encontró modelo guardado. Entrenando nuevo modelo...")

		records := loadCSV("data/salarios.csv")
		trainData, testData := splitData(records, 0.8)

		model = trainNaiveBayes(trainData)

		fmt.Println("\nRealizando Validación Cruzada (5-Fold)...")
		validacionCruzada(records, 5)

		if err := saveModel(model, "models/modelo.json"); err != nil {
			fmt.Println("Error guardando modelo:", err)
		} else {
			fmt.Println("Modelo entrenado y guardado correctamente.")
		}

		fmt.Println("\n=== Evaluación Final (Hold-out) ===")
		evaluate(model, testData)
	} else {
		fmt.Println("Modelo cargado exitosamente.")
		fmt.Println("\n=== Evaluación del Modelo Cargado ===")
		// Evaluamos con todo el dataset para mostrar métricas
		records := loadCSV("data/salarios.csv")
		evaluate(model, records)
	}

	interfazInteractiva(model)
}

// ==================== CARGA DE DATOS ====================

func loadCSV(path string) []Record {
	file, err := os.Open(path)
	if err != nil {
		panic("Error abriendo CSV: " + err.Error())
	}
	defer file.Close()

	reader := csv.NewReader(file)
	data, _ := reader.ReadAll()

	var records []Record
	for i, row := range data {
		if i == 0 || len(row) < 9 {
			continue
		}
		edad, _ := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
		promedio, _ := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)

		records = append(records, Record{
			Edad:            edad,
			Promedio:        promedio,
			Ingles:          normalizeInput(row[2]),
			Practicas:       normalizeInput(row[3]),
			Programacion:    normalizeInput(row[4]),
			Experiencia:     normalizeInput(row[5]),
			Certificaciones: normalizeInput(row[6]),
			Universidad:     normalizeInput(row[7]),
			Salario:         strings.TrimSpace(row[8]),
		})
	}
	return records
}

func splitData(data []Record, ratio float64) ([]Record, []Record) {
	rand.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
	trainSize := int(float64(len(data)) * ratio)
	return data[:trainSize], data[trainSize:]
}

// ==================== VALIDACIÓN CRUZADA ====================

func validacionCruzada(data []Record, kFold int) {
	shuffled := make([]Record, len(data))
	copy(shuffled, data)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	foldSize := len(shuffled) / kFold
	todosVerdaderos := []string{}
	todosPredichos := []string{}

	for fold := 0; fold < kFold; fold++ {
		testStart := fold * foldSize
		testEnd := testStart + foldSize
		if fold == kFold-1 {
			testEnd = len(shuffled)
		}

		test := shuffled[testStart:testEnd]
		train := append([]Record{}, shuffled[:testStart]...)
		train = append(train, shuffled[testEnd:]...)

		foldModel := trainNaiveBayes(train)

		for _, r := range test {
			real := r.Salario
			pred := predict(foldModel, r)
			todosVerdaderos = append(todosVerdaderos, real)
			todosPredichos = append(todosPredichos, pred)
		}
	}

	evaluarCompleto(todosVerdaderos, todosPredichos)
}

// ==================== FUNCIONES AUXILIARES ====================

func normalizeInput(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ==================== ESTADÍSTICAS ====================

func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func std(values []float64) float64 {
	m := mean(values)
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-m, 2)
	}
	variance /= float64(len(values))
	if variance < 1e-6 {
		return 1.0
	}
	return math.Sqrt(variance)
}

// ==================== ENTRENAMIENTO ====================

func trainNaiveBayes(data []Record) Model {
	model := Model{
		ClassCounts:  make(map[string]int),
		NumericStats: make(map[string]map[string]Stats),
		Categorical:  make(map[string]map[string]map[string]int),
		Total:        0,
	}

	for _, f := range []string{"Edad", "Promedio"} {
		model.NumericStats[f] = make(map[string]Stats)
	}
	for _, f := range []string{"Ingles", "Practicas", "Programacion", "Experiencia", "Certificaciones", "Universidad"} {
		model.Categorical[f] = make(map[string]map[string]int)
	}

	classEdad := make(map[string][]float64)
	classPromedio := make(map[string][]float64)

	for _, r := range data {
		class := r.Salario
		model.ClassCounts[class]++
		model.Total++

		classEdad[class] = append(classEdad[class], r.Edad)
		classPromedio[class] = append(classPromedio[class], r.Promedio)

		incrementCategorical(model.Categorical["Ingles"], class, r.Ingles)
		incrementCategorical(model.Categorical["Practicas"], class, r.Practicas)
		incrementCategorical(model.Categorical["Programacion"], class, r.Programacion)
		incrementCategorical(model.Categorical["Experiencia"], class, r.Experiencia)
		incrementCategorical(model.Categorical["Certificaciones"], class, r.Certificaciones)
		incrementCategorical(model.Categorical["Universidad"], class, r.Universidad)
	}

	for class, values := range classEdad {
		model.NumericStats["Edad"][class] = Stats{Mean: mean(values), Std: std(values)}
	}
	for class, values := range classPromedio {
		model.NumericStats["Promedio"][class] = Stats{Mean: mean(values), Std: std(values)}
	}

	return model
}

func incrementCategorical(feature map[string]map[string]int, class, value string) {
	if feature[class] == nil {
		feature[class] = make(map[string]int)
	}
	feature[class][value]++
}

// ==================== PROBABILIDADES Y PREDICCIÓN ====================

func gaussian(x, mean, std float64) float64 {
	if std == 0 {
		std = 1
	}
	exponent := math.Exp(-math.Pow(x-mean, 2) / (2 * math.Pow(std, 2)))
	return (1 / (math.Sqrt(2*math.Pi) * std)) * exponent
}

func categoricalProbability(feature map[string]map[string]int, class, value string) float64 {
	count := feature[class][value]
	total := 0
	for _, c := range feature[class] {
		total += c
	}
	if total == 0 {
		return 0.01
	}
	return float64(count+1) / float64(total+3)
}

func predict(model Model, r Record) string {
	bestClass := ""
	bestProb := math.Inf(-1)

	for class := range model.ClassCounts {
		prior := float64(model.ClassCounts[class]) / float64(model.Total)
		prob := math.Log(prior + 1e-10)

		prob += math.Log(gaussian(r.Edad, model.NumericStats["Edad"][class].Mean, model.NumericStats["Edad"][class].Std) + 1e-10)
		prob += math.Log(gaussian(r.Promedio, model.NumericStats["Promedio"][class].Mean, model.NumericStats["Promedio"][class].Std) + 1e-10)

		prob += math.Log(categoricalProbability(model.Categorical["Ingles"], class, r.Ingles) + 1e-10)
		prob += math.Log(categoricalProbability(model.Categorical["Practicas"], class, r.Practicas) + 1e-10)
		prob += math.Log(categoricalProbability(model.Categorical["Programacion"], class, r.Programacion) + 1e-10)
		prob += math.Log(categoricalProbability(model.Categorical["Experiencia"], class, r.Experiencia) + 1e-10)
		prob += math.Log(categoricalProbability(model.Categorical["Certificaciones"], class, r.Certificaciones) + 1e-10)
		prob += math.Log(categoricalProbability(model.Categorical["Universidad"], class, r.Universidad) + 1e-10)

		if prob > bestProb {
			bestProb = prob
			bestClass = class
		}
	}
	return bestClass
}

// ==================== EVALUACIÓN ====================

func evaluarCompleto(verdaderos, predichos []string) {
	fmt.Println("\n=== RESULTADOS VALIDACIÓN CRUZADA (5-Fold) ===")
	evaluarModelo(verdaderos, predichos)
}

func evaluate(model Model, testData []Record) {
	verdaderos := []string{}
	predichos := []string{}

	for _, r := range testData {
		pred := predict(model, r)
		verdaderos = append(verdaderos, r.Salario)
		predichos = append(predichos, pred)
	}
	evaluarModelo(verdaderos, predichos)
}

func evaluarModelo(verdaderos, predichos []string) {
	matrix := make(map[string]map[string]int)
	for _, c1 := range clases {
		matrix[c1] = make(map[string]int)
		for _, c2 := range clases {
			matrix[c1][c2] = 0
		}
	}

	correct := 0
	for i := range verdaderos {
		real := verdaderos[i]
		pred := predichos[i]
		matrix[real][pred]++
		if real == pred {
			correct++
		}
	}

	accuracy := float64(correct) / float64(len(verdaderos))

	// Matriz de Confusión
	fmt.Println("\nMatriz de Confusión:")
	fmt.Println("       Bajo   Medio   Alto")
	for _, real := range clases {
		fmt.Printf("%-6s", real)
		for _, pred := range clases {
			fmt.Printf("%6d ", matrix[real][pred])
		}
		fmt.Println()
	}

	fmt.Printf("\nAccuracy Global: %.2f%%\n", accuracy*100)

	// Métricas por clase
	fmt.Println("\nMétricas por Clase:")
	for _, c := range clases {
		tp := matrix[c][c]
		fp := 0
		fn := 0
		for _, otra := range clases {
			if otra != c {
				fp += matrix[otra][c]
				fn += matrix[c][otra]
			}
		}

		precision := 0.0
		if tp+fp > 0 {
			precision = float64(tp) / float64(tp+fp)
		}
		recall := 0.0
		if tp+fn > 0 {
			recall = float64(tp) / float64(tp+fn)
		}
		f1 := 0.0
		if precision+recall > 0 {
			f1 = 2 * (precision * recall) / (precision + recall)
		}

		fmt.Printf("%-6s - Precisión: %.3f | Recall: %.3f | F1: %.3f\n", c, precision, recall, f1)
	}
}

// ==================== INTERFAZ INTERACTIVA ====================

func interfazInteractiva(model Model) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n=== Predice tu salario en el primer trabajo ===")

	for {
		fmt.Println("\nIngresa tus datos:")

		edad := leerFloat("Edad", 18, 40)
		promedio := leerFloat("Promedio", 60, 100)
		ingles := leerOpcion("Nivel de Inglés", []string{"Bajo", "Medio", "Alto"})
		practicas := leerOpcion("¿Hiciste prácticas?", []string{"Si", "No"})
		programacion := leerOpcion("Nivel de Programación", []string{"Bajo", "Medio", "Alto"})
		experiencia := leerOpcion("Experiencia laboral", []string{"Ninguna", "Poca", "Media"})
		certificaciones := leerOpcion("¿Tienes certificaciones?", []string{"Si", "No"})
		universidad := leerOpcion("Universidad", []string{"Pública", "Privada", "Prestigiosa"})

		usuario := Record{
			Edad:            edad,
			Promedio:        promedio,
			Ingles:          normalizeInput(ingles),
			Practicas:       normalizeInput(practicas),
			Programacion:    normalizeInput(programacion),
			Experiencia:     normalizeInput(experiencia),
			Certificaciones: normalizeInput(certificaciones),
			Universidad:     normalizeInput(universidad),
		}

		prediccion := predict(model, usuario)
		fmt.Printf("\n🎯 Predicción: Tu salario esperado es **%s**\n", salarioRango[prediccion])

		fmt.Print("\n¿Quieres hacer otra predicción? (s/n): ")
		resp, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(resp)) != "s" {
			break
		}
	}
}

func leerFloat(pregunta string, min, max float64) float64 {
	for {
		fmt.Printf("%s (%g-%g): ", pregunta, min, max)
		var val float64
		_, err := fmt.Scanf("%f", &val)
		if err == nil && val >= min && val <= max {
			return val
		}
		fmt.Println("Valor inválido. Intenta de nuevo.")
	}
}

func leerOpcion(pregunta string, opciones []string) string {
	for {
		fmt.Printf("%s (%s): ", pregunta, strings.Join(opciones, "/"))
		var resp string
		fmt.Scanln(&resp)
		respNorm := normalizeInput(resp)

		for _, opt := range opciones {
			if normalizeInput(opt) == respNorm {
				return opt
			}
		}
		fmt.Println("Opción inválida. Inténtalo de nuevo.")
	}
}

// ==================== PERSISTENCIA ====================

func saveModel(model Model, path string) error {
	os.MkdirAll("models", os.ModePerm)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewEncoder(file).Encode(model)
}

func loadModel(path string) (Model, error) {
	file, err := os.Open(path)
	if err != nil {
		return Model{}, err
	}
	defer file.Close()

	var model Model
	err = json.NewDecoder(file).Decode(&model)
	return model, err
}
