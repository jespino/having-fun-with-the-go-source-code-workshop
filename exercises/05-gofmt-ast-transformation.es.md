# Ejercicio 5: Modificacion de gofmt - Indentacion y Transformacion del AST

> 📖 **¿Quieres aprender más?** Lee [The Parser](https://internals-for-interns.com/es/posts/the-go-parser/) en Internals for Interns para profundizar en cómo Go construye y trabaja con los Árboles de Sintaxis Abstracta (AST).

En este ejercicio, modificaras la herramienta de formateo de Go `gofmt` para que use 4 espacios en lugar de tabulaciones, y luego anadiras una transformacion personalizada del AST para reemplazar automaticamente la palabra "hello" por "helo" en cadenas de texto y comentarios. Esto te ensenara como funciona el formateador de Go, como los modos del printer controlan la indentacion y como anadir transformaciones personalizadas al pipeline de procesamiento del AST.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender como gofmt controla la indentacion y los modos del printer
- Aprender a modificar el comportamiento de formateo en gofmt y el paquete go/format
- Entender como gofmt procesa el codigo fuente de Go mediante la manipulacion del AST
- Saber como modificar cadenas de texto y comentarios en el AST
- Explorar la estructura del AST (Abstract Syntax Tree) de Go
- Crear transformaciones de codigo fuente personalizadas

## Contexto: Como Funciona gofmt

gofmt opera a traves de estas etapas:

1. **Parsear** → Convertir el codigo fuente a AST (Abstract Syntax Tree)
2. **Transformar** → Aplicar reglas de formateo al AST
3. **Imprimir** → Convertir el AST modificado de vuelta a codigo fuente formateado con la indentacion especifica

El comportamiento de la indentacion esta controlado por dos constantes clave:

- **`tabWidth`** → Ancho de la indentacion (por defecto: 8)
- **`printerMode`** → Flags que controlan el comportamiento del espaciado:
  - `printer.UseSpaces` → Usar espacios para el relleno
  - `printer.TabIndent` → Usar tabulaciones para la indentacion
  - `printerNormalizeNumbers` → Normalizar literales numericos

### Estructura del AST

Go representa el codigo fuente como un arbol de nodos. Vamos a usar estos dos nodos:

- **`*ast.BasicLit`** → Cadenas de texto, numeros, etc.
- **`*ast.Comment`** → Comentarios en el codigo fuente

## Paso 1: Navegar al Codigo Fuente de gofmt

```bash
cd go/src/cmd/gofmt
ls -la
```

Archivos clave:

- **`gofmt.go`** → Logica principal del programa y procesamiento de archivos
- **`simplify.go`** → Transformaciones de simplificacion del AST

## Paso 2: Cambiar la Indentacion a 4 Espacios

Antes de anadir transformaciones personalizadas, cambiemos gofmt para que use 4 espacios en lugar de tabulaciones para la indentacion.

### Modificar gofmt.go

**Edita `go/src/cmd/gofmt/gofmt.go`:**

Busca las constantes alrededor de la linea 50 (busca el comentario "Keep these in sync with go/format/format.go"):

```go
const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers
```

Cambia a:

```go
const (
	tabWidth    = 4
	printerMode = printer.UseSpaces | printerNormalizeNumbers
```

**Que cambio:**

- **`tabWidth`**: Cambiado de `8` a `4` (4 espacios por nivel de indentacion)
- **`printerMode`**: Eliminado el flag `printer.TabIndent` (esto elimina los caracteres de tabulacion y usa solo espacios)

### Modificar el Paquete go/format

El paquete `go/format` tambien necesita actualizarse para mantener el comportamiento consistente.

**Edita `go/src/go/format/format.go`:**

Busca las constantes alrededor de la linea 29 (mismo comentario que arriba):

```go
const (
	tabWidth    = 8
	printerMode = printer.UseSpaces | printer.TabIndent | printerNormalizeNumbers
```

Cambia a:

```go
const (
	tabWidth    = 4
	printerMode = printer.UseSpaces | printerNormalizeNumbers
```

### Entendiendo los Cambios

- **`tabWidth = 4`**: Cada nivel de indentacion usa 4 espacios
- **Eliminar `TabIndent`**: Sin este flag, el printer usa solo espacios (sin caracteres de tabulacion)
- **`UseSpaces`**: Asegura que se usen espacios para el relleno y la alineacion
- **Ambos archivos deben coincidir**: gofmt y go/format deben usar la misma configuracion para ser consistentes

## Paso 3: Recompilar y Probar la Indentacion

```bash
cd ../../../  # back to go/src
./make.bash
```

Crea un archivo de prueba `indent_test.go`:

```go
package main

import "fmt"

func main() {
	if true {
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
	}
}
```

Prueba la nueva indentacion:

```bash
cd ..  # to go/ directory
./bin/gofmt indent_test.go
```

Salida esperada (observa los 4 espacios en cada nivel):

```go
package main

import "fmt"

func main() {
    if true {
        for i := 0; i < 10; i++ {
            fmt.Println(i)
        }
    }
}
```

Cada nivel de indentacion ahora usa 4 espacios en lugar de tabulaciones.

## Paso 4: Anadir la Transformacion Hello→Helo

**Edita `gofmt.go`:**

Anade esta funcion de transformacion alrededor de la linea 76 (despues de la funcion `usage()`):

```go
// transformHelloToHelo walks the AST and replaces "hello" with "helo"
// in string literals and comments.
func transformHelloToHelo(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BasicLit:
			// Handle string literals
			if node.Kind == token.STRING {
				if strings.Contains(node.Value, "hello") {
					node.Value = strings.ReplaceAll(node.Value, "hello", "helo")
				}
			}
		case *ast.Comment:
			// Handle comments
			if strings.Contains(node.Text, "hello") {
				node.Text = strings.ReplaceAll(node.Text, "hello", "helo")
			}
		}
		return true // continue traversing
	})
}
```

### Entendiendo el Codigo

- **`ast.Inspect()`** - Recorre todos los nodos del AST
- **`*ast.BasicLit`** - Coincide con literales de cadena de texto
- **`node.Kind == token.STRING`** - Verifica que sea una cadena de texto (no un numero)
- **`*ast.Comment`** - Coincide con comentarios
- **`strings.ReplaceAll()`** - Realiza el reemplazo

## Paso 5: Integrar la Transformacion

**Todavia en `gofmt.go`:**

Busca la funcion `processFile` alrededor de la linea 238. Busca el bloque `if *simplifyAST` alrededor de la linea 263:

```go
	if *simplifyAST {
		simplify(file)
	}
```

Anade nuestra transformacion justo despues:

```go
	if *simplifyAST {
		simplify(file)
	}

	// Apply our custom hello→helo transformation
	transformHelloToHelo(file)
```

## Paso 6: Recompilar gofmt

```bash
cd ../../../  # back to go/src
./make.bash
```

## Paso 7: Probar Ambas Modificaciones Juntas

Crea un archivo `hello_test.go`:

```go
package main

import "fmt"

func main() {
    // Say hello to everyone
    message := "hello world"
    greeting := "Say hello!"

    /* This is a hello comment block */
    fmt.Println(message)
    fmt.Println(greeting)

    // Another hello comment
    fmt.Printf("hello %s\n", "Go")
}
```

```bash
../go/bin/gofmt hello_test.go
```

Salida esperada (observa tanto la indentacion de 4 espacios COMO la transformacion hello→helo):

```go
package main

import "fmt"

func main() {
    // Say helo to everyone
    message := "helo world"
    greeting := "Say helo!"

    /* This is a helo comment block */
    fmt.Println(message)
    fmt.Println(greeting)

    // Another helo comment
    fmt.Printf("helo %s\n", "Go")
}
```

Se aplicaron dos cambios:

1. Todas las instancias de "hello" se reemplazaron por "helo"
2. La indentacion usa 4 espacios en lugar de tabulaciones

## Paso 8: Probar el Formateo In-Place

```bash
# Format and overwrite the file
../go/bin/gofmt -w hello_test.go

# Verify the changes
cat hello_test.go
```

El archivo ahora esta permanentemente transformado con "helo" en lugar de "hello" y usando indentacion de 4 espacios.

## Que Hicimos

1. **Modificamos la Configuracion del Printer**: Cambiamos tabWidth y printerMode para usar 4 espacios
2. **Sincronizamos Dos Paquetes**: Actualizamos tanto gofmt como go/format para mantener la consistencia
3. **Anadimos un Visitante del AST**: Creamos una funcion para recorrer y modificar los nodos del AST
4. **Coincidencia de Patrones**: Identificamos cadenas de texto y comentarios
5. **Reemplazo de Texto**: Modificamos los valores de los nodos para reemplazar "hello" por "helo"
6. **Integracion**: Llamamos a la transformacion durante el procesamiento de gofmt
7. **Pruebas**: Verificamos los cambios tanto de indentacion como de transformacion

## Lo que Aprendimos

- **Configuracion del Printer**: Como gofmt controla la indentacion mediante tabWidth y printerMode
- **Consistencia entre Paquetes**: Por que gofmt y go/format deben mantenerse sincronizados
- **Manipulacion del AST**: Como recorrer y modificar el Abstract Syntax Tree de Go
- **Modificacion de Herramientas**: Como extender herramientas existentes de Go con multiples cambios
- **Transformacion de Codigo**: Implementar cambios sistematicos en el codigo fuente
- **Proceso de Compilacion**: Recompilar componentes de la cadena de herramientas de Go
- **Pruebas**: Verificar el comportamiento de herramientas personalizadas

## Ideas de Extension

Prueba estas modificaciones adicionales:

1. Anadir un flag de linea de comandos para activar/desactivar la transformacion
2. Soportar multiples reemplazos de palabras (hello→helo, world→universe)
3. Anadir opcion de sensibilidad a mayusculas/minusculas
4. Reemplazar solo palabras completas (no subcadenas dentro de palabras)
5. Hacer tabWidth configurable mediante un flag de linea de comandos
6. Anadir opcion para alternar entre tabulaciones y espacios

Ejemplo de adicion de flag:
```go
var replaceHello = flag.Bool("helo", false, "replace hello with helo")

// In processFile():
if *replaceHello {
    transformHelloToHelo(file)
}
```

## Limpieza

Para restaurar el gofmt original:

```bash
cd go/src/cmd/gofmt
git checkout gofmt.go
cd ../go/format
git checkout format.go
cd ../../../src
./make.bash
```

## Resumen

Has modificado gofmt con exito de dos formas poderosas.

```
Indentacion:     tabulaciones (ancho 8) → 4 espacios
Transformacion:  "hello world"  → "helo world"
                 // Say hello    → // Say helo

Cambios:  tabWidth=4 + eliminar flag TabIndent
         + ast.Inspect() → coincidencia de patrones → reemplazar texto
```

Ahora entiendes como herramientas como `gofmt`, `goimports` y `go fix` funcionan tanto a nivel del printer como del AST.

---

*Continua con el [Ejercicio 6](06-ssa-power-of-two-detector.es.md) o vuelve al [taller principal](../README.md)*
