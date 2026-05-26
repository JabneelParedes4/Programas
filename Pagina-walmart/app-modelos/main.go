package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ProductoDestacado struct {
	Producto string `json:"producto"`
	Ventas   int    `json:"ventas"`
}

type ProductoInfo struct {
	Producto string  `json:"producto"`
	Precio   float64 `json:"precio"`
}

type ReglaApriori struct {
	Antecedente []string `json:"antecedente"`
	Consecuente []string `json:"consecuente"`
	Soporte     float64  `json:"soporte"`
	Confianza   float64  `json:"confianza"`
}

type RecomendacionProducto struct {
	Producto string  `json:"producto"`
	Precio   float64 `json:"precio"`
	Score    float64 `json:"score"`
}

type ModeloRecomendaciones struct {
	Productos []string             `json:"productos"`
	Reglas    []ReglaApriori       `json:"reglas"`
	Top       []ProductoDestacado   `json:"top_productos"`
	Tiendas   []string              `json:"tiendas"`
	Precios   map[string]float64    `json:"precios"`
}

type CasoKNN struct {
	Personas int      `json:"personas"`
	M2       int      `json:"m2"`
	Genero   string   `json:"genero"`
	Despensa []string `json:"despensa"`
	Hogar    []string `json:"hogar"`
	Personal []string `json:"personal"`
}

type ModeloKNN struct {
	Casos []CasoKNN `json:"casos"`
}

type DespensaRequest struct {
	Personas int    `json:"personas"`
	M2       int    `json:"m2"`
	Genero   string `json:"genero"`
}

type HoraPico struct {
	Periodo string `json:"periodo"`
	Hora    string `json:"hora"`
	Ventas  int    `json:"ventas"`
}

type ModeloHorarios struct {
	Dia    []HoraPico `json:"dia"`
	Semana []HoraPico `json:"semana"`
	Mes    []HoraPico `json:"mes"`
}

type PeriodosDisponibles struct {
	Meses   []string `json:"meses"`
	Semanas []string `json:"semanas"`
}

type ConcurrenciaPeriodoDia struct {
	Fecha       string         `json:"fecha"`
	Dia         string         `json:"dia"`
	Horas       map[string]int `json:"horas"`
	HoraPico    string         `json:"hora_pico"`
	TicketsPico int            `json:"tickets_pico"`
	EsFestivo   bool           `json:"es_festivo"`
	EsQuincena  bool           `json:"es_quincena"`
	TotalDia    int            `json:"total_dia"`
}

type ConcurrenciaDia struct {
	Dia   string         `json:"dia"`
	Horas map[string]int `json:"horas"`
}

type SegmentoCliente struct {
	Segmento      string   `json:"segmento"`
	Tickets       int      `json:"tickets"`
	HoraPico      string   `json:"hora_pico"`
	ProductoTop   string   `json:"producto_top"`
	MetodoPagoTop string   `json:"metodo_pago_top"`
	Genero        string   `json:"genero"`
	TipoCliente   string   `json:"tipo_cliente"`
	City          string   `json:"city"`
	TotalVentas   float64  `json:"total_ventas"`
	GastoPromedio float64  `json:"gasto_promedio"`
	HoraCritica   string   `json:"hora_critica"`
	ProductosPico []string `json:"productos_pico"`
}

type RespuestaClientes struct {
	Segmentos       []SegmentoCliente `json:"segmentos"`
	ClientesUnicos int               `json:"clientes_unicos"`
}

type ModeloClientes struct {
	Segmentos []SegmentoCliente `json:"segmentos"`
}

var modelo ModeloRecomendaciones
var modeloKNN ModeloKNN
var modeloHorarios ModeloHorarios
var modeloClientes ModeloClientes

func guardarJSON(ruta string, data interface{}) error {
	file, err := os.Create(ruta)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func responderErrorJSON(w http.ResponseWriter, mensaje string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":    false,
		"error": mensaje,
	})
}

func cargarCSV(ruta string) (map[string][]string, map[string]int, map[string]bool, map[string]float64, error) {
	file, err := os.Open(ruta)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	transacciones := make(map[string][]string)
	frecuencia := make(map[string]int)
	tiendas := make(map[string]bool)

	sumaPrecios := make(map[string]float64)
	conteoPrecios := make(map[string]int)
	preciosPromedio := make(map[string]float64)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 8 {
			continue
		}

		invoiceID := strings.TrimSpace(row[0])
		tienda := strings.TrimSpace(row[3])
		producto := strings.ToLower(strings.TrimSpace(row[4]))

		precioTexto := strings.TrimSpace(row[7])
		precioTexto = strings.ReplaceAll(precioTexto, "$", "")
		precioTexto = strings.ReplaceAll(precioTexto, ",", "")

		precio, _ := strconv.ParseFloat(precioTexto, 64)

		if invoiceID == "" || producto == "" {
			continue
		}

		transacciones[invoiceID] = append(transacciones[invoiceID], producto)
		frecuencia[producto]++

		if precio > 0 {
			sumaPrecios[producto] += precio
			conteoPrecios[producto]++
		}

		if tienda != "" {
			tiendas[tienda] = true
		}
	}

	for producto, suma := range sumaPrecios {
		if conteoPrecios[producto] > 0 {
			preciosPromedio[producto] = suma / float64(conteoPrecios[producto])
		}
	}

	return transacciones, frecuencia, tiendas, preciosPromedio, nil
}

