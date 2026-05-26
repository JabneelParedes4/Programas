package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"
)

type Combinacion struct {
	Presupuesto string
	TipoComida  string
	Ocasion     string
	Vegano      string
	Calorias    int
	Platillo    string
}

var baseDatos []Combinacion

func main() {
	fmt.Println("============================================")
	fmt.Println("   🍽️  SISTEMA DE RECOMENDACIÓN CULINARIA  🍽️")
	fmt.Println("============================================")

	// Intentar cargar modelo guardado
	fmt.Println("\n[1/3] Buscando modelo guardado...")
	datosCargados, err := CargarModelo("guardado/recomendador.gob")

	if err != nil {
		fmt.Println("       ⚠ No se encontró modelo guardado. Entrenando uno nuevo...")

		// Cargar datos del CSV
		fmt.Println("\n[2/3] Cargando base de datos desde CSV...")
		cargarDatos("datos/combinaciones.csv")
		fmt.Printf("       ✓ %d combinaciones cargadas\n", len(baseDatos))

		// Guardar modelo entrenado
		fmt.Println("\n[3/3] Guardando modelo para uso futuro...")
		err = GuardarModelo(baseDatos, "guardado/recomendador.gob")
		if err != nil {
			log.Fatal("Error al guardar modelo:", err)
		}
		fmt.Println("       ✓ Modelo guardado exitosamente en guardado/recomendador.gob")
	} else {
		fmt.Printf("       ✓ Modelo cargado desde disco! (%d combinaciones)\n", len(datosCargados))
		baseDatos = datosCargados
	}

	// Sistema interactivo
	fmt.Println("\n🎯 SISTEMA DE RECOMENDACIÓN ACTIVO")
	recomendarInteractivo()
}

// Guardar modelo en archivo
func GuardarModelo(datos []Combinacion, ruta string) error {
	// Crear carpeta guardado si no existe
	if err := os.MkdirAll("guardado", 0755); err != nil {
		return err
	}

	archivo, err := os.Create(ruta)
	if err != nil {
		return err
	}
	defer archivo.Close()

	encoder := gob.NewEncoder(archivo)
	return encoder.Encode(datos)
}

// Cargar modelo desde archivo
func CargarModelo(ruta string) ([]Combinacion, error) {
	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer archivo.Close()

	var datos []Combinacion
	decoder := gob.NewDecoder(archivo)
	err = decoder.Decode(&datos)
	return datos, err
}

func cargarDatos(ruta string) {
	archivo, err := os.Open(ruta)
	if err != nil {
		log.Fatal("Error al abrir archivo:", err)
	}
	defer archivo.Close()

	reader := csv.NewReader(archivo)
	registros, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error al leer CSV:", err)
	}

	for i, record := range registros {
		if i == 0 {
			continue
		}
		if len(record) >= 6 {
			calorias := 0
			fmt.Sscanf(record[4], "%d", &calorias)

			baseDatos = append(baseDatos, Combinacion{
				Presupuesto: strings.TrimSpace(record[0]),
				TipoComida:  strings.TrimSpace(record[1]),
				Ocasion:     strings.TrimSpace(record[2]),
				Vegano:      strings.TrimSpace(record[3]),
				Calorias:    calorias,
				Platillo:    record[5],
			})
		}
	}
}

