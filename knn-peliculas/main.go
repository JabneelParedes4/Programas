package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Pelicula struct {
	Accion    float64
	Comedia   float64
	Terror    float64
	Romance   float64
	Nombre    string
	Distancia float64
}

var peliculas []Pelicula

var generos = []string{"Acción", "Comedia", "Terror", "Romance"}

func generoDominante(p Pelicula) string {
	maxVal := p.Accion
	idx := 0
	if p.Comedia > maxVal {
		maxVal = p.Comedia
		idx = 1
	}
	if p.Terror > maxVal {
		maxVal = p.Terror
		idx = 2
	}
	if p.Romance > maxVal {
		maxVal = p.Romance
		idx = 3
	}
	return generos[idx]
}

func distanciaEuclidiana(a, b Pelicula) float64 {
	return math.Sqrt(
		math.Pow(a.Accion-b.Accion, 2) +
			math.Pow(a.Comedia-b.Comedia, 2) +
			math.Pow(a.Terror-b.Terror, 2) +
			math.Pow(a.Romance-b.Romance, 2))
}

func predecirGenero(p Pelicula, train []Pelicula, k int) string {
	type vecino struct {
		dist float64
		gen  string
	}
	vecinos := make([]vecino, len(train))
	for i, t := range train {
		vecinos[i] = vecino{dist: distanciaEuclidiana(p, t), gen: generoDominante(t)}
	}
	sort.Slice(vecinos, func(i, j int) bool {
		return vecinos[i].dist < vecinos[j].dist
	})
	conteo := make(map[string]int)
	for i := 0; i < k && i < len(vecinos); i++ {
		conteo[vecinos[i].gen]++
	}
	maxGen := ""
	maxCnt := -1
	for gen, cnt := range conteo {
		if cnt > maxCnt {
			maxCnt = cnt
			maxGen = gen
		}
	}
	return maxGen
}

func evaluar(verdaderos, predichos []string) {
	clases := generos
	matriz := make(map[string]map[string]int)
	for _, c := range clases {
		matriz[c] = make(map[string]int)
		for _, c2 := range clases {
			matriz[c][c2] = 0
		}
	}
	for i := 0; i < len(verdaderos); i++ {
		matriz[verdaderos[i]][predichos[i]]++
	}

	fmt.Println("\n=== MATRIZ DE CONFUSIÓN ===")
	fmt.Printf("%-10s", "")
	for _, c := range clases {
		fmt.Printf("%-10s", c)
	}
	fmt.Println()
	for _, real := range clases {
		fmt.Printf("%-10s", real)
		for _, pred := range clases {
			fmt.Printf("%-10d", matriz[real][pred])
		}
		fmt.Println()
	}

	total := len(verdaderos)
	aciertos := 0
	for _, c := range clases {
		aciertos += matriz[c][c]
	}
	accuracy := float64(aciertos) / float64(total)

	fmt.Println("\n=== MÉTRICAS POR CLASE ===")
	precisionSum := 0.0
	recallSum := 0.0
	for _, c := range clases {
		tp := matriz[c][c]
		fp := 0
		fn := 0
		for _, otra := range clases {
			if otra != c {
				fp += matriz[otra][c]
				fn += matriz[c][otra]
			}
		}
		precision := float64(tp) / float64(tp+fp)
		if tp+fp == 0 {
			precision = 0
		}
		recall := float64(tp) / float64(tp+fn)
		if tp+fn == 0 {
			recall = 0
		}
		f1 := 2 * (precision * recall) / (precision + recall)
		if math.IsNaN(f1) {
			f1 = 0
		}
		fmt.Printf("%s - Precisión: %.3f, Recall: %.3f, F1: %.3f\n", c, precision, recall, f1)
		precisionSum += precision
		recallSum += recall
	}
	macroPrecision := precisionSum / float64(len(clases))
	macroRecall := recallSum / float64(len(clases))
	macroF1 := 2 * (macroPrecision * macroRecall) / (macroPrecision + macroRecall)
	if math.IsNaN(macroF1) {
		macroF1 = 0
	}
	fmt.Printf("\nAccuracy global: %.3f\n", accuracy)
	fmt.Printf("Macro Precisión: %.3f\n", macroPrecision)
	fmt.Printf("Macro Recall: %.3f\n", macroRecall)
	fmt.Printf("Macro F1: %.3f\n", macroF1)
}