func obtenerProductos(frecuencia map[string]int) []string {
	var productos []string
	for p := range frecuencia {
		productos = append(productos, p)
	}
	sort.Strings(productos)
	return productos
}

func obtenerTopProductos(frecuencia map[string]int, limite int) []ProductoDestacado {
	var lista []ProductoDestacado

	for producto, ventas := range frecuencia {
		lista = append(lista, ProductoDestacado{
			Producto: producto,
			Ventas:   ventas,
		})
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Ventas > lista[j].Ventas
	})

	if len(lista) > limite {
		lista = lista[:limite]
	}

	return lista
}

func obtenerTiendas(tiendasSet map[string]bool) []string {
	var tiendas []string
	for t := range tiendasSet {
		tiendas = append(tiendas, t)
	}
	sort.Strings(tiendas)
	return tiendas
}

func contieneProducto(productos []string, producto string) bool {
	producto = strings.ToLower(strings.TrimSpace(producto))

	for _, p := range productos {
		if strings.ToLower(strings.TrimSpace(p)) == producto {
			return true
		}
	}
	return false
}

func contieneItemset(productos []string, itemset []string) bool {
	for _, item := range itemset {
		if !contieneProducto(productos, item) {
			return false
		}
	}
	return true
}

func calcularSoporte(itemset []string, transacciones map[string][]string) float64 {
	if len(transacciones) == 0 {
		return 0
	}

	count := 0

	for _, productos := range transacciones {
		if contieneItemset(productos, itemset) {
			count++
		}
	}

	return float64(count) / float64(len(transacciones))
}

func generarReglasApriori(transacciones map[string][]string, productos []string) []ReglaApriori {
	var reglas []ReglaApriori

	minSupport := 0.001
	minConfidence := 0.10

	for i := 0; i < len(productos); i++ {
		for j := i + 1; j < len(productos); j++ {
			a := productos[i]
			b := productos[j]

			itemset := []string{a, b}
			soporteAB := calcularSoporte(itemset, transacciones)

			if soporteAB < minSupport {
				continue
			}

			soporteA := calcularSoporte([]string{a}, transacciones)
			soporteB := calcularSoporte([]string{b}, transacciones)

			if soporteA > 0 {
				confianzaAB := soporteAB / soporteA
				if confianzaAB >= minConfidence {
					reglas = append(reglas, ReglaApriori{
						Antecedente: []string{a},
						Consecuente: []string{b},
						Soporte:     soporteAB,
						Confianza:   confianzaAB,
					})
				}
			}

			if soporteB > 0 {
				confianzaBA := soporteAB / soporteB
				if confianzaBA >= minConfidence {
					reglas = append(reglas, ReglaApriori{
						Antecedente: []string{b},
						Consecuente: []string{a},
						Soporte:     soporteAB,
						Confianza:   confianzaBA,
					})
				}
			}
		}
	}

	sort.Slice(reglas, func(i, j int) bool {
		return reglas[i].Confianza > reglas[j].Confianza
	})

	return reglas
}

func entrenarModeloRecomendaciones() error {
	if err := os.MkdirAll("modelos", os.ModePerm); err != nil {
		return err
	}

	transacciones, frecuencia, tiendasSet, precios, err := cargarCSV("uploads/dataset_walmart.csv")
	if err != nil {
		return err
	}

	productos := obtenerProductos(frecuencia)
	reglas := generarReglasApriori(transacciones, productos)

	modelo = ModeloRecomendaciones{
		Productos: productos,
		Reglas:    reglas,
		Top:       obtenerTopProductos(frecuencia, 5),
		Tiendas:   obtenerTiendas(tiendasSet),
		Precios:   precios,
	}

	return guardarJSON("modelos/modelo_recomendaciones.model", modelo)
}

func cargarModeloRecomendaciones() {
	file, err := os.Open("modelos/modelo_recomendaciones.model")
	if err != nil {
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&modelo)
}

func recomendar(carrito []string) []RecomendacionProducto {
	score := make(map[string]float64)

	for _, regla := range modelo.Reglas {
		if contieneItemset(carrito, regla.Antecedente) {
			for _, producto := range regla.Consecuente {
				if contieneProducto(carrito, producto) {
					continue
				}

				score[producto] += regla.Confianza
			}
		}
	}

	var lista []RecomendacionProducto

	for producto, valor := range score {
		lista = append(lista, RecomendacionProducto{
			Producto: producto,
			Precio:   modelo.Precios[producto],
			Score:    valor,
		})
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Score > lista[j].Score
	})

	limite := 5
	if len(lista) < limite {
		limite = len(lista)
	}

	return lista[:limite]
}

