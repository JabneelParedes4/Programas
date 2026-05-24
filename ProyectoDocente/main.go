package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Docente struct {
	Edad                  float64
	Experiencia           float64
	HabilidadPedagogica   float64
	ParticipacionAlumnos  float64
	ResultadosAprendizaje float64
	Innovacion            float64
	Profesionalismo       float64
	SeguridadConfianza    float64
	FactorContextual      float64
	PuntajeEntrevista     float64
	Referencias           float64
	EsBueno               bool
}

type Estadisticas struct {
	Media float64
	Std   float64
}

type ModeloBayes struct {
	PriorBueno float64
	PriorMalo  float64

	StatsBuenos map[string]Estadisticas
	StatsMalos  map[string]Estadisticas
}

func main() {

	rand.Seed(time.Now().UnixNano())

	fmt.Println("===================================================")
	fmt.Println(" PREDICTOR DE BUEN DOCENTE - NAIVE BAYES")
	fmt.Println("===================================================")

	modelo, err := cargarModelo("modelo.gob")

	if err != nil {

		fmt.Println("\nNo existe modelo entrenado.")
		fmt.Println("Entrenando modelo bayesiano...")

		docentes, err := cargarCSV("data/docentes.csv")

		if err != nil {
			fmt.Println("Error cargando CSV:", err)
			return
		}

		fmt.Println("\n=========== ANÁLISIS DEL DATASET ===========")
		fmt.Printf("Total registros: %d\n", len(docentes))
		fmt.Println("Variables utilizadas: 11")
		fmt.Println("Modelo utilizado: Gaussian Naive Bayes")

		// CROSS VALIDATION
		crossValidation(docentes, 5)

		// ENTRENAMIENTO FINAL
		train, _ := dividirDatos(docentes)

		modelo = entrenarModelo(train)

		// GUARDAR MODELO
		err = guardarModelo(modelo, "modelo.gob")

		if err != nil {
			fmt.Println("Error guardando modelo:", err)
			return
		}

		fmt.Println("\nModelo entrenado y guardado correctamente.")
	} else {

		fmt.Println("\nModelo cargado correctamente.")
	}

	reader := bufio.NewReader(os.Stdin)

	for {

		fmt.Println("\n===================================")
		fmt.Println(" NUEVO DOCENTE")
		fmt.Println("===================================")

		docente := Docente{}

		docente.Edad = leerNumero(reader, "Edad")
		docente.Experiencia = leerNumero(reader, "Años de experiencia")
		docente.HabilidadPedagogica = leerNumero(reader, "Habilidad Pedagógica")
		docente.ParticipacionAlumnos = leerNumero(reader, "Participación de alumnos")
		docente.ResultadosAprendizaje = leerNumero(reader, "Resultados de aprendizaje")
		docente.Innovacion = leerNumero(reader, "Innovación")
		docente.Profesionalismo = leerNumero(reader, "Profesionalismo")
		docente.SeguridadConfianza = leerNumero(reader, "Seguridad y confianza")
		docente.FactorContextual = leerNumero(reader, "Factor contextual")
		docente.PuntajeEntrevista = leerNumero(reader, "Puntaje entrevista")
		docente.Referencias = leerNumero(reader, "Referencias")

		resultado, prob := predecir(docente, modelo)

		fmt.Println("\n=========== RESULTADO ===========")

		if resultado {

			fmt.Printf("Probabilidad de ser BUEN docente: %.2f%%\n", prob*100)
			fmt.Println("Resultado: RECOMENDADO")

		} else {

			fmt.Printf("Probabilidad de ser MAL docente: %.2f%%\n", prob*100)
			fmt.Println("Resultado: NO RECOMENDADO")
		}

		fmt.Print("\n¿Evaluar otro docente? (s/n): ")

		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))

		if resp != "s" {
			break
		}
	}

	fmt.Println("\nGracias por usar el sistema.")
}

func cargarCSV(nombre string) ([]Docente, error) {

	file, err := os.Open(nombre)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := csv.NewReader(file)

	data, err := reader.ReadAll()

	if err != nil {
		return nil, err
	}

	var docentes []Docente

	for i, row := range data {

		if i == 0 {
			continue
		}

		bueno := row[11] == "1"

		docente := Docente{
			Edad:                  parseFloat(row[0]),
			Experiencia:           parseFloat(row[1]),
			HabilidadPedagogica:   parseFloat(row[2]),
			ParticipacionAlumnos:  parseFloat(row[3]),
			ResultadosAprendizaje: parseFloat(row[4]),
			Innovacion:            parseFloat(row[5]),
			Profesionalismo:       parseFloat(row[6]),
			SeguridadConfianza:    parseFloat(row[7]),
			FactorContextual:      parseFloat(row[8]),
			PuntajeEntrevista:     parseFloat(row[9]),
			Referencias:           parseFloat(row[10]),
			EsBueno:               bueno,
		}

		docentes = append(docentes, docente)
	}

	return docentes, nil
}

