# Ejercicio 2: Añadir el Operador Flecha "=>" para Goroutines

> 📖 **¿Quieres saber más?** Lee [The Scanner](https://internals-for-interns.com/es/posts/the-go-lexer/) en Internals for Interns para una explicación detallada de cómo funciona el lexer/scanner de Go.

En este ejercicio, añadirás un nuevo operador flecha "=>" a Go que funciona como una sintaxis alternativa para iniciar goroutines. Esto te enseñará cómo modificar el scanner de Go para reconocer nuevos operadores y mapearlos a funcionalidad existente.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Entender cómo el scanner de Go tokeniza operadores
- Saber cómo añadir nueva sintaxis de operadores a Go
- Modificar la lógica de análisis léxico del scanner
- Probar tu modificación del scanner con código funcional
- Extender con éxito el vocabulario de operadores de Go

## Contexto: Cómo Funciona esta Modificación del Scanner

Este ejercicio demuestra **modificaciones a nivel de scanner** para añadir nueva sintaxis de operadores a Go. Modificaremos la lógica del scanner para reconocer una nueva secuencia de operador "=>" y mapearla a un token existente. Esto es lo que lograremos:

- **Mejora del Scanner**: Añadir reconocimiento para la secuencia del operador "=>"
- **Mapeo de Token**: Mapear "=>" al token `_Go` existente (igual que la palabra clave "go")
- **Sintaxis Alternativa**: Crear `=> myFunction()` como equivalente a `go myFunction()`
- **Impacto Mínimo**: No se necesitan cambios en el parser ni en el compilador, solo lógica del scanner

Este enfoque nos permite crear sintaxis alternativa elegante sin modificar las partes más profundas del compilador.

## Paso 1: Navegar al Scanner

```bash
cd go/src/cmd/compile/internal/syntax
```

### Entender la Estructura del Scanner

Examinemos cómo el scanner maneja el operador "=" en `scanner.go`. Mira la línea 325:

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

El scanner primero consume el carácter "=" con `s.nextch()`, luego verifica si el siguiente carácter también es "=" (para el operador de comparación `==`). Si no lo es, se queda con "=" (asignación).

## Paso 2: Añadir la Lógica del Operador Flecha

Necesitamos añadir lógica para reconocer "=>" y tratarlo como el token `_Go`.

**Edita `scanner.go`:**

Encuentra el case "=" en la línea 325 y modifícalo para que también verifique ">":

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

### Entendiendo el Cambio en el Código

- **`if s.ch == '>'`**: Verifica si el siguiente carácter después de "=" es ">" (recuerda que `s.nextch()` ya consumió el "=")
- **`s.nextch()`**: Consume el carácter ">" del lexer
- **`s.lit = "=>"`**: Establece el valor literal para depuración y mensajes de error
- **`s.tok = _Go`**: Asigna el mismo token que la palabra clave "go"
- **`break`**: Sale del case para evitar caer en `_Assign`

## Paso 3: Recompilar el Compilador

Ahora recompilemos la toolchain de Go con nuestros cambios:

```bash
cd ../../../  # back to go/src
./make.bash
```

Si hay errores de compilación, revisa tus cambios y corrígelos.

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

Deberías ver una salida como esta:

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

Asegúrate de que no hemos roto el scanner:

```bash
cd /path/to/workshop/go/src
../bin/go test cmd/compile/internal/syntax -short
```

## Lo que Hicimos

1. **Modificamos la Lógica del Scanner**: Añadimos reconocimiento de "=>" al case existente de "="
2. **Reutilizamos un Token Existente**: Mapeamos "=>" al token `_Go` en lugar de crear uno nuevo
3. **Preservamos la Funcionalidad Existente**: Los operadores "=" y "==" siguen funcionando normalmente
4. **Impacto Mínimo del Cambio**: No se necesitaron cambios en el parser ni en el IR

## Lo que Aprendimos

- **Lógica del Scanner**: Cómo Go tokeniza secuencias de operadores
- **Reconocimiento de Operadores**: Añadir nuevos operadores mediante modificación del scanner
- **Reutilización de Tokens**: Mapear nueva sintaxis a tokens existentes
- **Estrategia de Testing**: Validar cambios del scanner con código real
- **Proceso de Compilación**: Recompilar Go con modificaciones del scanner

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir ":>" como otra alternativa a "go"
2. Añadir "~>" para operaciones asíncronas
3. Añadir ">>>" como operador de triple flecha
4. Hacer que el operador flecha funcione en diferentes contextos

## Siguientes Pasos

Has añadido exitosamente un nuevo operador al scanner de Go. Ahora entiendes cómo modificar el scanner para crear sintaxis alternativa para funcionalidad existente. Esta técnica se puede aplicar para crear otros atajos de operadores y azúcar sintáctico en el lenguaje.

En el Ejercicio 3, tomaremos un enfoque diferente y exploraremos **modificaciones del parser**, aprendiendo cómo modificar el parser para manejar múltiples tokens consecutivos.

## Limpieza

Para restaurar el código fuente original de Go:

```bash
cd /path/to/workshop/go/src/cmd/compile/internal/syntax
git checkout scanner.go
cd ../../../
./make.bash  # Recompilar con el código original
```

## Resumen

El operador flecha "=>" ahora funciona como alternativa a "go" para lanzar goroutines:

```go
// Ahora son equivalentes:
go myFunction()
=> myFunction()

// ¡Ambos crean goroutines de la misma manera!
```

Este ejercicio demostró cómo las modificaciones a nivel de scanner pueden añadir nueva sintaxis con cambios mínimos en el código.

---

*Continúa al [Ejercicio 3](03-parser-multiple-go.es.md) o vuelve al [taller principal](../README.md)*
