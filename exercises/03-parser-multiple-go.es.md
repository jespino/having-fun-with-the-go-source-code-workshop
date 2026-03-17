# Ejercicio 3: Múltiples Palabras Clave "go" - Mejora del Parser

> 📖 **¿Quieres saber más?** Lee [The Parser](https://internals-for-interns.com/es/posts/the-go-parser/) en Internals for Interns para una explicación detallada de cómo el parser de Go construye Árboles de Sintaxis Abstracta.

En este ejercicio, modificarás el parser de Go para aceptar múltiples palabras clave "go" consecutivas al iniciar goroutines. Esto te enseñará cómo mejorar la lógica del parser para manejar patrones de sintaxis repetitivos manteniendo el mismo comportamiento semántico.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Entender la estructura del parser de Go y el consumo de tokens
- Saber cómo modificar la lógica del parser para extensiones de sintaxis
- Probar modificaciones del parser con código funcional

## Introducción: ¿Qué es un Parser?

El parser es la segunda fase del compilador, justo después del scanner. Mientras el scanner produce un flujo plano de tokens, el trabajo del parser es darle **estructura** a ese flujo construyendo un **Árbol de Sintaxis Abstracta (AST)** — un árbol que representa las relaciones jerárquicas en tu código.

Por ejemplo, una sentencia `go sayHello()` se convierte en un nodo del árbol de tipo `CallStmt` con `Tok: _Go` y un nodo hijo que representa la llamada a la función `sayHello()`. El parser sabe que después de ver un token `go`, debe seguir una expresión de llamada a función — esto es la gramática del lenguaje.

El parser de Go usa una técnica llamada **descenso recursivo**: tiene una función para cada regla gramatical (archivo, declaración, sentencia, expresión), y estas funciones se llaman entre sí de arriba hacia abajo. El punto de entrada `fileOrNil()` parsea la cláusula del paquete, luego los imports, luego las declaraciones. Cada declaración puede contener sentencias, y cada sentencia puede contener expresiones.

El parser consume tokens uno a uno usando `p.next()`, y comprueba el token actual con `p.tok`. El parser se encuentra en `go/src/cmd/compile/internal/syntax/parser.go`.

## Paso 1: Navegar al Parser

```bash
cd go/src/cmd/compile/internal/syntax
```

### Entender la Lógica Actual del Parser

Examinemos cómo el parser maneja actualmente la sentencia "go" en `parser.go`. Mira alrededor de la línea 2675:

```go
// go/src/cmd/compile/internal/syntax/parser.go:2673-2676
...
return s

case _Go, _Defer:
    return p.callStmt()
...
```

El parser reconoce el token `_Go` e inmediatamente llama a `p.callStmt()` para manejar la creación de la goroutine.

Encuentra el método `callStmt()` en `parser.go` en la línea 977. Aquí es donde añadiremos nuestra lógica de múltiples "go":

```go
// go/src/cmd/compile/internal/syntax/parser.go:976-985
// callStmt parses call-like statements that can be preceded by 'defer' and 'go'.
func (p *parser) callStmt() *CallStmt {
    if trace {
        defer p.trace("callStmt")()
    }

    s := new(CallStmt)
    s.pos = p.pos()
    s.Tok = p.tok // _Defer or _Go
    p.next()
    ...
}
```

La línea clave es `s.Tok = p.tok` que captura si es una sentencia "defer" o "go", seguida de `p.next()` que consume el token.

## Paso 2: Añadir Soporte para Múltiples "go"

Necesitamos modificar el método `callStmt()` para consumir múltiples tokens "go" consecutivos manteniendo el mismo significado semántico.

**Edita `parser.go`:**

Encuentra la línea 985 donde se llama a `p.next()` y añade nuestra lógica de múltiples "go" justo después:

```go
// go/src/cmd/compile/internal/syntax/parser.go:982-990
s := new(CallStmt)
s.pos = p.pos()
s.Tok = p.tok // _Defer or _Go
p.next()

// Allow multiple consecutive "go" keywords (go go go ...)
if s.Tok == _Go {
    for p.tok == _Go {
        p.next()
    }
}

...
```

### Entendiendo el Cambio en el Código

- **`if s.Tok == _Go`**: Solo aplica la lógica de múltiples palabras clave a sentencias "go" (no a "defer")
- **`for p.tok == _Go`**: Sigue consumiendo tokens "go" mientras aparezcan consecutivamente
- **`p.next()`**: Avanza más allá de cada token "go" adicional
- **Preservación**: `s.Tok` sigue siendo `_Go`, por lo que el significado semántico no cambia

## Paso 3: Recompilar el Compilador

Ahora recompilemos la toolchain de Go con nuestros cambios:

```bash
cd ../../../  # back to go/src
./make.bash
```

Si hay errores de compilación, revisa tus cambios y corrígelos.

## Paso 4: Probar Múltiples Palabras Clave "go"

Crea un programa de prueba para verificar que nuestra sintaxis de múltiples "go" funciona:

```bash
mkdir -p /tmp/multiple-go-test
cd /tmp/multiple-go-test
```

Crea un archivo test.go:

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    fmt.Printf("Hello from %s!\n", name)
}

