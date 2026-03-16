# Ejercicio 5: Modificación de gofmt - Indentación y Transformación del AST

> 📖 **¿Quieres aprender más?** Lee [The Parser](https://internals-for-interns.com/es/posts/the-go-parser/) en Internals for Interns para profundizar en cómo Go construye y trabaja con los Árboles de Sintaxis Abstracta (AST).

En este ejercicio, modificarás la herramienta de formateo de Go `gofmt` para que use 4 espacios en lugar de tabulaciones, y luego añadirás una transformación personalizada del AST para reemplazar automáticamente la palabra "hello" por "helo" en cadenas de texto y comentarios. Esto te enseñará cómo funciona el formateador de Go, cómo los modos del printer controlan la indentación y cómo añadir transformaciones personalizadas al pipeline de procesamiento del AST.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Entender cómo gofmt controla la indentación y los modos del printer
- Aprender a modificar el comportamiento de formateo en gofmt y el paquete go/format
- Entender cómo gofmt procesa el código fuente de Go mediante la manipulación del AST
- Saber cómo modificar cadenas de texto y comentarios en el AST
- Explorar la estructura del AST (Abstract Syntax Tree) de Go
- Crear transformaciones de código fuente personalizadas

## Contexto: Cómo Funciona gofmt

gofmt opera a través de estas etapas:

1. **Parsear** → Convertir el código fuente a AST (Abstract Syntax Tree)
2. **Transformar** → Aplicar reglas de formateo al AST
3. **Imprimir** → Convertir el AST modificado de vuelta a código fuente formateado con la indentación específica

El comportamiento de la indentación está controlado por dos constantes clave:

- **`tabWidth`** → Ancho de la indentación (por defecto: 8)
- **`printerMode`** → Flags que controlan el comportamiento del espaciado:
  - `printer.UseSpaces` → Usar espacios para el relleno
  - `printer.TabIndent` → Usar tabulaciones para la indentación
  - `printerNormalizeNumbers` → Normalizar literales numéricos

### Estructura del AST

Go representa el código fuente como un árbol de nodos. Vamos a usar estos dos nodos:

- **`*ast.BasicLit`** → Cadenas de texto, números, etc.
- **`*ast.Comment`** → Comentarios en el código fuente

## Paso 1: Navegar al Código Fuente de gofmt

```bash
cd go/src/cmd/gofmt
ls -la
```

Archivos clave:

- **`gofmt.go`** → Lógica principal del programa y procesamiento de archivos
- **`simplify.go`** → Transformaciones de simplificación del AST

## Paso 2: Cambiar la Indentación a 4 Espacios

Antes de añadir transformaciones personalizadas, cambiemos gofmt para que use 4 espacios en lugar de tabulaciones para la indentación.

### Modificar gofmt.go

**Edita `go/src/cmd/gofmt/gofmt.go`:**

Busca las constantes alrededor de la línea 50 (busca el comentario "Keep these in sync with go/format/format.go"):

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

**Qué cambió:**

- **`tabWidth`**: Cambiado de `8` a `4` (4 espacios por nivel de indentación)
- **`printerMode`**: Eliminado el flag `printer.TabIndent` (esto elimina los caracteres de tabulación y usa solo espacios)

### Modificar el Paquete go/format

El paquete `go/format` también necesita actualizarse para mantener el comportamiento consistente.

**Edita `go/src/go/format/format.go`:**

Busca las constantes alrededor de la línea 29 (mismo comentario que arriba):

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

- **`tabWidth = 4`**: Cada nivel de indentación usa 4 espacios
- **Eliminar `TabIndent`**: Sin este flag, el printer usa solo espacios (sin caracteres de tabulación)
- **`UseSpaces`**: Asegura que se usen espacios para el relleno y la alineación
- **Ambos archivos deben coincidir**: gofmt y go/format deben usar la misma configuración para ser consistentes

## Paso 3: Recompilar y Probar la Indentación

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

Prueba la nueva indentación:

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

Cada nivel de indentación ahora usa 4 espacios en lugar de tabulaciones.

## Paso 4: Añadir la Transformación Hello→Helo

**Edita `gofmt.go`:**

Añade esta función de transformación alrededor de la línea 76 (después de la función `usage()`):

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

### Entendiendo el Código

- **`ast.Inspect()`** - Recorre todos los nodos del AST
- **`*ast.BasicLit`** - Coincide con literales de cadena de texto
- **`node.Kind == token.STRING`** - Verifica que sea una cadena de texto (no un número)
- **`*ast.Comment`** - Coincide con comentarios
- **`strings.ReplaceAll()`** - Realiza el reemplazo

## Paso 5: Integrar la Transformación

**Todavía en `gofmt.go`:**

Busca la función `processFile` alrededor de la línea 238. Busca el bloque `if *simplifyAST` alrededor de la línea 263:

```go
	if *simplifyAST {
		simplify(file)
	}
```

Añade nuestra transformación justo después:

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

Salida esperada (observa tanto la indentación de 4 espacios COMO la transformación hello→helo):

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
2. La indentación usa 4 espacios en lugar de tabulaciones

## Paso 8: Probar el Formateo In-Place

```bash
# Format and overwrite the file
../go/bin/gofmt -w hello_test.go

# Verify the changes
cat hello_test.go
```

¡El archivo ahora está permanentemente transformado con "helo" en lugar de "hello" y usando indentación de 4 espacios!

## Qué Hicimos

1. **Modificamos la Configuración del Printer**: Cambiamos tabWidth y printerMode para usar 4 espacios
2. **Sincronizamos Dos Paquetes**: Actualizamos tanto gofmt como go/format para mantener la consistencia
3. **Añadimos un Visitante del AST**: Creamos una función para recorrer y modificar los nodos del AST
4. **Coincidencia de Patrones**: Identificamos cadenas de texto y comentarios
5. **Reemplazo de Texto**: Modificamos los valores de los nodos para reemplazar "hello" por "helo"
6. **Integración**: Llamamos a la transformación durante el procesamiento de gofmt
7. **Pruebas**: Verificamos los cambios tanto de indentación como de transformación

## Lo que Aprendimos

- **Configuración del Printer**: Cómo gofmt controla la indentación mediante tabWidth y printerMode
- **Consistencia entre Paquetes**: Por qué gofmt y go/format deben mantenerse sincronizados
- **Manipulación del AST**: Cómo recorrer y modificar el Abstract Syntax Tree de Go
- **Modificación de Herramientas**: Cómo extender herramientas existentes de Go con múltiples cambios
- **Transformación de Código**: Implementar cambios sistemáticos en el código fuente
- **Proceso de Compilación**: Recompilar componentes de la cadena de herramientas de Go
- **Pruebas**: Verificar el comportamiento de herramientas personalizadas

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir un flag de línea de comandos para activar/desactivar la transformación
2. Soportar múltiples reemplazos de palabras (hello→helo, world→universe)
3. Añadir opción de sensibilidad a mayúsculas/minúsculas
4. Reemplazar solo palabras completas (no subcadenas dentro de palabras)
5. Hacer tabWidth configurable mediante un flag de línea de comandos
6. Añadir opción para alternar entre tabulaciones y espacios

Ejemplo de adición de flag:
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

¡Has modificado gofmt con éxito de dos formas poderosas!

```
Indentación:     tabulaciones (ancho 8) → 4 espacios
Transformación:  "hello world"  → "helo world"
                 // Say hello    → // Say helo

Cambios:  tabWidth=4 + eliminar flag TabIndent
         + ast.Inspect() → coincidencia de patrones → reemplazar texto
```

Ahora entiendes cómo herramientas como `gofmt`, `goimports` y `go fix` funcionan tanto a nivel del printer como del AST.

---

*Continúa con el [Ejercicio 6](06-ssa-power-of-two-detector.es.md) o vuelve al [taller principal](../README.md)*
