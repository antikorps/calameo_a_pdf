package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"
	"unicode"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Globales
var cliente = http.Client{
	Timeout: 8 * time.Second,
}
var wg sync.WaitGroup

// Structs
type CanalIntercambio struct {
	Error       bool
	Numero      int
	RutaArchivo string
}

type ManejadorDescarga struct {
	UrlBase                string
	ErrorValidacion        bool
	ErrorValidacionMensaje error
	ErrorCrearPDF          bool
	ErrorCrearPDFMensaje   error
	CarpetaDestino         string
	ImageSRC               string
	ColeccionImagenes      map[int]string
	DescargasSimultaneas   int
	Cliente                http.Client
	Nombre                 string
	Canal                  chan CanalIntercambio
	TiempoEspera           int
}

// Utilidades
func normalizarTexto(nombre string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	saneado, _, e := transform.String(t, nombre)
	regexNoAlfanumero := regexp.MustCompile(`[^a-zA-Z\d:]`)
	saneado = regexNoAlfanumero.ReplaceAllString(saneado, "")
	if e != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return saneado
}

func incorporarCabeceras(peticion *http.Request) {
	peticion.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:107.0) Gecko/20100101 Firefox/107.0")
	peticion.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	peticion.Header.Set("Accept-Language", "en-US,en;q=0.5")
	peticion.Header.Set("Connection", "keep-alive")
	peticion.Header.Set("Upgrade-Insecure-Requests", "1")
	peticion.Header.Set("Sec-Fetch-Dest", "document")
	peticion.Header.Set("Sec-Fetch-Mode", "navigate")
	peticion.Header.Set("Sec-Fetch-Site", "none")
	peticion.Header.Set("Sec-Fetch-User", "?1")
	peticion.Header.Set("Pragma", "no-cache")
	peticion.Header.Set("Cache-Control", "no-cache")
}

// Métodos
func NuevoManejador(url, carpetaDestino string, numeroDescargas, tiempoEspera int) ManejadorDescarga {
	return ManejadorDescarga{
		Cliente:              cliente,
		UrlBase:              url,
		CarpetaDestino:       carpetaDestino,
		DescargasSimultaneas: numeroDescargas,
		TiempoEspera:         tiempoEspera,
	}
}

func (m *ManejadorDescarga) Validar() {
	peticion, peticionError := http.NewRequest("GET", m.UrlBase, nil)
	if peticionError != nil {
		m.ErrorValidacion = true
		m.ErrorValidacionMensaje = peticionError
		return
	}
	incorporarCabeceras(peticion)
	respuesta, respuestaError := m.Cliente.Do(peticion)
	if respuestaError != nil {
		m.ErrorValidacion = true
		m.ErrorValidacionMensaje = respuestaError
		return
	}
	defer respuesta.Body.Close()
	if respuesta.StatusCode > 200 || respuesta.StatusCode > 299 {
		m.ErrorValidacion = true
		m.ErrorValidacionMensaje = errors.New("la petición a la URL indicada a recibido un status code incorrecto " + respuesta.Status)
	}
	doc, docError := io.ReadAll(respuesta.Body)
	if docError != nil {
		m.ErrorValidacion = true
		m.ErrorValidacionMensaje = docError
	}

	html := string(doc)
	expRegMin := regexp.MustCompile(`\n`)
	html = expRegMin.ReplaceAllString(html, "")

	expRegImagen := regexp.MustCompile(`.*<link rel="image_src" href="(.*?)/p[0-9].*`)
	imageSRC := expRegImagen.ReplaceAllString(html, "$1")

	m.ImageSRC = imageSRC + "/p"

	regexTitulo := regexp.MustCompile(`.*<title>(.*?)</title>.*`)
	titulo := regexTitulo.ReplaceAllString(html, "$1")
	m.Nombre = normalizarTexto(titulo)

}

