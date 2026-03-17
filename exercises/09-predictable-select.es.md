# Ejercicio 9: Select Predecible - Haciendo las Sentencias Select Deterministas

> 📖 **¿Quieres saber más?** Lee [The Scheduler](https://internals-for-interns.com/es/posts/go-runtime-scheduler/) en Internals for Interns para una exploración en profundidad del runtime y la planificación de goroutines en Go.

En este ejercicio, modificarás la sentencia `select` de Go para que sea determinista en lugar de aleatoria. Por defecto, Go aleatoriza qué caso se elige cuando varios channels están listos. Nosotros lo cambiaremos para que siempre elija los casos en el mismo orden.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Comprender cómo está implementada la sentencia `select` de Go
- Saber por qué Go usa aleatorización (equidad vs. inanición)
- Modificar el algoritmo de selección de channels en el runtime
- Probar el comportamiento de selección determinista vs. aleatorio

## Introducción: ¿Cómo Funciona Select Internamente?

La sentencia `select` está implementada en la función del runtime `selectgo()` en `runtime/select.go`. Cuando tu código llega a un `select` con múltiples casos, el runtime necesita decidir qué caso ejecutar. Lo hace usando dos arrays:

- **`pollorder`**: Determina el orden en que los casos se **comprueban** para ver si están listos. Por defecto, este orden se aleatoriza usando `cheaprandn()` para asegurar la equidad — ningún channel tiene prioridad sobre los demás.
- **`lockorder`**: Determina el orden en que se **adquieren** los locks de los channels (ordenados por dirección para prevenir deadlocks).

El runtime primero mezcla los casos en `pollorder`, luego los recorre en ese orden. Si el channel de un caso está listo (tiene datos para recibir o espacio para enviar), se selecciona ese caso. Si ningún caso está listo y hay un `default`, se ejecuta el default. De lo contrario, la goroutine se aparca en las colas de espera de todos los channels y espera hasta que uno esté listo.

La aleatorización en `pollorder` es lo que hace que `select` sea no determinista — ejecutar el mismo `select` con los mismos channels listos elegirá casos diferentes cada vez. Esta es una decisión de diseño deliberada para evitar que los programas dependan accidentalmente del orden de los casos.

## Contexto: Go Aleatoriza el Select

Por defecto, cuando varios channels están listos, Go aleatoriza cuál se ejecuta:

```go
select {
case v := <-ch1:  // Sometimes chosen
case v := <-ch2:  // Sometimes chosen
case v := <-ch3:  // Sometimes chosen
}
// Random selection prevents starvation
```

Lo haremos determinista:

```go
select {
case v := <-ch1:  // ALWAYS chosen first when ready
case v := <-ch2:  // Only if ch1 not ready
case v := <-ch3:  // Only if ch1 and ch2 not ready
}
// Predictable, source-order selection
```

## Paso 1: Crear un Test para Ver la Aleatorización Actual

Crea un archivo `random_select_demo.go`:

```go
package main

func main() {
    ch1 := make(chan int, 1)
    ch2 := make(chan int, 1)
    ch3 := make(chan int, 1)

    // Fill all channels so they're all ready
    ch1 <- 1
    ch2 <- 2
    ch3 <- 3

    // Run select 10 times to see randomization
    for i := 0; i < 10; i++ {
        select {
        case v := <-ch1:
            println("Round", i, ": Selected ch1 (value", v, ")")
            ch1 <- 1 // Refill
        case v := <-ch2:
            println("Round", i, ": Selected ch2 (value", v, ")")
            ch2 <- 2 // Refill
        case v := <-ch3:
            println("Round", i, ": Selected ch3 (value", v, ")")
            ch3 <- 3 // Refill
        }
    }
}
```

Ejecuta con el Go actual para ver la selección aleatoria:

```bash
go run random_select_demo.go
```

La salida muestra selección aleatoria:

```
Round 0: Selected ch3 (value 3)
Round 1: Selected ch1 (value 1)
Round 2: Selected ch2 (value 2)
...
```

## Paso 2: Navegar a la Implementación del Select

```bash
cd go/src/runtime
```

El archivo `select.go` contiene toda la implementación de la sentencia select. La función clave es `selectgo()`, que se encarga de la selección de casos.

## Paso 3: Comprender el Código de Aleatorización

Busca alrededor de la línea 191 en `select.go`:

```go
// go/src/runtime/select.go:191
j := cheaprandn(uint32(norder + 1))  // Random index!
pollorder[norder] = pollorder[j]
pollorder[j] = uint16(i)
norder++
```

Esto implementa el algoritmo para aleatorizar el orden de los casos:

- `cheaprandn()` genera un número pseudoaleatorio
- Los casos se colocan en posiciones aleatorias en el array `pollorder`
- Luego select comprueba los casos en este orden aleatorizado

## Paso 4: Hacer el Select Determinista

**Edita `select.go`:**

Encuentra la línea 191 y cambia la aleatorización para que sea determinista:

```go
// go/src/runtime/select.go:191
// Original:
j := cheaprandn(uint32(norder + 1))
pollorder[norder] = pollorder[j]
pollorder[j] = uint16(i)

// Change to:
pollorder[norder] = uint16(len(scases)-1-i)
```

### Entendiendo el Cambio en el Código


- **`uint16(len(scases)-1-i)`**: Se usa orden inverso aquí
- **Resultado**: pollorder ahora siempre está ordenado en el orden del código fuente
- **Efecto**: Los casos mantienen su orden del código fuente en `pollorder`

## Paso 5: Recompilar el Runtime de Go

```bash
cd ../  # back to go/src
./make.bash
```

## Paso 6: Probar el Comportamiento Determinista

```bash
../go/bin/go run random_select_demo.go
```

Ahora deberías ver una **salida determinista**:

```
Round 0: Selected ch1 (value 1)
Round 1: Selected ch1 (value 1)
Round 2: Selected ch1 (value 1)
Round 3: Selected ch1 (value 1)
...
```

¡Perfecto! `ch1` es **siempre** elegido porque es el primero en el código, no más orden aleatorio.

## Entendiendo lo que Hicimos

1. **Eliminamos la Aleatorización**: Reemplazamos `cheaprandn()` con un índice determinista
2. **Mantuvimos el Orden del Código**: Los casos ahora se comprueban en el orden en que aparecen
3. **Mejora de Rendimiento**: Ligeramente más rápido (sin generación de números aleatorios)
4. **Cambio de Semántica**: Misma sintaxis, comportamiento diferente en tiempo de ejecución

## Lo que Aprendimos

- **Modificación del Runtime**: Cómo alterar el comportamiento fundamental del lenguaje
- **Compromisos de Diseño**: Equidad vs. determinismo en sistemas concurrentes
- **Internos de Select**: Cómo funcionan `selectgo` y `pollorder`
- **Pruebas de Comportamiento**: Validar cambios semánticos con programas de prueba

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir un modo de orden inverso (comprobar casos del último al primero)
2. Añadir niveles de prioridad basados en la posición del caso
3. Registrar estadísticas de selección para depuración
4. Hacer la aleatorización configurable mediante una variable de entorno

## Limpieza

Para restaurar el comportamiento aleatorio original de Go:

```bash
cd go/src/runtime
git checkout select.go
cd ../
./make.bash
```

## Resumen

Has transformado el `select` de Go de un selector aleatorio y equitativo en un sistema de prioridad predecible y determinista:

```go
// Before: Random selection (fair but unpredictable)
select {
case <-ch1: // 33% chance
case <-ch2: // 33% chance
case <-ch3: // 33% chance
}

// After: Deterministic selection (predictable but may starve)
select {
case <-ch1: // Always chosen when ready
case <-ch2: // Only if ch1 not ready
case <-ch3: // Only if ch1 and ch2 not ready
}
```

Este ejercicio demostró cómo las modificaciones del runtime pueden cambiar fundamentalmente el comportamiento del lenguaje y expuso compromisos importantes en el diseño de sistemas concurrentes.

---

*Continúa con el [Ejercicio 10](10-java-style-stack-traces.es.md) o vuelve al [taller principal](../README.md)*
