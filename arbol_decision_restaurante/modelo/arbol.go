package modelo

import (
	"fmt"
	"math"
)

type Nodo struct {
	EsHoja     bool
	Prediccion int
	Indice     int     // 0:presupuesto, 1:tipo_comida, 2:ocasion, 3:vegano, 4:calorias
	Valor      string  // Para variables categóricas
	Umbral     float64 // Para variables numéricas
	Izquierdo  *Nodo
	Derecho    *Nodo
}

type ArbolDecision struct {
	Raiz *Nodo
}

// Mapas de conversión
var PresupuestoMap = map[string]int{"bajo": 0, "medio": 1, "alto": 2}
var TipoComidaMap = map[string]int{"italiana": 0, "mexicana": 1, "asiatica": 2}
var OcasionMap = map[string]int{"familiar": 0, "cita": 1, "negocio": 2}
var VeganoMap = map[string]int{"no": 0, "si": 1}

// Mapa de predicciones (combinaciones de platillos)
var PrediccionMap = map[int]string{
	// ITALIANA
	0: "🍕 Pizza Margarita + 🥤 Refresco + 🍰 Tiramisú",
	1: "🍕 Pizza Pepperoni + 🥤 Refresco + 🍫 Tiramisú",
	2: "🍝 Espagueti Bolognesa + 🥤 Refresco + 🍮 Panna Cotta",
	3: "🍝 Pasta Carbonara + 🍷 Vino tinto + 🍮 Panna Cotta",
	4: "🍲 Lasagna Pequeña + 🥤 Refresco + 🍰 Tiramisú",
	5: "🥟 Ravioles Queso + 🥤 Refresco + 🍮 Flan",
	6: "🍲 Lasagna + 🥗 Ensalada + 🍫 Tiramisú",
	7: "🍚 Risotto de Setas + 🥂 Espumante + 🍓 Frutillas",
	8: "🥟 Ravioles de Espinaca + 🥤 Jugo + 🍎 Tarta",

	// MEXICANA
	9:  "🌮 Tacos al Pastor + 🥤 Agua Jamaica + 🍩 Churros",
	10: "🌮 Enchiladas Verdes + 🥤 Refresco + 🍩 Buñuelos",
	11: "🧀 Quesadillas + 🥑 Guacamole + 🍮 Flan",
	12: "🍛 Mole Poblano + 🥤 Horchata + 🍩 Buñuelos",
	13: "🌯 Enchiladas Suizas + 🥑 Guacamole + 🍮 Arroz con leche",
	14: "🌮 Tostadas Veganas + 🥤 Agua Jamaica + 🍉 Frutas",
	15: "🐟 Ceviche Vegano + 🥤 Limonada + 🍍 Frutas",
	16: "🥩 Carne Asada + 🥤 Horchata + 🍩 Churros",

	// ASIATICA
	17: "🍚 Arroz Frito + 🥤 Té + 🥟 Rollitos primavera",
	18: "🍜 Chow Mein + 🥤 Té + 🥟 Wantanes",
	19: "🥟 Gyozas + 🥤 Té + 🍚 Arroz",
	20: "🥟 Rollos Primavera + 🥤 Té + 🍚 Arroz Frito",
	21: "🍲 Tailandés Verde + 🥤 Agua + 🍚 Arroz jazmín",
	22: "🍜 Ramen Vegetariano + 🥤 Té + 🥟 Gyozas",
	23: "🍣 Sushi Mixto + 🍶 Sake + 🍡 Mochi",
	24: "🍜 Pad Thai + 🥤 Limonada + 🥟 Rollitos",
	25: "🥘 Pad Thai Camarón + 🥤 Limonada + 🥟 Rollitos",
	26: "🍛 Curry Amarillo + 🥤 Té + 🍚 Arroz",
	27: "🍛 Teriyaki Pollo + 🥤 Té + 🍚 Arroz",
	28: "🍛 Teriyaki Salmón + 🥤 Té + 🍚 Arroz",
	29: "🍲 Bibimbap + 🥤 Té + 🥟 Kimchi",
}

func (arbol *ArbolDecision) Predict(filas [][]float64) []int {
	predicciones := make([]int, len(filas))
	for i, fila := range filas {
		predicciones[i] = arbol.prediccionUnica(fila, arbol.Raiz)
	}
	return predicciones
}

