# Ejercicio 2: Anadir el Operador Flecha "=>" para Goroutines

> 📖 **Quieres saber mas?** Lee [The Scanner](https://internals-for-interns.com/posts/the-go-lexer/) en Internals for Interns para una explicacion detallada de como funciona el lexer/scanner de Go.

En este ejercicio, anadiras un nuevo operador flecha "=>" a Go que funciona como una sintaxis alternativa para iniciar goroutines. Esto te ensenara como modificar el scanner de Go para reconocer nuevos operadores y mapearlos a funcionalidad existente.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender como el scanner de Go tokeniza operadores
- Saber como anadir nueva sintaxis de operadores a Go
- Modificar la logica de analisis lexico del scanner
- Probar tu modificacion del scanner con codigo funcional
- Extender con exito el vocabulario de operadores de Go

## Contexto: Como Funciona esta Modificacion del Scanner

Este ejercicio demuestra **modificaciones a nivel de scanner** para anadir nueva sintaxis de operadores a Go. Modificaremos la logica del scanner para reconocer una nueva secuencia de operador "=>" y mapearla a un token existente. Esto es lo que lograremos:

- **Mejora del Scanner**: Anadir reconocimiento para la secuencia del operador "=>"
- **Mapeo de Token**: Mapear "=>" al token `_Go` existente (igual que la palabra clave "go")
- **Sintaxis Alternativa**: Crear `=> myFunction()` como equivalente a `go myFunction()`
- **Impacto Minimo**: No se necesitan cambios en el parser ni en el compilador, solo logica del scanner

Este enfoque nos permite crear sintaxis alternativa elegante sin modificar las partes mas profundas del compilador.

## Paso 1: Navegar al Scanner

```bash
cd go/src/cmd/compile/internal/syntax
```

### Entender la Estructura del Scanner

Examinemos como el scanner maneja el operador "=" en `scanner.go`. Mira la linea 325:

```go
// go/src/cmd/compile/internal/syntax/scanner.go:325
case '=':
    s.nextch()
    if s.ch == '=' {
        s.nextch()
        s.op, s.prec = Eql, precCmp
        s.tok = _Operator
        break
    }
    s.tok = _Assign
```

El scanner primero consume el caracter "=" con `s.nextch()`, luego verifica si el siguiente caracter tambien es "=" (para el operador de comparacion `==`). Si no lo es, se queda con "=" (asignacion).

## Paso 2: Anadir la Logica del Operador Flecha

Necesitamos anadir logica para reconocer "=>" y tratarlo como el token `_Go`.

**Edita `scanner.go`:**

Encuentra el case "=" en la linea 325 y modificalo para que tambien verifique ">":

```go
// go/src/cmd/compile/internal/syntax/scanner.go:325
case '=':
    s.nextch()
    if s.ch == '=' {
        s.nextch()
        s.op, s.prec = Eql, precCmp
        s.tok = _Operator
        break
    }
    if s.ch == '>' {
        s.nextch()
        s.lit = "=>"
        s.tok = _Go
        break
    }
    s.tok = _Assign
```

### Entendiendo el Cambio en el Codigo

- **`if s.ch == '>'`**: Verifica si el siguiente caracter despues de "=" es ">" (recuerda que `s.nextch()` ya consumio el "=")
- **`s.nextch()`**: Consume el caracter ">" del lexer
- **`s.lit = "=>"`**: Establece el valor literal para depuracion y mensajes de error
- **`s.tok = _Go`**: Asigna el mismo token que la palabra clave "go"
- **`break`**: Sale del case para evitar caer en `_Assign`

## Paso 3: Recompilar el Compilador

Ahora recompilemos la toolchain de Go con nuestros cambios:

```bash
cd ../../../  # back to go/src
./make.bash
```

Si hay errores de compilacion, revisa tus cambios y corrígelos.

## Paso 4: Probar el Nuevo Operador Flecha

Crea un programa de prueba para verificar que nuestro nuevo operador "=>" funciona:

```bash
mkdir -p /tmp/arrow-test
cd /tmp/arrow-test
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
    fmt.Println("Testing => arrow operator...")

    // Test regular go keyword
    go sayHello("regular go")

    // Test our new => operator
    => sayHello("arrow operator")

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
Testing => arrow operator...
Hello from regular go!
Hello from arrow operator!
All done!
```

## Paso 5: Probar Operadores Mixtos de Go

Probemos escenarios mixtos usando tanto la palabra clave tradicional "go" como nuestro nuevo operador flecha "=>":

Crea un archivo mixed-test.go:

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(50 * time.Millisecond)
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup

    fmt.Println("Starting workers with mixed syntax...")

    // Mix of regular go and => operators
    for i := 1; i <= 4; i++ {
        wg.Add(1)
        if i%2 == 0 {
            go worker(i, &wg)  // Regular go
        } else {
            => worker(i, &wg)  // Arrow operator
        }
    }

    wg.Wait()
    fmt.Println("All workers completed!")
}
```

Ejecuta el programa de prueba mixto:

```bash
/path/to/workshop/go/bin/go run mixed-test.go
```

## Paso 6: Ejecutar los Tests del Scanner

Asegurate de que no hemos roto el scanner:

```bash
cd /path/to/workshop/go/src
../bin/go test cmd/compile/internal/syntax -short
```

## Lo que Hicimos

1. **Modificamos la Logica del Scanner**: Anadimos reconocimiento de "=>" al case existente de "="
2. **Reutilizamos un Token Existente**: Mapeamos "=>" al token `_Go` en lugar de crear uno nuevo
3. **Preservamos la Funcionalidad Existente**: Los operadores "=" y "==" siguen funcionando normalmente
4. **Impacto Minimo del Cambio**: No se necesitaron cambios en el parser ni en el IR

## Lo que Aprendimos

- **Logica del Scanner**: Como Go tokeniza secuencias de operadores
- **Reconocimiento de Operadores**: Anadir nuevos operadores mediante modificacion del scanner
- **Reutilizacion de Tokens**: Mapear nueva sintaxis a tokens existentes
- **Estrategia de Testing**: Validar cambios del scanner con codigo real
- **Proceso de Compilacion**: Recompilar Go con modificaciones del scanner

## Ideas de Extension

Prueba estas modificaciones adicionales:

1. Anadir ":>" como otra alternativa a "go"
2. Anadir "~>" para operaciones asincronas
3. Anadir ">>>" como operador de triple flecha
4. Hacer que el operador flecha funcione en diferentes contextos

## Siguientes Pasos

Has anadido exitosamente un nuevo operador al scanner de Go. Ahora entiendes como modificar el scanner para crear sintaxis alternativa para funcionalidad existente. Esta tecnica se puede aplicar para crear otros atajos de operadores y azucar sintactico en el lenguaje.

En el Ejercicio 3, tomaremos un enfoque diferente y exploraremos **modificaciones del parser**, aprendiendo como modificar el parser para manejar multiples tokens consecutivos.

## Limpieza

Para restaurar el codigo fuente original de Go:

```bash
cd /path/to/workshop/go/src/cmd/compile/internal/syntax
git checkout scanner.go
cd ../../../
./make.bash  # Recompilar con el codigo original
```

## Resumen

El operador flecha "=>" ahora funciona como alternativa a "go" para lanzar goroutines:

```go
// Ahora son equivalentes:
go myFunction()
=> myFunction()

// Ambos crean goroutines de la misma manera!
```

Este ejercicio demostro como las modificaciones a nivel de scanner pueden anadir nueva sintaxis con cambios minimos en el codigo.

---

*Continua al [Ejercicio 3](03-parser-multiple-go.es.md) o vuelve al [taller principal](../README.md)*
