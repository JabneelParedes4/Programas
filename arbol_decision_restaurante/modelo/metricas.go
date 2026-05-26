package modelo

import (
	"fmt"
)

func MatrizConfusion(verdaderos, predicciones []int, nClases int) {
	matriz := make([][]int, nClases)
	for i := range matriz {
		matriz[i] = make([]int, nClases)
	}
	for i := 0; i < len(verdaderos); i++ {
		matriz[verdaderos[i]][predicciones[i]]++
	}

	fmt.Println("\n📊 Matriz de Confusión:")
	fmt.Print("      ")
	for i := 0; i < nClases; i++ {
		fmt.Printf("Rec%d  ", i)
	}
	fmt.Println()
	for i := 0; i < nClases; i++ {
		fmt.Printf("Real%d ", i)
		for j := 0; j < nClases; j++ {
			fmt.Printf(" %4d  ", matriz[i][j])
		}
		fmt.Println()
	}
}

func Exactitud(verdaderos, predicciones []int) float64 {
	if len(verdaderos) == 0 {
		return 0
	}
	aciertos := 0
	for i := 0; i < len(verdaderos); i++ {
		if verdaderos[i] == predicciones[i] {
			aciertos++
		}
	}
	return float64(aciertos) / float64(len(verdaderos))
}

func MostrarMetricas(verdaderos, predicciones []int) {
	acc := Exactitud(verdaderos, predicciones)
	fmt.Printf("\n=== 📈 MÉTRICAS DEL RECOMENDADOR ===\n")
	fmt.Printf("🎯 Exactitud: %.2f%%\n", acc*100)
	fmt.Println("=====================================\n")
}