func (arbol *ArbolDecision) prediccionUnica(fila []float64, nodo *Nodo) int {
	if nodo.EsHoja {
		return nodo.Prediccion
	}

	// Para características numéricas (calorías)
	if nodo.Indice == 4 {
		if fila[nodo.Indice] <= nodo.Umbral {
			return arbol.prediccionUnica(fila, nodo.Izquierdo)
		}
		return arbol.prediccionUnica(fila, nodo.Derecho)
	}

	// Para características categóricas
	if fila[nodo.Indice] == float64(nodo.Valor[0]-'0') {
		return arbol.prediccionUnica(fila, nodo.Izquierdo)
	}
	return arbol.prediccionUnica(fila, nodo.Derecho)
}

func EntrenarArbol(datos [][]float64, etiquetas []int, profundidad int) *ArbolDecision {
	return &ArbolDecision{
		Raiz: construirNodo(datos, etiquetas, profundidad),
	}
}

func construirNodo(datos [][]float64, etiquetas []int, profundidad int) *Nodo {
	// Si todas las etiquetas son iguales o profundidad máxima
	claseUnica := true
	clase := etiquetas[0]
	for _, e := range etiquetas {
		if e != clase {
			claseUnica = false
			break
		}
	}

	if claseUnica || profundidad >= 8 || len(datos) < 3 {
		return &Nodo{EsHoja: true, Prediccion: modaEtiquetas(etiquetas)}
	}

	// Encontrar mejor división
	mejorGanancia := -1.0
	mejorIndice := -1
	mejorValor := ""
	mejorUmbral := 0.0
	mejorIzqEtqs := []int{}
	mejorDerEtqs := []int{}

	// Probar cada característica
	for i := 0; i < 5; i++ {
		if i == 4 { // Calorías (numérica)
			// Probar diferentes umbrales
			valores := make([]float64, len(datos))
			for j := 0; j < len(datos); j++ {
				valores[j] = datos[j][i]
			}

			for _, umbral := range valores {
				var izqEtqs, derEtqs []int
				for j := 0; j < len(datos); j++ {
					if datos[j][i] <= umbral {
						izqEtqs = append(izqEtqs, etiquetas[j])
					} else {
						derEtqs = append(derEtqs, etiquetas[j])
					}
				}
				if len(izqEtqs) > 0 && len(derEtqs) > 0 {
					ganancia := gananciaInformacion(etiquetas, izqEtqs, derEtqs)
					if ganancia > mejorGanancia {
						mejorGanancia = ganancia
						mejorIndice = i
						mejorUmbral = umbral
						mejorIzqEtqs = izqEtqs
						mejorDerEtqs = derEtqs
						mejorValor = ""
					}
				}
			}
		} else { // Categóricas
			// Probar cada valor posible (0,1,2)
			for valor := 0; valor <= 2; valor++ {
				var izqEtqs, derEtqs []int
				for j := 0; j < len(datos); j++ {
					if int(datos[j][i]) == valor {
						izqEtqs = append(izqEtqs, etiquetas[j])
					} else {
						derEtqs = append(derEtqs, etiquetas[j])
					}
				}
				if len(izqEtqs) > 0 && len(derEtqs) > 0 {
					ganancia := gananciaInformacion(etiquetas, izqEtqs, derEtqs)
					if ganancia > mejorGanancia {
						mejorGanancia = ganancia
						mejorIndice = i
						mejorValor = string(rune(valor + '0'))
						mejorUmbral = 0
						mejorIzqEtqs = izqEtqs
						mejorDerEtqs = derEtqs
					}
				}
			}
		}
	}

	if mejorIndice == -1 {
		return &Nodo{EsHoja: true, Prediccion: modaEtiquetas(etiquetas)}
	}

	// Dividir datos según la mejor condición
	var izqDatos, derDatos [][]float64

	for j := 0; j < len(datos); j++ {
		if mejorIndice == 4 { // Calorías
			if datos[j][mejorIndice] <= mejorUmbral {
				izqDatos = append(izqDatos, datos[j])
			} else {
				derDatos = append(derDatos, datos[j])
			}
		} else { // Categóricas
			if int(datos[j][mejorIndice]) == int(mejorValor[0]-'0') {
				izqDatos = append(izqDatos, datos[j])
			} else {
				derDatos = append(derDatos, datos[j])
			}
		}
	}

	return &Nodo{
		EsHoja:    false,
		Indice:    mejorIndice,
		Valor:     mejorValor,
		Umbral:    mejorUmbral,
		Izquierdo: construirNodo(izqDatos, mejorIzqEtqs, profundidad+1),
		Derecho:   construirNodo(derDatos, mejorDerEtqs, profundidad+1),
	}
}

