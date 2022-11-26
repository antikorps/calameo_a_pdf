import { GoDescargar } from "../wailsjs/go/main/App";
import { GoRutaDestino } from "../wailsjs/go/main/App";

// Selectores
const $botonGenerar = document.querySelector("#generar");
const $inputDescargas = document.querySelector("#numero-descargas");
const $textareaUrls = document.querySelector("textarea");
const $botonCarpeta = document.querySelector("#carpeta-destino");
const $inputEspera = document.querySelector("#tiempo-espera");
const $contenedorResumen = document.querySelector("#resumen");

// Globales
var rutaDestino = "";

// Escuchadores
$botonCarpeta.addEventListener("click", jsCarpetaDestino);
$botonGenerar.addEventListener("click", jsProcesar);

// Funciones
async function jsCarpetaDestino(evento) {
  evento.preventDefault();
  try {
    rutaDestino = await GoRutaDestino();
  } catch (error) {
    rutaDestino = "";
    alert(error);
  }
}

function obtenerUrls() {
  const contenidoLineas = $textareaUrls.value.split("\n");
  let urls = [];
  for (let linea of contenidoLineas) {
    if (linea === "") {
      continue;
    }
    if (!linea.includes("/read/")) {
      alert(
        `La URL ${linea} no se va a procesar porque no se ha encontrado /read/ y parece incorrecta`,
      );
      continue;
    }
    urls.push(linea.trim());
  }
  return urls;
}

async function jsProcesar(evento) {
  evento.preventDefault();

  $botonGenerar.setAttribute("aria-busy", true);
  $contenedorResumen.innerHTML = "";

  if (rutaDestino === "") {
    alert("Debe escogerse una carpeta de destino");
    $botonGenerar.removeAttribute("aria-busy");
    return;
  }

  const descargasSimultaneas = parseInt($inputDescargas.value);
  const tiempoEspera = parseInt($inputEspera.value);
  const urls = obtenerUrls();

  for (const url of urls) {
    let bloque = `
    <article>
      <p>Procesando: <a href="${url}" target="_blank">${url}</a></p>
        <progress indeterminate="true"></progress>
      <p class="resultado-descarga"></p>
    </article>
    `;
    $contenedorResumen.insertAdjacentHTML("beforeend", bloque);

    let $resultado = document.querySelectorAll(
      ".resultado-descarga:last-of-type ",
    )
    $resultado = $resultado[$resultado.length - 1];

    let $progreso = document.querySelectorAll("progress:last-of-type");
    $progreso = $progreso[$progreso.length - 1];

    let $articulo = document.querySelectorAll("article:last-of-type");
    $articulo = $articulo[$articulo.length - 1];
    try {
      let respuesta = await GoDescargar(
        url,
        rutaDestino,
        descargasSimultaneas,
        tiempoEspera,
      );
      $resultado.innerHTML = respuesta;
      $articulo.style.border = "2px solid lightgreen";
    } catch (error) {
      $resultado.innerHTML = error;
      $articulo.style.border = "2px solid tomato";
    }
    $progreso.remove();
  }

  $botonGenerar.removeAttribute("aria-busy");
}
