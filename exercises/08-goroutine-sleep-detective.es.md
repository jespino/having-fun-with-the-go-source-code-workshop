# Ejercicio 8: Detective de Goroutines Dormidas - Monitoreo del Estado del Runtime

> 📖 **¿Quieres saber más?** Lee [The Scheduler](https://internals-for-interns.com/es/posts/go-runtime-scheduler/) en Internals for Interns para una exploración en profundidad de la planificación de goroutines y las transiciones de estado en Go.

En este ejercicio, modificarás el scheduler del runtime de Go para registrar las transiciones de estado de las goroutines. Cada vez que una goroutine se duerma esperando algo, se anunciará: "Hello, I'm goroutine 42, going to sleep waiting for channel receive".

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Comprender las transiciones de estado del scheduler de goroutines de Go
- Saber dónde se bloquean las goroutines en el runtime
- Modificar el scheduler para obtener información de depuración

## Introducción: ¿Cómo Funciona el Scheduler?

El scheduler de Go usa el **modelo GMP** (Goroutines, Machines, Processors) para mapear potencialmente miles de goroutines sobre un pequeño número de hilos del sistema operativo. La idea clave es que cuando un hilo del SO se bloquea (por ejemplo, en una syscall), los recursos de planificación (el P) pueden desacoplarse y moverse a otro hilo, manteniendo el flujo de trabajo.

Las goroutines no tienen un hilo de scheduler dedicado que las gestione. En su lugar, gestionan sus propias transiciones mediante un **patrón de autoservicio**: cuando una goroutine necesita esperar (por un channel, mutex, sleep, etc.), llama a `gopark()` que se aparca a sí misma, se añade a la cola de espera apropiada, y luego llama a `schedule()` para encontrar la siguiente goroutine ejecutable. Cuando la condición de espera se satisface, `goready()` mueve la goroutine de vuelta al estado ejecutable.

El scheduler elige la siguiente goroutine a ejecutar siguiendo un orden de prioridad: primero trabajo del GC, luego el slot local `runnext`, luego la cola local de ejecución, luego la cola global (comprobada cada 61 llamadas para prevenir la inanición), luego resultados del network poller, y finalmente work-stealing de otros Ps.

Entender este flujo de planificación es esencial porque en este ejercicio añadiremos registros en el punto exacto donde las goroutines transicionan al estado de espera.

## Contexto: Estados de una Goroutine

Go gestiona las goroutines a través de diferentes estados:

- **`_Grunnable`** - Lista para ejecutarse pero sin estar en ejecución
- **`_Grunning`** - En ejecución actualmente
- **`_Gwaiting`** - Bloqueada esperando algo (¡nuestro objetivo!)
- **`_Gsyscall`** - Ejecutando una llamada al sistema
- ...

Cuando una goroutine necesita esperar (por channels, mutexes, sleep, etc.), se "aparca" y pasa al estado `_Gwaiting`.

## Paso 1: Comprender el Mecanismo de Aparcamiento

La función `gopark` es invocada por TODAS las primitivas de sincronización cuando una goroutine necesita esperar.

```bash
cd go/src/runtime
grep -n "func gopark" proc.go
```

Funciones clave:

- **`gopark()`** - Inicia el aparcamiento de una goroutine
- **`park_m()`** - Cambia efectivamente el estado a `_Gwaiting`

## Paso 2: Encontrar el Código de Transición de Estado

```bash
# Observa dónde cambia realmente el estado
grep -n -A 5 "func park_m" proc.go
```

Alrededor de la línea 4275, verás:

```go
casgstatus(gp, _Grunning, _Gwaiting)
```

Esta es la línea exacta donde una goroutine pasa de en ejecución a en espera. ¡Perfecto para nuestro registro!

## Paso 3: Añadir el Registro de Goroutines Dormidas

**Edita `proc.go`:**

Necesitarás añadir registros en tres ubicaciones donde las goroutines pasan al estado de espera:

### Ubicación 1: Función `casGToWaiting` (alrededor de la línea 1388)

Encuentra la función `casGToWaiting` y añade el registro después de establecer el motivo de espera:

```go
func casGToWaiting(gp *g, old uint32, reason waitReason) {
	// Set the wait reason before calling casgstatus, because casgstatus will use it.
	gp.waitreason = reason
	if gp.goid > 1 { // Skip system goroutines 0 and 1
		print("Hello, I'm goroutine ", gp.goid, ", going to sleep waiting for ", gp.waitreason.String(), "\n")
	}
	casgstatus(gp, old, _Gwaiting)
}
```

### Ubicación 2: Función `casGFromPreempted` (alrededor de la línea 1430)

Encuentra donde las goroutines interrumpidas pasan al estado de espera. Añade el registro después de establecer el `waitreason` pero antes del `CompareAndSwap`:

```go
func casGFromPreempted(gp *g, old, new uint32) bool {
	if old != _Gpreempted || new != _Gwaiting {
		throw("bad g transition")
	}
	gp.waitreason = waitReasonPreempted
	if gp.goid > 1 { // Skip system goroutines 0 and 1
		print("Hello, I'm goroutine ", gp.goid, ", going to sleep waiting for ", gp.waitreason.String(), "\n")
	}
	if !gp.atomicstatus.CompareAndSwap(_Gpreempted, _Gwaiting) {
		return false
	}
	if bubble := gp.bubble; bubble != nil {
		bubble.changegstatus(gp, _Gpreempted, _Gwaiting)
	}
	return true
}
```

### Ubicación 3: Función `park_m` (alrededor de la línea 4275)

Encuentra la función `park_m` y añade el registro antes de la llamada directa a `casgstatus`:

```go
// Add this before: casgstatus(gp, _Grunning, _Gwaiting)
if gp.goid > 1 { // Skip system goroutines 0 and 1
    print("Hello, I'm goroutine ", gp.goid, ", going to sleep waiting for ", gp.waitreason.String(), "\n")
}
casgstatus(gp, _Grunning, _Gwaiting)
```

### Entendiendo el Código

- **`gp.goid`** - ID único de la goroutine
- **`gp.waitreason.String()`** - Motivo de espera legible (channel, mutex, sleep, etc.)
- **`print()`** - Función de impresión del runtime (escribe en stderr)
- **`gp.goid > 1`** - Omite las goroutines del sistema para reducir el ruido

## Paso 4: Recompilar el Runtime de Go

```bash
cd ../  # back to go/src
./make.bash
```

## Paso 5: Probar el Bloqueo en Channels

Crea un archivo `channel_demo.go`:

```go
package main

import "time"

func main() {
    ch := make(chan string)

    // Start goroutine that will block on receive
    go func() {
        msg := <-ch  // Should trigger our logging!
        println("Received:", msg)
    }()

    // Let the goroutine start and block
    time.Sleep(100 * time.Millisecond)

    // Send something
    ch <- "Hello!"
    time.Sleep(10 * time.Millisecond)
}
```

Compila y ejecuta con nuestro Go modificado:

```bash
../go/bin/go build channel_demo.go
./channel_demo
```

**Nota:** Primero compilamos el binario y luego lo ejecutamos directamente. Esto evita mezclar las goroutines del compilador/proceso de compilación con las goroutines de nuestro programa, obteniendo una salida más limpia.

Salida esperada:

```
Hello, I'm goroutine 4, going to sleep waiting for GC scavenge wait
Hello, I'm goroutine 3, going to sleep waiting for GC sweep wait
Hello, I'm goroutine 2, going to sleep waiting for force gc (idle)
Hello, I'm goroutine 6, going to sleep waiting for chan receive
Hello, I'm goroutine 5, going to sleep waiting for GOMAXPROCS updater (idle)
Received: Hello!
```

Ahora puedes ver las goroutines bloqueándose.

## Entendiendo lo que Hicimos

1. **Encontramos la Función de Aparcamiento**: Localizamos dónde las goroutines pasan al estado de espera
2. **Añadimos Registro**: Insertamos una instrucción print antes del cambio de estado
3. **Capturamos el Motivo de Espera**: Usamos `gp.waitreason.String()` para una salida legible
4. **Probamos Escenarios**: Verificamos con channels, mutexes, sleep y select

Motivos de espera comunes que verás:

- `chan receive` / `chan send`
- `sync mutex lock`
- `sleep`
- `GC`

## Lo que Aprendimos

- **Ciclo de Vida de una Goroutine**: Cómo las goroutines transicionan entre estados
- **Mecanismo de Aparcamiento**: Las funciones `gopark` y `park_m`
- **Internos de Sincronización**: Dónde los channels, mutexes y select causan bloqueos
- **Depuración del Runtime**: Cómo añadir observabilidad al runtime de Go
- **Visibilidad de la Concurrencia**: Observación en tiempo real de las operaciones de bloqueo

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir registro de despertar de goroutines (cuando reanudan la ejecución)
2. Añadir iconos para diferentes motivos de espera (channel, mutex, sleep)
3. Incluir marcas de tiempo para medir la duración del bloqueo
4. Filtrar el registro solo por motivos de espera específicos

## Limpieza

Para eliminar el registro:

```bash
cd go/src/runtime
git checkout proc.go
cd ../
./make.bash
```

## Resumen

¡Has obtenido visión de rayos X del modelo de concurrencia de Go! Tu runtime modificado ahora anuncia cada operación de bloqueo de goroutines:

```
Hello, I'm goroutine 18, going to sleep waiting for chan receive
Hello, I'm goroutine 19, going to sleep waiting for sync mutex lock
Hello, I'm goroutine 20, going to sleep waiting for sleep
```

Este ejercicio reveló el funcionamiento interno del scheduler de Go y cómo las primitivas de sincronización interactúan con el runtime.

---

*Continúa con el [Ejercicio 9](09-predictable-select.es.md) o vuelve al [taller principal](../README.md)*