func main() {
    fmt.Println("Testing multiple go keywords...")

    // Test regular single go
    go sayHello("single go")

    // Test double go
    go go sayHello("double go")

    // Test triple go
    go go go sayHello("triple go")

    // Test quadruple go
    go go go go sayHello("quadruple go")

    // Wait a bit to see output
    time.Sleep(100 * time.Millisecond)
    fmt.Println("All done!")
}
```

Ejecuta el programa de prueba con tu Go personalizado:

```bash
/path/to/workshop/go/bin/go run test.go
```

Deberías ver una salida como esta:

```
Testing multiple go keywords...
Hello from single go!
Hello from double go!
Hello from triple go!
Hello from quadruple go!
All done!
```

## Paso 5: Ejecutar los Tests del Parser

Asegúrate de que no hemos roto el parser:

```bash
cd /path/to/workshop/go/src
../bin/go test cmd/compile/internal/syntax -short
```

## Lo que Hicimos

1. **Mejora del Parser**: Modificamos `callStmt()` para manejar múltiples tokens "go" consecutivos
2. **Consumo de Tokens**: Añadimos un bucle para consumir tokens "go" adicionales después del primero
3. **Preservación Semántica**: Múltiples palabras clave "go" siguen creando exactamente una goroutine
4. **Cambio Dirigido**: Solo afecta a sentencias "go", no a sentencias "defer"

## Lo que Aprendimos

- **Lógica del Parser**: Cómo Go procesa secuencias de tokens para convertirlas en sentencias
- **Consumo de Tokens**: Técnicas para consumir múltiples tokens del mismo tipo
- **Testing del Parser**: Validar cambios del parser con casos de prueba diversos

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir soporte similar para "defer defer defer" (más desafiante)
2. Añadir un límite máximo (por ejemplo, máximo 5 palabras clave "go" consecutivas)
3. Registrar cuántas palabras clave "go" se usaron para depuración
4. Hacer que las múltiples palabras clave afecten la prioridad de la goroutine

## Siguientes Pasos

Has mejorado exitosamente el parser de Go para manejar patrones de sintaxis repetitivos.

En el [Ejercicio 4: Parámetros de Inlining del Compilador](./04-compiler-inlining-parameters.md), cambiaremos el enfoque para explorar cómo funciona la optimización del compilador de Go, aprendiendo a ajustar los parámetros de inlining para controlar el tamaño del binario.

## Limpieza

Para restaurar el código fuente original de Go:

```bash
cd /path/to/workshop/go/src/cmd/compile/internal/syntax
git checkout parser.go
cd ../../../
./make.bash  # Recompilar con el código original
```

## Resumen

Múltiples palabras clave "go" ahora funcionan para iniciar goroutines:

```go
// Todas son equivalentes y crean exactamente una goroutine:
go myFunction()
go go myFunction()
go go go myFunction()
go go go go myFunction()

// El parser consume todos los tokens "go" consecutivos
// ¡pero el comportamiento semántico sigue siendo el mismo!
```

Este ejercicio demostró cómo las modificaciones a nivel de parser pueden añadir azúcar sintáctico expresivo preservando la semántica subyacente del lenguaje.

---

*Continúa al [Ejercicio 4](04-compiler-inlining-parameters.md) o vuelve al [taller principal](../README.md)*