func entrenarModeloKNN() error {
	if err := os.MkdirAll("modelos", os.ModePerm); err != nil {
		return err
	}

	modeloKNN = ModeloKNN{
		Casos: []CasoKNN{
			{Personas: 1, M2: 40, Genero: "H", Despensa: []string{"arroz", "frijol", "huevo", "leche", "tortillas"}, Hogar: []string{"detergente", "papel higiénico"}, Personal: []string{}},
			{Personas: 2, M2: 60, Genero: "M", Despensa: []string{"leche", "pan blanco", "huevo", "yogurt", "galletas"}, Hogar: []string{"detergente", "papel higiénico"}, Personal: []string{}},
			{Personas: 3, M2: 80, Genero: "H-M", Despensa: []string{"arroz", "frijol", "leche", "pan blanco", "queso", "tortillas"}, Hogar: []string{"detergente", "papel higiénico"}, Personal: []string{}},
			{Personas: 4, M2: 100, Genero: "H-M", Despensa: []string{"arroz", "frijol", "huevo", "leche", "pan blanco", "pan dulce", "tortillas"}, Hogar: []string{"detergente", "papel higiénico"}, Personal: []string{}},
			{Personas: 5, M2: 130, Genero: "H-M", Despensa: []string{"arroz", "frijol", "huevo", "leche", "queso", "tortillas", "galletas"}, Hogar: []string{"detergente", "papel higiénico"}, Personal: []string{}},
			{Personas: 6, M2: 160, Genero: "H-M", Despensa: []string{"arroz", "frijol", "huevo", "leche", "pan blanco", "queso", "tortillas", "sabritas"}, Hogar: []string{"detergente", "papel higiénico"}, Personal: []string{}},
		},
	}

	return guardarJSON("modelos/knn_despensa.model", modeloKNN)
}

func cargarModeloKNN() {
	file, err := os.Open("modelos/knn_despensa.model")
	if err != nil {
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&modeloKNN)
}

func distanciaKNN(req DespensaRequest, caso CasoKNN) float64 {
	dPersonas := float64(req.Personas - caso.Personas)
	dM2 := float64(req.M2-caso.M2) / 20.0

	dGenero := 0.0
	if strings.ToUpper(req.Genero) != strings.ToUpper(caso.Genero) {
		dGenero = 1.5
	}

	return math.Sqrt(dPersonas*dPersonas + dM2*dM2 + dGenero*dGenero)
}

func ordenarScore(score map[string]int) []string {
	type Par struct {
		Producto string
		Score    int
	}

	var lista []Par

	for p, s := range score {
		lista = append(lista, Par{Producto: p, Score: s})
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Score > lista[j].Score
	})

	var resultado []string

	for _, item := range lista {
		resultado = append(resultado, item.Producto)
	}

	return resultado
}

func recomendarDespensaKNN(req DespensaRequest) map[string][]string {
	type Vecino struct {
		Caso      CasoKNN
		Distancia float64
	}

	var vecinos []Vecino

	for _, caso := range modeloKNN.Casos {
		vecinos = append(vecinos, Vecino{
			Caso:      caso,
			Distancia: distanciaKNN(req, caso),
		})
	}

	sort.Slice(vecinos, func(i, j int) bool {
		return vecinos[i].Distancia < vecinos[j].Distancia
	})

	k := 3
	if len(vecinos) < k {
		k = len(vecinos)
	}

	despensaScore := make(map[string]int)
	hogarScore := make(map[string]int)
	personalScore := make(map[string]int)

	for i := 0; i < k; i++ {
		for _, p := range vecinos[i].Caso.Despensa {
			despensaScore[p]++
		}
		for _, p := range vecinos[i].Caso.Hogar {
			hogarScore[p]++
		}
		for _, p := range vecinos[i].Caso.Personal {
			personalScore[p]++
		}
	}

	return map[string][]string{
		"despensa": ordenarScore(despensaScore),
		"hogar":    ordenarScore(hogarScore),
		"personal": ordenarScore(personalScore),
	}
}

func obtenerClaveFecha(fechaTexto string, tipo string) string {
	fechaTexto = strings.TrimSpace(fechaTexto)

	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		"02/01/06",
		"2006/01/02",
	}

	var fecha time.Time
	var err error

	for _, layout := range layouts {
		fecha, err = time.Parse(layout, fechaTexto)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fechaTexto
	}

	switch tipo {
	case "dia":
		return fecha.Format("2006-01-02")
	case "semana":
		year, week := fecha.ISOWeek()
		return fmt.Sprintf("%d-S%02d", year, week)
	case "mes":
		return fecha.Format("2006-01")
	default:
		return fecha.Format("2006-01-02")
	}
}

func obtenerHora(horaCompleta string) string {
	horaCompleta = strings.TrimSpace(horaCompleta)

	if horaCompleta == "" {
		return ""
	}

	partes := strings.Split(horaCompleta, ":")
	if len(partes) == 0 {
		return ""
	}

	return partes[0] + ":00"
}

func calcularPicosPorPeriodo(rows [][]string, tipo string) []HoraPico {
	ventas := make(map[string]map[string]map[string]bool)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 3 {
			continue
		}

		invoiceID := strings.TrimSpace(row[0])
		fechaTexto := strings.TrimSpace(row[1])
		horaTexto := strings.TrimSpace(row[2])

		if invoiceID == "" || fechaTexto == "" || horaTexto == "" {
			continue
		}

		periodo := obtenerClaveFecha(fechaTexto, tipo)
		hora := obtenerHora(horaTexto)

		if hora == "" {
			continue
		}

		if _, ok := ventas[periodo]; !ok {
			ventas[periodo] = make(map[string]map[string]bool)
		}

		if _, ok := ventas[periodo][hora]; !ok {
			ventas[periodo][hora] = make(map[string]bool)
		}

		ventas[periodo][hora][invoiceID] = true
	}

	var resultado []HoraPico

	for periodo, horas := range ventas {
		horaTop := ""
		ventasTop := 0

		for hora, tickets := range horas {
			total := len(tickets)

			if total > ventasTop {
				ventasTop = total
				horaTop = hora
			}
		}

		resultado = append(resultado, HoraPico{
			Periodo: periodo,
			Hora:    horaTop,
			Ventas:  ventasTop,
		})
	}

	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].Periodo < resultado[j].Periodo
	})

	return resultado
}