func dividirDatos(docentes []Docente) ([]Docente, []Docente) {

	sizeTrain := int(float64(len(docentes)) * 0.8)

	train := docentes[:sizeTrain]
	test := docentes[sizeTrain:]

	return train, test
}

func parseFloat(s string) float64 {

	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)

	return v
}

func entrenarModelo(docentes []Docente) ModeloBayes {

	var buenos []Docente
	var malos []Docente

	for _, d := range docentes {

		if d.EsBueno {
			buenos = append(buenos, d)
		} else {
			malos = append(malos, d)
		}
	}

	modelo := ModeloBayes{}

	modelo.PriorBueno = float64(len(buenos)) / float64(len(docentes))
	modelo.PriorMalo = float64(len(malos)) / float64(len(docentes))

	modelo.StatsBuenos = calcularEstadisticas(buenos)
	modelo.StatsMalos = calcularEstadisticas(malos)

	return modelo
}

func calcularEstadisticas(docentes []Docente) map[string]Estadisticas {

	stats := make(map[string]Estadisticas)

	campos := map[string][]float64{
		"Edad":                  {},
		"Experiencia":           {},
		"HabilidadPedagogica":   {},
		"ParticipacionAlumnos":  {},
		"ResultadosAprendizaje": {},
		"Innovacion":            {},
		"Profesionalismo":       {},
		"SeguridadConfianza":    {},
		"FactorContextual":      {},
		"PuntajeEntrevista":     {},
		"Referencias":           {},
	}

	for _, d := range docentes {

		campos["Edad"] = append(campos["Edad"], d.Edad)
		campos["Experiencia"] = append(campos["Experiencia"], d.Experiencia)
		campos["HabilidadPedagogica"] = append(campos["HabilidadPedagogica"], d.HabilidadPedagogica)
		campos["ParticipacionAlumnos"] = append(campos["ParticipacionAlumnos"], d.ParticipacionAlumnos)
		campos["ResultadosAprendizaje"] = append(campos["ResultadosAprendizaje"], d.ResultadosAprendizaje)
		campos["Innovacion"] = append(campos["Innovacion"], d.Innovacion)
		campos["Profesionalismo"] = append(campos["Profesionalismo"], d.Profesionalismo)
		campos["SeguridadConfianza"] = append(campos["SeguridadConfianza"], d.SeguridadConfianza)
		campos["FactorContextual"] = append(campos["FactorContextual"], d.FactorContextual)
		campos["PuntajeEntrevista"] = append(campos["PuntajeEntrevista"], d.PuntajeEntrevista)
		campos["Referencias"] = append(campos["Referencias"], d.Referencias)
	}

	for nombre, valores := range campos {

		media := promedio(valores)
		std := desviacion(valores, media)

		stats[nombre] = Estadisticas{
			Media: media,
			Std:   std,
		}
	}

	return stats
}

func promedio(nums []float64) float64 {

	sum := 0.0

	for _, n := range nums {
		sum += n
	}

	return sum / float64(len(nums))
}

func desviacion(nums []float64, media float64) float64 {

	sum := 0.0

	for _, n := range nums {
		sum += math.Pow(n-media, 2)
	}

	return math.Sqrt(sum / float64(len(nums)))
}

func gaussian(x, media, std float64) float64 {

	if std == 0 {
		std = 0.0001
	}

	exponente := math.Exp(-math.Pow(x-media, 2) / (2 * math.Pow(std, 2)))

	return (1 / (math.Sqrt(2*math.Pi) * std)) * exponente
}