func recomendarInteractivo() {
	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║   RESPUESTAS RÁPIDAS (responde con #)   ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Println("║ Presupuesto:  1=Bajo  2=Medio  3=Alto   ║")
	fmt.Println("║ Tipo comida:  1=Italiana 2=Mexicana 3=Asiatica ║")
	fmt.Println("║ Ocasión:      1=Familiar 2=Cita 3=Negocio ║")
	fmt.Println("║ Vegano:       1=No  2=Sí               ║")
	fmt.Println("║ Calorías:     Número (ej: 500)         ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	for {
		fmt.Println("\n" + strings.Repeat("─", 50))

		presupuesto := preguntarOpcion("💰 Presupuesto (1=Bajo, 2=Medio, 3=Alto): ", 3)
		tipoComida := preguntarOpcion("🍽️  Tipo de comida (1=Italiana, 2=Mexicana, 3=Asiatica): ", 3)
		ocasion := preguntarOpcion("🎉 Ocasión (1=Familiar, 2=Cita, 3=Negocio): ", 3)
		vegano := preguntarOpcion("🌱 Vegano (1=No, 2=Sí): ", 2)

		fmt.Print("🔥 Calorías máximas: ")
		var caloriasMax int
		fmt.Scan(&caloriasMax)

		// Convertir números a texto
		presupuestoTexto := map[int]string{1: "bajo", 2: "medio", 3: "alto"}[presupuesto]
		tipoTexto := map[int]string{1: "italiana", 2: "mexicana", 3: "asiatica"}[tipoComida]
		ocasionTexto := map[int]string{1: "familiar", 2: "cita", 3: "negocio"}[ocasion]
		veganoTexto := map[int]string{1: "no", 2: "si"}[vegano]

		// Buscar recomendación
		recomendacion := buscarRecomendacion(presupuestoTexto, tipoTexto, ocasionTexto, veganoTexto, caloriasMax)

		// Mostrar resultado
		fmt.Println("\n" + strings.Repeat("⭐", 40))
		fmt.Println("   🍽️  RECOMENDACIÓN DEL CHEF:")
		fmt.Printf("\n   %s\n", recomendacion)
		fmt.Println("\n" + strings.Repeat("⭐", 40))

		fmt.Print("\n¿Otra recomendación? (s/n): ")
		var continuar string
		fmt.Scan(&continuar)
		if continuar != "s" && continuar != "S" {
			fmt.Println("\n¡Buen provecho! 🍽️")
			break
		}
	}
}

func buscarRecomendacion(presupuesto, tipoComida, ocasion, vegano string, caloriasMax int) string {
	// Primero buscar coincidencia exacta
	for _, combo := range baseDatos {
		if combo.Presupuesto == presupuesto &&
			combo.TipoComida == tipoComida &&
			combo.Ocasion == ocasion &&
			combo.Vegano == vegano &&
			combo.Calorias <= caloriasMax {
			return formatearPlatillo(combo.Platillo)
		}
	}

	// Si no hay coincidencia exacta, buscar sin filtro de calorías
	for _, combo := range baseDatos {
		if combo.Presupuesto == presupuesto &&
			combo.TipoComida == tipoComida &&
			combo.Ocasion == ocasion &&
			combo.Vegano == vegano {
			return formatearPlatillo(combo.Platillo)
		}
	}

	// Si no hay coincidencia, dar recomendación por defecto
	return recomendacionPorDefecto(tipoComida, vegano)
}

func formatearPlatillo(platillo string) string {
	menu := map[string]string{
		"pizza_margarita":       "🍕 Pizza Margarita + 🥤 Refresco + 🍰 Tiramisú",
		"pizza_pepperoni":       "🍕 Pizza Pepperoni + 🥤 Refresco + 🍰 Tiramisú",
		"espagueti_bolognesa":   "🍝 Espagueti Bolognesa + 🥤 Refresco + 🍮 Panna Cotta",
		"pasta_carbonara":       "🍝 Pasta Carbonara + 🍷 Vino tinto + 🍮 Panna Cotta",
		"lasagna_pequena":       "🍲 Lasagna Pequeña + 🥤 Refresco + 🍰 Tiramisú",
		"ravioles_queso":        "🥟 Ravioles de Queso + 🥤 Refresco + 🍮 Flan",
		"lasagna":               "🍲 Lasagna + 🥗 Ensalada + 🍫 Tiramisú",
		"risotto_setas":         "🍚 Risotto de Setas + 🥂 Espumante + 🍓 Frutillas",
		"ravioles_espinaca":     "🥟 Ravioles de Espinaca + 🥤 Jugo + 🍎 Tarta",
		"osobuco":               "🍖 Osobuco + 🍷 Vino tinto + 🍮 Tiramisú",
		"filete_parmesano":      "🥩 Filete al Parmesano + 🍷 Vino + 🍰 Tiramisú",
		"ensalada_caprese":      "🥗 Ensalada Caprese + 🥂 Limonada + 🍓 Frutillas",
		"filete_trufa":          "🥩 Filete con Trufa + 🍷 Vino + 🍫 Tiramisú",
		"risotto_vegano":        "🍚 Risotto Vegano + 🥤 Jugo + 🍓 Frutillas",
		"penne_arrabiata":       "🍝 Penne Arrabiata + 🥤 Refresco + 🍮 Panna Cotta",
		"tacos_pastor":          "🌮 Tacos al Pastor + 🥤 Agua de Jamaica + 🍩 Churros",
		"enchiladas_verdes":     "🌮 Enchiladas Verdes + 🥤 Refresco + 🍩 Buñuelos",
		"quesadillas":           "🧀 Quesadillas + 🥑 Guacamole + 🍮 Flan",
		"mole_poblano":          "🍛 Mole Poblano + 🥤 Horchata + 🍩 Buñuelos",
		"enchiladas_suizas":     "🌯 Enchiladas Suizas + 🥑 Guacamole + 🍮 Arroz con leche",
		"tostadas_veganas":      "🌮 Tostadas Veganas + 🥤 Agua de Jamaica + 🍉 Frutas",
		"ceviche_vegano":        "🐟 Ceviche Vegano + 🥤 Limonada + 🍍 Frutas",
		"carne_asada":           "🥩 Carne Asada + 🥤 Horchata + 🍩 Churros",
		"pozole":                "🍲 Pozole + 🥤 Refresco + 🍩 Buñuelos",
		"cochinita_pibil":       "🐷 Cochinita Pibil + 🥤 Horchata + 🍮 Flan",
		"chiles_en_nogada":      "🌶️ Chiles en Nogada + 🥤 Agua + 🍮 Arroz con leche",
		"lomo_enchilado":        "🥩 Lomo Enchilado + 🥤 Horchata + 🍩 Churros",
		"camarones_a_la_diabla": "🍤 Camarones a la Diabla + 🥤 Refresco + 🍮 Flan",
		"tacos_veganos":         "🌮 Tacos Veganos + 🥤 Agua de Jamaica + 🍉 Frutas",
		"enchiladas_veganas":    "🌯 Enchiladas Veganas + 🥤 Agua + 🍉 Frutas",
		"arroz_frito":           "🍚 Arroz Frito + 🥤 Té + 🥟 Rollitos primavera",
		"chow_mein":             "🍜 Chow Mein + 🥤 Té + 🥟 Wantanes",
		"gyozas":                "🥟 Gyozas + 🥤 Té + 🍚 Arroz",
		"rollos_primavera":      "🥟 Rollos Primavera + 🥤 Té + 🍚 Arroz Frito",
		"ramen_basico":          "🍜 Ramen Tradicional + 🥤 Té + 🥟 Gyozas",
		"udon":                  "🍜 Udon + 🥤 Té + 🥟 Tempura",
		"tailandes_verde":       "🥘 Tailandés Verde + 🥤 Agua + 🍚 Arroz jazmín",
		"ramen_vegetariano":     "🍜 Ramen Vegetariano + 🥤 Té + 🥟 Gyozas",
		"sushi_mixto":           "🍣 Sushi Mixto + 🍶 Sake + 🍡 Mochi",
		"pad_thai":              "🍜 Pad Thai + 🥤 Limonada + 🥟 Rollitos",
		"pad_thai_camaron":      "🍤 Pad Thai con Camarón + 🥤 Limonada + 🥟 Rollitos",
		"curry_amarillo":        "🍛 Curry Amarillo + 🥤 Té + 🍚 Arroz",
		"teriyaki_pollo":        "🍗 Teriyaki Pollo + 🥤 Té + 🍚 Arroz",
		"teriyaki_salmon":       "🐟 Teriyaki Salmón + 🥤 Té + 🍚 Arroz",
		"bibimbap":              "🍲 Bibimbap + 🥤 Té + 🥟 Kimchi",
		"sashimi_variado":       "🐟 Sashimi Variado + 🍶 Sake + 🍡 Mochi",
		"curry_real":            "👑 Curry Real Tailandés + 🥤 Té + 🍚 Arroz jazmín",
		"sushi_premium":         "🍣 Sushi Premium + 🍶 Sake + 🍡 Mochi",
		"rollos_veganos":        "🥗 Rollos Veganos + 🥤 Té + 🥟 Gyozas",
		"ramen_vegano_premium":  "🍜 Ramen Vegano Premium + 🥤 Té + 🥟 Gyozas",
		"curry_rojo_vegano":     "🍛 Curry Rojo Vegano + 🥤 Té + 🍚 Arroz jazmín",
		"pad_thai_vegano":       "🍜 Pad Thai Vegano + 🥤 Limonada + 🥟 Rollitos",
		"arroz_tailandes":       "🍚 Arroz Tailandés + 🥤 Té + 🥟 Rollitos",
	}

	if val, ok := menu[platillo]; ok {
		return val
	}

	return fmt.Sprintf("📋 %s", strings.ReplaceAll(platillo, "_", " "))
}

func recomendacionPorDefecto(tipoComida, vegano string) string {
	if vegano == "si" {
		return "🥗 Ensalada Vegana Especial + 🥤 Jugo Natural + 🍎 Frutas de temporada"
	}

	switch tipoComida {
	case "italiana":
		return "🍝 Pasta de la Casa + 🍷 Vino de la Casa + 🍰 Tiramisú"
	case "mexicana":
		return "🌮 Combinación Mexicana + 🥤 Horchata + 🍩 Churros"
	case "asiatica":
		return "🍜 Sopa Ramen + 🥤 Té Verde + 🥟 Gyozas"
	default:
		return "🍽️ Menú Ejecutivo del Día + 🥤 Bebida + 🍮 Postre"
	}
}

func preguntarOpcion(mensaje string, max int) int {
	var opcion int
	for {
		fmt.Print(mensaje)
		_, err := fmt.Scan(&opcion)
		if err != nil {
			fmt.Println("❌ Error: ingresa un número válido")
			var descartar string
			fmt.Scan(&descartar)
			continue
		}
		if opcion >= 1 && opcion <= max {
			return opcion
		}
		fmt.Printf("❌ Opción inválida. Elige 1-%d\n", max)
	}
}