func entrenarModeloHorarios() error {
	if err := os.MkdirAll("modelos", os.ModePerm); err != nil {
		return err
	}

	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	modeloHorarios = ModeloHorarios{
		Dia:    calcularPicosPorPeriodo(rows, "dia"),
		Semana: calcularPicosPorPeriodo(rows, "semana"),
		Mes:    calcularPicosPorPeriodo(rows, "mes"),
	}

	return guardarJSON("modelos/horarios.model", modeloHorarios)
}

func cargarModeloHorarios() {
	file, err := os.Open("modelos/horarios.model")
	if err != nil {
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&modeloHorarios)
}

func parseFechaCSV(fechaTexto string) (time.Time, bool) {
	fechaTexto = strings.TrimSpace(fechaTexto)

	layouts := []string{
		"02/01/06",
		"02/01/2006",
		"2006-01-02",
		"2006/01/02",
	}

	for _, layout := range layouts {
		fecha, err := time.Parse(layout, fechaTexto)
		if err == nil {
			return fecha, true
		}
	}

	return time.Time{}, false
}

func nombreDiaCorto(fecha time.Time) string {
	switch fecha.Weekday() {
	case time.Monday:
		return "LUN"
	case time.Tuesday:
		return "MAR"
	case time.Wednesday:
		return "MIÉ"
	case time.Thursday:
		return "JUE"
	case time.Friday:
		return "VIE"
	case time.Saturday:
		return "SÁB"
	default:
		return "DOM"
	}
}

func esQuincena(fecha time.Time) bool {
	// Solo enero-marzo 2026
	if fecha.Year() != 2026 {
		return false
	}

	mes := fecha.Month()

	if mes != time.January &&
		mes != time.February &&
		mes != time.March {
		return false
	}

	// =========================
	// Primera quincena
	// =========================

	quincena15 := time.Date(
		fecha.Year(),
		fecha.Month(),
		15,
		0, 0, 0, 0,
		time.Local,
	)

	// Si cae sábado -> viernes 14
	if quincena15.Weekday() == time.Saturday {
		quincena15 = quincena15.AddDate(0, 0, -1)
	}

	// Si cae domingo -> viernes 13
	if quincena15.Weekday() == time.Sunday {
		quincena15 = quincena15.AddDate(0, 0, -2)
	}

	// =========================
	// Último día hábil
	// =========================

	ultimoHabil := ultimoDiaHabilDelMes(
		fecha.Year(),
		fecha.Month(),
	)

	// Coincide con quincena
	if mismaFecha(fecha, quincena15) {
		return true
	}

	if mismaFecha(fecha, ultimoHabil) {
		return true
	}

	return false
}

func ultimoDiaHabilDelMes(anio int, mes time.Month) time.Time {
	// Día 0 del siguiente mes = último día del mes actual
	ultimoDia := time.Date(anio, mes+1, 0, 0, 0, 0, 0, time.Local)

	// Si cae sábado, se recorre a viernes
	if ultimoDia.Weekday() == time.Saturday {
		return ultimoDia.AddDate(0, 0, -1)
	}

	// Si cae domingo, se recorre a viernes
	if ultimoDia.Weekday() == time.Sunday {
		return ultimoDia.AddDate(0, 0, -2)
	}

	return ultimoDia
}

func mismaFecha(a, b time.Time) bool {
	return a.Year() == b.Year() &&
		a.Month() == b.Month() &&
		a.Day() == b.Day()
}

func esFestivoMexico(fecha time.Time) bool {
	clave := fecha.Format("01-02")

	festivos := map[string]bool{
		"01-01": true,
		"02-05": true,
		"03-21": true,
		"05-01": true,
		"09-16": true,
		"11-20": true,
		"12-25": true,
	}

	return festivos[clave]
}

func obtenerPeriodosDisponibles() (PeriodosDisponibles, error) {
	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		return PeriodosDisponibles{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return PeriodosDisponibles{}, err
	}

	mesesSet := make(map[string]bool)
	semanasSet := make(map[string]bool)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 2 {
			continue
		}

		fecha, ok := parseFechaCSV(row[1])
		if !ok {
			continue
		}

		mes := fecha.Format("2006-01")
		anio, semana := fecha.ISOWeek()
		semanaTexto := fmt.Sprintf("%d-S%02d", anio, semana)

		mesesSet[mes] = true
		semanasSet[semanaTexto] = true
	}

	var meses []string
	var semanas []string

	for m := range mesesSet {
		meses = append(meses, m)
	}

	for s := range semanasSet {
		semanas = append(semanas, s)
	}

	sort.Strings(meses)
	sort.Strings(semanas)

	return PeriodosDisponibles{
		Meses:   meses,
		Semanas: semanas,
	}, nil
}