func predecir(d Docente, modelo ModeloBayes) (bool, float64) {

	// VALIDACIONES LÓGICAS

	if d.Experiencia < 4 {
		return false, 0.95
	}

	if d.HabilidadPedagogica < 70 {
		return false, 0.90
	}

	if d.ResultadosAprendizaje < 70 {
		return false, 0.90
	}

	probBueno := math.Log(modelo.PriorBueno)
	probMalo := math.Log(modelo.PriorMalo)

	valores := map[string]float64{
		"Edad":                  d.Edad,
		"Experiencia":           d.Experiencia,
		"HabilidadPedagogica":   d.HabilidadPedagogica,
		"ParticipacionAlumnos":  d.ParticipacionAlumnos,
		"ResultadosAprendizaje": d.ResultadosAprendizaje,
		"Innovacion":            d.Innovacion,
		"Profesionalismo":       d.Profesionalismo,
		"SeguridadConfianza":    d.SeguridadConfianza,
		"FactorContextual":      d.FactorContextual,
		"PuntajeEntrevista":     d.PuntajeEntrevista,
		"Referencias":           d.Referencias,
	}

	for campo, valor := range valores {

		sb := modelo.StatsBuenos[campo]
		sm := modelo.StatsMalos[campo]

		probBueno += math.Log(gaussian(valor, sb.Media, sb.Std))
		probMalo += math.Log(gaussian(valor, sm.Media, sm.Std))
	}

	if probBueno > probMalo {

		prob := math.Exp(probBueno) / (math.Exp(probBueno) + math.Exp(probMalo))

		return true, prob
	}

	prob := math.Exp(probMalo) / (math.Exp(probBueno) + math.Exp(probMalo))

	return false, prob
}

func crossValidation(docentes []Docente, folds int) {

	rand.Shuffle(len(docentes), func(i, j int) {
		docentes[i], docentes[j] = docentes[j], docentes[i]
	})

	foldSize := len(docentes) / folds

	totalTP := 0
	totalTN := 0
	totalFP := 0
	totalFN := 0

	totalAccuracy := 0.0

	for i := 0; i < folds; i++ {

		inicio := i * foldSize
		fin := inicio + foldSize

		if i == folds-1 {
			fin = len(docentes)
		}

		test := docentes[inicio:fin]

		train := append(docentes[:inicio], docentes[fin:]...)

		modelo := entrenarModelo(train)

		tp := 0
		tn := 0
		fp := 0
		fn := 0

		for _, d := range test {

			pred, _ := predecir(d, modelo)

			if pred && d.EsBueno {
				tp++
			} else if !pred && !d.EsBueno {
				tn++
			} else if pred && !d.EsBueno {
				fp++
			} else if !pred && d.EsBueno {
				fn++
			}
		}

		accuracy := float64(tp+tn) / float64(tp+tn+fp+fn)

		totalAccuracy += accuracy

		totalTP += tp
		totalTN += tn
		totalFP += fp
		totalFN += fn

		fmt.Printf("\nFold %d Accuracy: %.2f%%\n", i+1, accuracy*100)
	}

	accuracyFinal := totalAccuracy / float64(folds)

	precision := float64(totalTP) / float64(totalTP+totalFP)

	recall := float64(totalTP) / float64(totalTP+totalFN)

	f1 := 2 * (precision * recall) / (precision + recall)

	fmt.Println("\n====================================")
	fmt.Println(" VALIDACIÓN CRUZADA - CROSS VALIDATION")
	fmt.Println("====================================")

	fmt.Println("\nMATRIZ DE CONFUSIÓN GLOBAL")

	fmt.Printf("TP: %d\n", totalTP)
	fmt.Printf("TN: %d\n", totalTN)
	fmt.Printf("FP: %d\n", totalFP)
	fmt.Printf("FN: %d\n", totalFN)

	fmt.Println("\nMÉTRICAS FINALES")

	fmt.Printf("Accuracy Promedio: %.2f%%\n", accuracyFinal*100)
	fmt.Printf("Precision: %.2f%%\n", precision*100)
	fmt.Printf("Recall: %.2f%%\n", recall*100)
	fmt.Printf("F1 Score: %.2f%%\n", f1*100)
}

func guardarModelo(modelo ModeloBayes, archivo string) error {

	file, err := os.Create(archivo)

	if err != nil {
		return err
	}

	defer file.Close()

	encoder := gob.NewEncoder(file)

	return encoder.Encode(modelo)
}

func cargarModelo(archivo string) (ModeloBayes, error) {

	file, err := os.Open(archivo)

	if err != nil {
		return ModeloBayes{}, err
	}

	defer file.Close()

	decoder := gob.NewDecoder(file)

	var modelo ModeloBayes

	err = decoder.Decode(&modelo)

	return modelo, err
}

func leerNumero(reader *bufio.Reader, texto string) float64 {

	for {

		fmt.Printf("%s: ", texto)

		input, _ := reader.ReadString('\n')

		input = strings.TrimSpace(input)

		valor, err := strconv.ParseFloat(input, 64)

		if err == nil {
			return valor
		}

		fmt.Println("Número inválido.")
	}
}