func gananciaInformacion(padre, izquierdo, derecho []int) float64 {
	total := float64(len(padre))
	totalIzq := float64(len(izquierdo))
	totalDer := float64(len(derecho))

	if total == 0 {
		return 0
	}

	entropiaPadre := entropia(padre)
	entropiaPonderada := (totalIzq/total)*entropia(izquierdo) + (totalDer/total)*entropia(derecho)

	return entropiaPadre - entropiaPonderada
}

func entropia(etiquetas []int) float64 {
	if len(etiquetas) == 0 {
		return 0
	}
	conteo := make(map[int]int)
	for _, e := range etiquetas {
		conteo[e]++
	}
	ent := 0.0
	total := float64(len(etiquetas))
	for _, c := range conteo {
		prob := float64(c) / total
		if prob > 0 {
			ent -= prob * math.Log2(prob)
		}
	}
	return ent
}

func modaEtiquetas(etiquetas []int) int {
	if len(etiquetas) == 0 {
		return 0
	}
	conteo := make(map[int]int)
	maxCount := 0
	moda := etiquetas[0]
	for _, e := range etiquetas {
		conteo[e]++
		if conteo[e] > maxCount {
			maxCount = conteo[e]
			moda = e
		}
	}
	return moda
}

func (arbol *ArbolDecision) ImprimirArbol() {
	imprimirNodo(arbol.Raiz, 0)
}

func imprimirNodo(nodo *Nodo, nivel int) {
	indentacion := ""
	for i := 0; i < nivel; i++ {
		indentacion += "  "
	}

	if nodo.EsHoja {
		fmt.Printf("%s→ %s\n", indentacion, PrediccionMap[nodo.Prediccion])
		return
	}

	nombreCaract := ""
	switch nodo.Indice {
	case 0:
		nombreCaract = "Presupuesto"
	case 1:
		nombreCaract = "Tipo de comida"
	case 2:
		nombreCaract = "Ocasión"
	case 3:
		nombreCaract = "Vegano"
	case 4:
		nombreCaract = "Calorías máximas"
	}

	if nodo.Indice == 4 {
		fmt.Printf("%s¿%s <= %.0f?\n", indentacion, nombreCaract, nodo.Umbral)
	} else {
		valorNombre := ""
		switch nodo.Valor {
		case "0":
			if nodo.Indice == 0 {
				valorNombre = "bajo"
			} else if nodo.Indice == 1 {
				valorNombre = "italiana"
			} else if nodo.Indice == 2 {
				valorNombre = "familiar"
			} else {
				valorNombre = "no"
			}
		case "1":
			if nodo.Indice == 0 {
				valorNombre = "medio"
			} else if nodo.Indice == 1 {
				valorNombre = "mexicana"
			} else if nodo.Indice == 2 {
				valorNombre = "cita"
			} else {
				valorNombre = "si"
			}
		case "2":
			if nodo.Indice == 0 {
				valorNombre = "alto"
			} else if nodo.Indice == 1 {
				valorNombre = "asiatica"
			} else if nodo.Indice == 2 {
				valorNombre = "negocio"
			}
		}
		fmt.Printf("%s¿%s = %s?\n", indentacion, nombreCaract, valorNombre)
	}
	fmt.Printf("%s  ├── Sí → ", indentacion)
	imprimirNodo(nodo.Izquierdo, nivel+1)
	fmt.Printf("%s  └── No → ", indentacion)
	imprimirNodo(nodo.Derecho, nivel+1)
}

// Agrega esta función al final del archivo modelo/arbol.go
func (arbol *ArbolDecision) RecomendarPorCercania(caracteristicas []float64) string {
	// Si el árbol devuelve -1 o no encuentra, dar recomendación por defecto
	prediccion := arbol.Predict([][]float64{caracteristicas})
	if prediccion[0] >= 0 && prediccion[0] < len(PrediccionMap) {
		return PrediccionMap[prediccion[0]]
	}

	// Recomendación por defecto basada en el tipo de comida
	tipoComida := int(caracteristicas[1])
	switch tipoComida {
	case 0: // Italiana
		return "🍝 Pasta del Chef + 🥂 Vino de la casa + 🍮 Tiramisú"
	case 1: // Mexicana
		return "🌮 Tacos de la casa + 🥤 Agua fresca + 🍩 Churros"
	case 2: // Asiatica
		return "🍜 Ramen especial + 🥤 Té verde + 🥟 Gyozas"
	default:
		return "🍽️ Menú ejecutivo del día + 🥤 Bebida + 🍰 Postre"
	}
}