func validacionCruzada(datos []Pelicula, kFold, kNeighbors int) {
	shuffled := make([]Pelicula, len(datos))
	copy(shuffled, datos)
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
		train := append(shuffled[:testStart], shuffled[testEnd:]...)

		for _, p := range test {
			real := generoDominante(p)
			pred := predecirGenero(p, train, kNeighbors)
			todosVerdaderos = append(todosVerdaderos, real)
			todosPredichos = append(todosPredichos, pred)
		}
	}
	evaluar(todosVerdaderos, todosPredichos)
}

func cargarCSV() ([]Pelicula, error) {
	file, err := os.Open("peliculas.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	datos, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var pelis []Pelicula
	for i, fila := range datos {
		if i == 0 {
			continue
		}
		accion, _ := strconv.ParseFloat(fila[0], 64)
		comedia, _ := strconv.ParseFloat(fila[1], 64)
		terror, _ := strconv.ParseFloat(fila[2], 64)
		romance, _ := strconv.ParseFloat(fila[3], 64)
		pelis = append(pelis, Pelicula{
			Accion:  accion,
			Comedia: comedia,
			Terror:  terror,
			Romance: romance,
			Nombre:  fila[4],
		})
	}
	return pelis, nil
}

func guardarModelo(pelis []Pelicula, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	return encoder.Encode(pelis)
}

func cargarModelo(filename string) ([]Pelicula, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var pelis []Pelicula
	err = decoder.Decode(&pelis)
	return pelis, err
}

// Leer un número entre 1 y 5 desde teclado
func leerValor(pregunta string) float64 {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(pregunta + " (1-5): ")
		texto, _ := reader.ReadString('\n')
		texto = strings.TrimSpace(texto)
		val, err := strconv.ParseFloat(texto, 64)
		if err == nil && val >= 1 && val <= 5 {
			return val
		}
		fmt.Println("Por favor ingresa un número entre 1 y 5.")
	}
}

func recomendarPeliculas(usuario Pelicula, pelis []Pelicula, topN int) {
	// Copiar y calcular distancias
	copia := make([]Pelicula, len(pelis))
	copy(copia, pelis)
	for i := range copia {
		copia[i].Distancia = math.Sqrt(
			math.Pow(usuario.Accion-copia[i].Accion, 2) +
				math.Pow(usuario.Comedia-copia[i].Comedia, 2) +
				math.Pow(usuario.Terror-copia[i].Terror, 2) +
				math.Pow(usuario.Romance-copia[i].Romance, 2))
	}
	sort.Slice(copia, func(i, j int) bool {
		return copia[i].Distancia < copia[j].Distancia
	})
	fmt.Println("\n🎬 Tus recomendaciones personalizadas:")
	for i := 0; i < topN && i < len(copia); i++ {
		fmt.Printf("%d. 🍿 %s (distancia: %.2f)\n", i+1, copia[i].Nombre, copia[i].Distancia)
	}
}

func main() {
	// Cargar o entrenar modelo
	var err error
	peliculas, err = cargarModelo("modelo.knn")
	if err != nil {
		fmt.Println("No se encontró modelo guardado. Cargando CSV y evaluando...")
		peliculas, err = cargarCSV()
		if err != nil {
			panic("Error al leer peliculas.csv: " + err.Error())
		}
		validacionCruzada(peliculas, 5, 3)
		if err := guardarModelo(peliculas, "modelo.knn"); err != nil {
			fmt.Println("Error guardando modelo:", err)
		} else {
			fmt.Println("Modelo guardado en modelo.knn")
		}
	} else {
		fmt.Println("Modelo cargado exitosamente desde modelo.knn")
	}

	fmt.Println("\n===== SISTEMA DE RECOMENDACIÓN DE PELÍCULAS (KNN) =====")
	fmt.Println("Responde las siguientes preguntas sobre tus preferencias (1 = nada, 5 = mucho)")

	// Bucle interactivo
	for {
		usuario := Pelicula{
			Accion:  leerValor("¿Cuánto te gusta la ACCIÓN?"),
			Comedia: leerValor("¿Cuánto te gusta la COMEDIA?"),
			Terror:  leerValor("¿Cuánto te gusta el TERROR?"),
			Romance: leerValor("¿Cuánto te gusta el ROMANCE?"),
		}
		recomendarPeliculas(usuario, peliculas, 3)

		fmt.Print("\n¿Quieres hacer otra consulta? (s/n): ")
		reader := bufio.NewReader(os.Stdin)
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))
		if resp != "s" && resp != "si" {
			break
		}
	}
	fmt.Println("¡Gracias por usar CineMatch CLI! 🎉")
}