func obtenerConcurrenciaPeriodo(tipo string, periodo string, sucursal string) ([]ConcurrenciaPeriodoDia, error) {
	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make(map[string]map[string]map[string]bool)
	totalDia := make(map[string]map[string]bool)
	fechasReal := make(map[string]time.Time)

	sucursal = strings.TrimSpace(sucursal)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 4 {
			continue
		}

		invoiceID := strings.TrimSpace(row[0])
		fechaTexto := strings.TrimSpace(row[1])
		horaTexto := strings.TrimSpace(row[2])
		city := strings.TrimSpace(row[3])

		if sucursal != "" && sucursal != "general" && city != sucursal {
			continue
		}

		if invoiceID == "" || fechaTexto == "" || horaTexto == "" {
			continue
		}

		fecha, ok := parseFechaCSV(fechaTexto)
		if !ok {
			continue
		}

		mes := fecha.Format("2006-01")
		anio, semana := fecha.ISOWeek()
		semanaTexto := fmt.Sprintf("%d-S%02d", anio, semana)

		if tipo == "mes" && mes != periodo {
			continue
		}

		if tipo == "semana" && semanaTexto != periodo {
			continue
		}

		hora := obtenerHora(horaTexto)
		if hora == "" {
			continue
		}

		fechaClave := fecha.Format("2006-01-02")
		fechasReal[fechaClave] = fecha

		if _, ok := data[fechaClave]; !ok {
			data[fechaClave] = make(map[string]map[string]bool)
		}

		if _, ok := data[fechaClave][hora]; !ok {
			data[fechaClave][hora] = make(map[string]bool)
		}

		if _, ok := totalDia[fechaClave]; !ok {
			totalDia[fechaClave] = make(map[string]bool)
		}

		data[fechaClave][hora][invoiceID] = true
		totalDia[fechaClave][invoiceID] = true
	}

	var resultado []ConcurrenciaPeriodoDia

	for fechaClave, horas := range data {
		fecha := fechasReal[fechaClave]

		horasConteo := make(map[string]int)
		horaPico := ""
		ticketsPico := 0

		for h := 0; h <= 23; h++ {
			hora := fmt.Sprintf("%02d:00", h)
			total := len(horas[hora])
			horasConteo[hora] = total

			if total > ticketsPico {
				ticketsPico = total
				horaPico = hora
			}
		}

		resultado = append(resultado, ConcurrenciaPeriodoDia{
			Fecha:       fechaClave,
			Dia:         nombreDiaCorto(fecha),
			Horas:       horasConteo,
			HoraPico:    horaPico,
			TicketsPico: ticketsPico,
			EsFestivo:   esFestivoMexico(fecha),
			EsQuincena:  esQuincena(fecha),
			TotalDia:    len(totalDia[fechaClave]),
		})
	}

	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].Fecha < resultado[j].Fecha
	})

	return resultado, nil
}

func obtenerConcurrenciaSemanal() ([]ConcurrenciaDia, error) {
	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	diasOrden := []string{"LUN", "MAR", "MIÉ", "JUE", "VIE", "SÁB", "DOM"}

	data := make(map[string]map[string]map[string]bool)

	for _, d := range diasOrden {
		data[d] = make(map[string]map[string]bool)
	}

	layouts := []string{
		"02/01/06",
		"02/01/2006",
		"2006-01-02",
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 3 {
			continue
		}

		invoiceID := strings.TrimSpace(row[0])
		fechaTexto := strings.TrimSpace(row[1])
		horaTexto := strings.TrimSpace(row[2])

		if invoiceID == "" || fechaTexto == "" || horaTexto == "" {
			continue
		}

		var fecha time.Time
		var err error

		for _, layout := range layouts {
			fecha, err = time.Parse(layout, fechaTexto)
			if err == nil {
				break
			}
		}

		if err != nil {
			continue
		}

		diaIndex := int(fecha.Weekday())
		// Go: Sunday=0, Monday=1
		var dia string
		switch diaIndex {
		case 1:
			dia = "LUN"
		case 2:
			dia = "MAR"
		case 3:
			dia = "MIÉ"
		case 4:
			dia = "JUE"
		case 5:
			dia = "VIE"
		case 6:
			dia = "SÁB"
		default:
			dia = "DOM"
		}

		partesHora := strings.Split(horaTexto, ":")
		if len(partesHora) == 0 {
			continue
		}

		hora := partesHora[0] + ":00"

		if _, ok := data[dia][hora]; !ok {
			data[dia][hora] = make(map[string]bool)
		}

		data[dia][hora][invoiceID] = true
	}

	var resultado []ConcurrenciaDia

	for _, dia := range diasOrden {
		horas := make(map[string]int)

		for h := 0; h <= 23; h++ {
			hora := fmt.Sprintf("%02d:00", h)
			horas[hora] = len(data[dia][hora])
		}

		resultado = append(resultado, ConcurrenciaDia{
			Dia:   dia,
			Horas: horas,
		})
	}

	return resultado, nil
}

func obtenerMasFrecuente(mapa map[string]int) string {
	type Par struct {
		Clave string
		Valor int
	}

	var lista []Par

	for clave, valor := range mapa {
		lista = append(lista, Par{Clave: clave, Valor: valor})
	}

	if len(lista) == 0 {
		return ""
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Valor > lista[j].Valor
	})

	return lista[0].Clave
}

func top3Keys(m map[string]int) []string {
	type kv struct {
		Key string
		Val int
	}

	var lista []kv

	for k, v := range m {
		lista = append(lista, kv{Key: k, Val: v})
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Val > lista[j].Val
	})

	var resultado []string

	for i := 0; i < len(lista) && i < 3; i++ {
		resultado = append(resultado, lista[i].Key)
	}

	return resultado
}

