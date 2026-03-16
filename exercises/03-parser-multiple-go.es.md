# Ejercicio 3: Multiples Palabras Clave "go" - Mejora del Parser

> 📖 **¿Quieres saber más?** Lee [The Parser](https://internals-for-interns.com/es/posts/the-go-parser/) en Internals for Interns para una explicación detallada de cómo el parser de Go construye Árboles de Sintaxis Abstracta.

En este ejercicio, modificaras el parser de Go para aceptar multiples palabras clave "go" consecutivas al iniciar goroutines. Esto te ensenara como mejorar la logica del parser para manejar patrones de sintaxis repetitivos manteniendo el mismo comportamiento semantico.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender la estructura del parser de Go y el consumo de tokens
- Saber como modificar la logica del parser para extensiones de sintaxis
- Probar modificaciones del parser con codigo funcional

## Paso 1: Navegar al Parser

```bash
cd go/src/cmd/compile/internal/syntax
```

### Entender la Logica Actual del Parser

Examinemos como el parser maneja actualmente la sentencia "go" en `parser.go`. Mira alrededor de la linea 2675:

```go
// go/src/cmd/compile/internal/syntax/parser.go:2673-2676
...
return s

case _Go, _Defer:
    return p.callStmt()
...
```

El parser reconoce el token `_Go` e inmediatamente llama a `p.callStmt()` para manejar la creacion de la goroutine.

Encuentra el metodo `callStmt()` en `parser.go` en la linea 977. Aqui es donde anadiremos nuestra logica de multiples "go":

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

La linea clave es `s.Tok = p.tok` que captura si es una sentencia "defer" o "go", seguida de `p.next()` que consume el token.

## Paso 2: Anadir Soporte para Multiples "go"

Necesitamos modificar el metodo `callStmt()` para consumir multiples tokens "go" consecutivos manteniendo el mismo significado semantico.

**Edita `parser.go`:**

Encuentra la linea 985 donde se llama a `p.next()` y anade nuestra logica de multiples "go" justo despues:

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

### Entendiendo el Cambio en el Codigo

- **`if s.Tok == _Go`**: Solo aplica la logica de multiples palabras clave a sentencias "go" (no a "defer")
- **`for p.tok == _Go`**: Sigue consumiendo tokens "go" mientras aparezcan consecutivamente
- **`p.next()`**: Avanza mas alla de cada token "go" adicional
- **Preservacion**: `s.Tok` sigue siendo `_Go`, por lo que el significado semantico no cambia

## Paso 3: Recompilar el Compilador

Ahora recompilemos la toolchain de Go con nuestros cambios:

```bash
cd ../../../  # back to go/src
./make.bash
```

Si hay errores de compilacion, revisa tus cambios y corrígelos.

## Paso 4: Probar Multiples Palabras Clave "go"

Crea un programa de prueba para verificar que nuestra sintaxis de multiples "go" funciona:

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

Deberias ver una salida como esta:

```
Testing multiple go keywords...
Hello from single go!
Hello from double go!
Hello from triple go!
Hello from quadruple go!
All done!
```

## Paso 5: Ejecutar los Tests del Parser

Asegurate de que no hemos roto el parser:

```bash
cd /path/to/workshop/go/src
../bin/go test cmd/compile/internal/syntax -short
```

## Lo que Hicimos

1. **Mejora del Parser**: Modificamos `callStmt()` para manejar multiples tokens "go" consecutivos
2. **Consumo de Tokens**: Anadimos un bucle para consumir tokens "go" adicionales despues del primero
3. **Preservacion Semantica**: Multiples palabras clave "go" siguen creando exactamente una goroutine
4. **Cambio Dirigido**: Solo afecta a sentencias "go", no a sentencias "defer"

## Lo que Aprendimos

- **Logica del Parser**: Como Go procesa secuencias de tokens para convertirlas en sentencias
- **Consumo de Tokens**: Tecnicas para consumir multiples tokens del mismo tipo
- **Testing del Parser**: Validar cambios del parser con casos de prueba diversos

## Ideas de Extension

Prueba estas modificaciones adicionales:

1. Anadir soporte similar para "defer defer defer" (mas desafiante)
2. Anadir un limite maximo (por ejemplo, maximo 5 palabras clave "go" consecutivas)
3. Registrar cuantas palabras clave "go" se usaron para depuracion
4. Hacer que las multiples palabras clave afecten la prioridad de la goroutine

## Siguientes Pasos

Has mejorado exitosamente el parser de Go para manejar patrones de sintaxis repetitivos.

En el [Ejercicio 4: Parametros de Inlining del Compilador](./04-compiler-inlining-parameters.md), cambiaremos el enfoque para explorar como funciona la optimizacion del compilador de Go, aprendiendo a ajustar los parametros de inlining para controlar el tamano del binario.

## Limpieza

Para restaurar el codigo fuente original de Go:

```bash
cd /path/to/workshop/go/src/cmd/compile/internal/syntax
git checkout parser.go
cd ../../../
./make.bash  # Recompilar con el codigo original
```

## Resumen

Multiples palabras clave "go" ahora funcionan para iniciar goroutines:

```go
// Todas son equivalentes y crean exactamente una goroutine:
go myFunction()
go go myFunction()
go go go myFunction()
go go go go myFunction()

// El parser consume todos los tokens "go" consecutivos
// pero el comportamiento semantico sigue siendo el mismo!
```

Este ejercicio demostro como las modificaciones a nivel de parser pueden anadir azucar sintactico expresivo preservando la semantica subyacente del lenguaje.

---

*Continua al [Ejercicio 4](04-compiler-inlining-parameters.md) o vuelve al [taller principal](../README.md)*