func (m *ManejadorDescarga) DescargarArchivo(numeroPagina int) {
	defer wg.Done()
	urlPagina := fmt.Sprint(m.ImageSRC, numeroPagina, ".jpg")
	peticion, peticionError := http.NewRequest("GET", urlPagina, nil)
	if peticionError != nil {
		m.Canal <- CanalIntercambio{
			Error: true,
		}
		return
	}
	incorporarCabeceras(peticion)
	respuesta, respuestaError := m.Cliente.Do(peticion)
	if respuestaError != nil {
		m.Canal <- CanalIntercambio{
			Error: true,
		}
		return
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode < 200 || respuesta.StatusCode > 299 {
		m.Canal <- CanalIntercambio{
			Error: true,
		}
		return
	}

	nombreArchivo := fmt.Sprint(m.Nombre, "_", numeroPagina, ".jpg")
	rutaArchivo := filepath.Join(m.CarpetaDestino, nombreArchivo)
	archivoDestino, archivoDestinoError := os.Create(rutaArchivo)
	if archivoDestinoError != nil {
		m.Canal <- CanalIntercambio{
			Error: true,
		}
		return
	}

	_, descargaError := io.Copy(archivoDestino, respuesta.Body)
	if descargaError != nil {
		m.Canal <- CanalIntercambio{
			Error: true,
		}
		return
	}

	m.Canal <- CanalIntercambio{
		Error:       false,
		Numero:      numeroPagina,
		RutaArchivo: rutaArchivo,
	}
}

func (m *ManejadorDescarga) Descargar() {
	if m.ErrorValidacion {
		return
	}

	m.ColeccionImagenes = make(map[int]string)

	numeroPagina := 1

	continuarBucle := true

	for continuarBucle {
		canal := make(chan CanalIntercambio)
		m.Canal = canal

		for i := 0; i < m.DescargasSimultaneas; i++ {
			wg.Add(1)
			go m.DescargarArchivo(numeroPagina)
			numeroPagina++
		}

		go func() {
			wg.Wait()
			close(m.Canal)
		}()

		time.Sleep(time.Duration(m.TiempoEspera) * time.Second)

		for c := range m.Canal {
			if c.Error {
				continuarBucle = false
				continue
			}
			m.ColeccionImagenes[c.Numero] = c.RutaArchivo
		}
	}
}

func (m *ManejadorDescarga) CrearPDF() {
	// 	Ordena todas las páginas (desorden provocado por gorutinas)
	var numerosPaginas []int
	for c := range m.ColeccionImagenes {
		numerosPaginas = append(numerosPaginas, c)
	}
	sort.Ints(numerosPaginas)
	var paginasOrdenadas []string
	for _, v := range numerosPaginas {
		paginasOrdenadas = append(paginasOrdenadas, m.ColeccionImagenes[v])
	}
	// ImportImagesFile añade páginas si el archivo existe, eliminar para evitar solapamiento
	rutaArchivo := filepath.Join(m.CarpetaDestino, fmt.Sprint(m.Nombre, ".pdf"))
	os.Remove(rutaArchivo)
	errorPDF := api.ImportImagesFile(paginasOrdenadas, rutaArchivo, nil, nil)

	if errorPDF != nil {
		m.ErrorCrearPDF = true
		m.ErrorCrearPDFMensaje = errorPDF
	}
}

func (m *ManejadorDescarga) BorrarArchivos() {
	for _, v := range m.ColeccionImagenes {
		os.Remove(v)
	}
}

func (m *ManejadorDescarga) GenerarRespuesta() (string, error) {

	if m.ErrorValidacion {
		return "", m.ErrorValidacionMensaje
	}

	if m.ErrorCrearPDF {
		return "", m.ErrorCrearPDFMensaje
	}

	archivoDefinitivoNombre := fmt.Sprint(m.Nombre, ".pdf")
	archivoDefinitivoRuta := filepath.Join(m.CarpetaDestino, archivoDefinitivoNombre)
	respuesta := fmt.Sprint("PDF generado con ", len(m.ColeccionImagenes), " páginas. Comprobar que coincide con el total esperado. Archivo disponible en: ", archivoDefinitivoRuta)

	return respuesta, nil

}