func entrenarModeloClientes() error {
	if err := os.MkdirAll("modelos", os.ModePerm); err != nil {
		return err
	}

	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return err
	}

	type Acumulador struct {
	Tickets      map[string]bool
	Horas       map[string]int
	Productos   map[string]int
	MetodosPago map[string]int
	Genero      string
	TipoCliente string
	City        string
	TotalVentas float64
}

	segmentos := make(map[string]*Acumulador)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 14 {
			continue
		}

		invoiceID := strings.TrimSpace(row[0])
		horaCompleta := strings.TrimSpace(row[2])
		city := strings.TrimSpace(row[3])
producto := strings.ToLower(strings.TrimSpace(row[4]))
metodoPago := strings.TrimSpace(row[11])
genero := strings.TrimSpace(row[12])
tipoCliente := strings.TrimSpace(row[13])

totalTexto := strings.TrimSpace(row[10])
totalTexto = strings.ReplaceAll(totalTexto, "$", "")
totalTexto = strings.ReplaceAll(totalTexto, ",", "")

totalFloat, err := strconv.ParseFloat(totalTexto, 64)
if err != nil {
	totalFloat = 0
}

		if invoiceID == "" || producto == "" || genero == "" || tipoCliente == "" {
			continue
		}

		partesHora := strings.Split(horaCompleta, ":")
		hora := ""
		if len(partesHora) > 0 && partesHora[0] != "" {
			hora = partesHora[0] + ":00"
		}

		segmento := fmt.Sprintf(
	"%s | %s | %s",
	genero,
	tipoCliente,
	city,
)

		if _, ok := segmentos[segmento]; !ok {
			segmentos[segmento] = &Acumulador{
	Tickets:      make(map[string]bool),
	Horas:       make(map[string]int),
	Productos:   make(map[string]int),
	MetodosPago: make(map[string]int),
	Genero:      genero,
	TipoCliente: tipoCliente,
	City:        city,
	TotalVentas: 0,
}
		}

		segmentos[segmento].Tickets[invoiceID] = true
		segmentos[segmento].Productos[producto]++
		segmentos[segmento].TotalVentas += totalFloat

		if hora != "" {
			segmentos[segmento].Horas[hora]++
		}

		if metodoPago != "" {
			segmentos[segmento].MetodosPago[metodoPago]++
		}
	}

	var resultado []SegmentoCliente

	for nombre, acc := range segmentos {

	tickets := len(acc.Tickets)

	gastoPromedio := 0.0
	if tickets > 0 {
		gastoPromedio = acc.TotalVentas / float64(tickets)
	}

	resultado = append(resultado, SegmentoCliente{
		Segmento:      nombre,
		Tickets:       tickets,
		HoraPico:      obtenerMasFrecuente(acc.Horas),
		ProductoTop:   obtenerMasFrecuente(acc.Productos),
		MetodoPagoTop: obtenerMasFrecuente(acc.MetodosPago),
		Genero:        acc.Genero,
		TipoCliente:   acc.TipoCliente,
		City:          acc.City,
		TotalVentas:   acc.TotalVentas,
		GastoPromedio: gastoPromedio,
		HoraCritica:   obtenerMasFrecuente(acc.Horas),
		ProductosPico: top3Keys(acc.Productos),
	})
}

	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].Tickets > resultado[j].Tickets
	})

	modeloClientes = ModeloClientes{
		Segmentos: resultado,
	}

	return guardarJSON("modelos/clientes.model", modeloClientes)
}

func generarModeloClientesPorPeriodo(periodo string, modo string) (RespuestaClientes, error) {
	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		return RespuestaClientes{}, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return RespuestaClientes{}, err
	}

	type Acumulador struct {
		Tickets      map[string]bool
		Horas       map[string]int
		Productos   map[string]int
		MetodosPago map[string]int
		Genero      string
		TipoCliente string
		City        string
		TotalVentas float64
	}

	clientesUnicosGlobal := make(map[string]bool)
	segmentos := make(map[string]*Acumulador)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 14 {
			continue
		}

		invoiceID := strings.TrimSpace(row[0])
		fechaTexto := strings.TrimSpace(row[1])
		horaCompleta := strings.TrimSpace(row[2])
		city := strings.TrimSpace(row[3])
		producto := strings.ToLower(strings.TrimSpace(row[4]))
		totalTexto := strings.TrimSpace(row[10])
		metodoPago := strings.TrimSpace(row[11])
		genero := strings.TrimSpace(row[12])
		tipoCliente := strings.TrimSpace(row[13])

		if invoiceID == "" || fechaTexto == "" || producto == "" || genero == "" || tipoCliente == "" {
			continue
		}

		fecha, ok := parseFechaCSV(fechaTexto)
		if !ok {
			continue
		}

		mesRegistro := fecha.Format("2006-01")

		if periodo != "general" && mesRegistro != periodo {
			continue
		}

		clientesUnicosGlobal[invoiceID] = true

		totalTexto = strings.ReplaceAll(totalTexto, "$", "")
		totalTexto = strings.ReplaceAll(totalTexto, ",", "")

		totalFloat, err := strconv.ParseFloat(totalTexto, 64)
		if err != nil {
			totalFloat = 0
		}

		partesHora := strings.Split(horaCompleta, ":")
		hora := ""
		if len(partesHora) > 0 && partesHora[0] != "" {
			hora = partesHora[0] + ":00"
		}

		segmento := ""

if modo == "sucursales" {
	segmento = fmt.Sprintf(
		"%s | %s | %s",
		genero,
		tipoCliente,
		city,
	)
} else {
	segmento = fmt.Sprintf(
		"%s | %s",
		genero,
		tipoCliente,
	)
}

		if _, ok := segmentos[segmento]; !ok {
			segmentos[segmento] = &Acumulador{
				Tickets:      make(map[string]bool),
				Horas:       make(map[string]int),
				Productos:   make(map[string]int),
				MetodosPago: make(map[string]int),
				Genero:      genero,
				TipoCliente: tipoCliente,
				City:        city,
				TotalVentas: 0,
			}
		}

		segmentos[segmento].Tickets[invoiceID] = true
		segmentos[segmento].Productos[producto]++
		segmentos[segmento].TotalVentas += totalFloat

		if hora != "" {
			segmentos[segmento].Horas[hora]++
		}

		if metodoPago != "" {
			segmentos[segmento].MetodosPago[metodoPago]++
		}
	}

	var resultado []SegmentoCliente

	for nombre, acc := range segmentos {
		tickets := len(acc.Tickets)

		gastoPromedio := 0.0
		if tickets > 0 {
			gastoPromedio = acc.TotalVentas / float64(tickets)
		}

		resultado = append(resultado, SegmentoCliente{
			Segmento:      nombre,
			Tickets:       tickets,
			HoraPico:      obtenerMasFrecuente(acc.Horas),
			ProductoTop:   obtenerMasFrecuente(acc.Productos),
			MetodoPagoTop: obtenerMasFrecuente(acc.MetodosPago),
			Genero:        acc.Genero,
			TipoCliente:   acc.TipoCliente,
			City:          acc.City,
			TotalVentas:   acc.TotalVentas,
			GastoPromedio: gastoPromedio,
			HoraCritica:   obtenerMasFrecuente(acc.Horas),
			ProductosPico: top3Keys(acc.Productos),
		})
	}

	sort.Slice(resultado, func(i, j int) bool {
		return resultado[i].Tickets > resultado[j].Tickets
	})

	return RespuestaClientes{
	Segmentos:       resultado,
	ClientesUnicos: len(clientesUnicosGlobal),
}, nil
}

func cargarModeloClientes() {
	file, err := os.Open("modelos/clientes.model")
	if err != nil {
		return
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&modeloClientes)
}

func main() {
	fmt.Println("Servidor iniciado en http://localhost:8080")

	os.MkdirAll("modelos", os.ModePerm)
	os.MkdirAll("uploads", os.ModePerm)
	os.MkdirAll("static", os.ModePerm)

	cargarModeloRecomendaciones()
	cargarModeloKNN()
	cargarModeloHorarios()
	cargarModeloClientes()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"mensaje":"API funcionando"}`))
	})

	http.HandleFunc("/api/entrenar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := entrenarModeloRecomendaciones()
		if err != nil {
			responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":        true,
			"mensaje":   "Modelo Apriori de recomendaciones entrenado correctamente",
			"productos": len(modelo.Productos),
			"reglas":    len(modelo.Reglas),
			"tiendas":   len(modelo.Tiendas),
		})
	})

	http.HandleFunc("/api/entrenar-despensa", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := entrenarModeloKNN()
		if err != nil {
			responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"mensaje": "Modelo KNN de despensa entrenado correctamente",
			"casos":   len(modeloKNN.Casos),
		})
	})

	http.HandleFunc("/api/entrenar-horarios", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := entrenarModeloHorarios()
		if err != nil {
			responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"mensaje": "Modelo de horarios entrenado correctamente",
			"dia":    len(modeloHorarios.Dia),
			"semana": len(modeloHorarios.Semana),
			"mes":    len(modeloHorarios.Mes),
		})
	})

	http.HandleFunc("/api/entrenar-clientes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := entrenarModeloClientes()
		if err != nil {
			responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":        true,
			"mensaje":   "Modelo de clientes entrenado correctamente",
			"segmentos": len(modeloClientes.Segmentos),
		})
	})

	http.HandleFunc("/api/productos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if len(modelo.Productos) == 0 {
			err := entrenarModeloRecomendaciones()
			if err != nil {
				responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		json.NewEncoder(w).Encode(modelo.Productos)
	})

	http.HandleFunc("/api/productos-precios", func(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	sucursal := r.URL.Query().Get("sucursal")

	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type ProductoInfo struct {
		Producto  string  `json:"producto"`
		Precio    float64 `json:"precio"`
		Categoria string  `json:"categoria"`
	}

	productosMap := make(map[string]ProductoInfo)

	for i, row := range rows {

		if i == 0 {
			continue
		}

		if len(row) < 11 {
			continue
		}

		city := strings.TrimSpace(row[3])
		producto := strings.TrimSpace(row[4])
		categoria := strings.TrimSpace(row[5])

		precioTexto := strings.TrimSpace(row[7])

		precioTexto = strings.ReplaceAll(precioTexto, "$", "")
		precioTexto = strings.ReplaceAll(precioTexto, ",", "")

		precio, _ := strconv.ParseFloat(precioTexto, 64)

		if sucursal != "" && city != sucursal {
			continue
		}

		if producto == "" {
			continue
		}

		if _, ok := productosMap[producto]; !ok {

			productosMap[producto] = ProductoInfo{
				Producto:  producto,
				Precio:    precio,
				Categoria: categoria,
			}
		}
	}

	var lista []ProductoInfo

	for _, p := range productosMap {
		lista = append(lista, p)
	}

	sort.Slice(lista, func(i, j int) bool {
		return lista[i].Producto < lista[j].Producto
	})

	json.NewEncoder(w).Encode(lista)
})

	http.HandleFunc("/api/tiendas", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if len(modelo.Tiendas) == 0 {
			err := entrenarModeloRecomendaciones()
			if err != nil {
				responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		json.NewEncoder(w).Encode(modelo.Tiendas)
	})
	
http.HandleFunc("/api/productos-destacados", func(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	sucursal := r.URL.Query().Get("sucursal")

	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type ProductoInfo struct {
		Producto string `json:"producto"`
		Ventas   int    `json:"ventas"`
	}

	contador := make(map[string]int)

	for i, row := range rows {

		if i == 0 {
			continue
		}

		if len(row) < 14 {
			continue
		}

		city := strings.TrimSpace(row[3])
		producto := strings.TrimSpace(row[4])


		if producto == "" {
			continue
		}

		if sucursal != "" && city != sucursal {
			continue
		}

		contador[producto]++
	}

	resultado := obtenerTopProductos(contador, 5)

json.NewEncoder(w).Encode(resultado)
})

	http.HandleFunc("/api/recomendar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query().Get("carrito")
		if query == "" {
			responderErrorJSON(w, "Falta el parámetro carrito", http.StatusBadRequest)
			return
		}

		if len(modelo.Reglas) == 0 {
			err := entrenarModeloRecomendaciones()
			if err != nil {
				responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		carrito := strings.Split(query, ",")

		for i := range carrito {
			carrito[i] = strings.ToLower(strings.TrimSpace(carrito[i]))
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":              true,
			"carrito":         carrito,
			"recomendaciones": recomendar(carrito),
		})
	})

	http.HandleFunc("/api/recomendar-despensa", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		personas, _ := strconv.Atoi(r.URL.Query().Get("personas"))
		m2, _ := strconv.Atoi(r.URL.Query().Get("m2"))
		genero := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("genero")))

		if personas <= 0 || m2 <= 0 || genero == "" {
			responderErrorJSON(w, "Faltan parámetros válidos: personas, m2, genero", http.StatusBadRequest)
			return
		}

		if len(modeloKNN.Casos) == 0 {
			err := entrenarModeloKNN()
			if err != nil {
				responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		req := DespensaRequest{
			Personas: personas,
			M2:       m2,
			Genero:   genero,
		}

		resultado := recomendarDespensaKNN(req)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":              true,
			"entrada":         req,
			"recomendaciones": resultado,
		})
	})

	http.HandleFunc("/api/horarios-pico", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if len(modeloHorarios.Dia) == 0 && len(modeloHorarios.Semana) == 0 && len(modeloHorarios.Mes) == 0 {
		err := entrenarModeloHorarios()
		if err != nil {
			responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	periodo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("periodo")))

	modo := strings.TrimSpace(r.URL.Query().Get("modo"))

if modo == "" {
	modo = "general"
}

	var datos []HoraPico

	switch periodo {
	case "dia":
		datos = modeloHorarios.Dia
	case "semana":
		datos = modeloHorarios.Semana
	case "mes":
		datos = modeloHorarios.Mes
	default:
		datos = modeloHorarios.Dia
		periodo = "dia"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"periodo": periodo,
		"picos":   datos,
	})
})

	http.HandleFunc("/api/clientes-pico", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
	if periodo == "" {
		periodo = "general"
	}

	modo := strings.TrimSpace(r.URL.Query().Get("modo"))
	if modo == "" {
		modo = "general"
	}

	modeloTemp, err := generarModeloClientesPorPeriodo(periodo, modo)
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
	"ok":              true,
	"periodo":         periodo,
	"modo":            modo,
	"clientes_unicos": modeloTemp.ClientesUnicos,
	"segmentos":       modeloTemp.Segmentos,
})
})

	http.HandleFunc("/api/concurrencia-semanal", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := obtenerConcurrenciaSemanal()
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
})

http.HandleFunc("/api/periodos-disponibles", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	periodos, err := obtenerPeriodosDisponibles()
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":       true,
		"periodos": periodos,
	})
})

http.HandleFunc("/api/concurrencia-periodo", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tipo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tipo")))
	periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))

	if tipo != "mes" && tipo != "semana" {
		responderErrorJSON(w, "Tipo inválido. Usa mes o semana.", http.StatusBadRequest)
		return
	}

	if periodo == "" {
		responderErrorJSON(w, "Falta el parámetro periodo.", http.StatusBadRequest)
		return
	}

	sucursal := strings.TrimSpace(r.URL.Query().Get("sucursal"))
data, err := obtenerConcurrenciaPeriodo(tipo, periodo, sucursal)
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":      true,
		"tipo":    tipo,
		"periodo": periodo,
		"data":    data,
	})
})

http.HandleFunc("/api/sucursales-csv", func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	file, err := os.Open("uploads/dataset_walmart.csv")
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		responderErrorJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}

	set := make(map[string]bool)

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 4 {
			continue
		}

		city := strings.TrimSpace(row[3])
		if city != "" {
			set[city] = true
		}
	}

	var sucursales []string
	for s := range set {
		sucursales = append(sucursales, s)
	}

	sort.Strings(sucursales)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":         true,
		"sucursales": sucursales,
	})
})

	http.ListenAndServe(":8080", nil)
}