# Ejercicio 7: Go Paciente - Haciendo que Go Espere a las Goroutines

> 📖 **¿Quieres aprender mas?** Lee [The Bootstrap](https://internals-for-interns.com/posts/understanding-go-runtime/) y [The Scheduler](https://internals-for-interns.com/posts/go-runtime-scheduler/) en Internals for Interns para profundizar en el arranque del runtime de Go y la planificacion de goroutines.

En este ejercicio, modificaras el runtime de Go para que espere a que todas las goroutines terminen antes de que el programa finalice. Actualmente, cuando `main()` retorna, Go termina inmediatamente incluso si hay goroutines todavia en ejecucion. Haremos que Go sea "paciente" esperando a que todas las goroutines terminen.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender el proceso de terminacion de programas en Go
- Saber como contar las goroutines activas
- Modificar la funcion principal del runtime para cambiar el comportamiento del programa
- Entender los compromisos de la espera automatica de goroutines

## Contexto: Comportamiento Actual de Terminacion de Go

Actualmente, cuando escribes:

```go
package main

import "time"

func main() {
    go func() {
        time.Sleep(2 * time.Second)
        println("Goroutine finished!")
    }()
    println("Main finished!")
    // Program exits immediately, goroutine never completes
}
```

**Salida:**
```
Main finished!
```

La goroutine nunca llega a imprimir porque el programa finaliza cuando `main()` retorna.

Cambiaremos esto para que Go espere pacientemente a que todas las goroutines terminen:

**Nueva salida:**
```
Main finished!
Goroutine finished!
```

## Paso 1: Entender la Funcion Main del Runtime

La funcion `main()` del runtime de Go en `runtime/proc.go` es la responsable de ejecutar la funcion `main()` de tu programa. Examinemos como funciona:

```bash
cd go/src/runtime
```

Abre `proc.go` y busca la funcion `main()`. Cerca del principio (alrededor de las lineas 136-137), veras como el runtime se enlaza con el main de tu programa:

```go
//go:linkname main_main main.main
func main_main()
```

Esta directiva `//go:linkname` le dice al linker que conecte la funcion `main_main` del runtime con la funcion `main.main` de tu programa. Asi es como el runtime puede llamar a codigo de tu paquete main.

Mas abajo en la misma funcion `main()` (alrededor de la linea 289), veras donde se llama:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()

... // tear-down process continues
```

**Como funciona:**

1. Se ejecuta el proceso de bootstrap del runtime de Go
2. La funcion `main()` del runtime se ejecuta primero
3. Un poco mas de proceso de bootstrap
4. Se llama a `main_main` (que es la funcion `main()` de tu programa via linkname)
5. Tu funcion `main()` se ejecuta - **la responsabilidad se delega a tu codigo**
6. Cuando tu `main()` retorna, el control vuelve a la funcion `main()` del runtime
7. El runtime continua con el **proceso de desmontaje** del programa (limpieza y salida)

Actualmente, el desmontaje comienza inmediatamente despues de que tu `main()` retorna, sin esperar a otras goroutines.

## Paso 2: Anadir la Logica de Espera de Goroutines

Anadiremos codigo para esperar hasta que solo quede 1 goroutine (la propia goroutine principal).

**Edita `runtime/proc.go`:**

Busca la seccion alrededor de las lineas 289-290 donde se llama a `main_main`:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()
```

Anade la logica de espera justo despues de la llamada a `fn()`:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()

// Wait until only 1 goroutine is running (the main goroutine)
for gcount(false) > 1 {
	Gosched()
}
```

### Entendiendo el Codigo

- **`gcount(false)`** - Funcion del runtime que devuelve el numero de goroutines activas (el argumento `false` excluye las goroutines del sistema del conteo)
- **`gcount(false) > 1`** - Mientras haya mas goroutines aparte de la principal ejecutandose
- **`Gosched()`** - Cede el procesador, permitiendo que otras goroutines se ejecuten
- **El bucle termina** - Cuando solo queda la goroutine principal (conteo = 1)

## Paso 3: Recompilar la Cadena de Herramientas de Go

```bash
cd go/src
./make.bash
```

Esto recompila el runtime con tu logica de espera paciente de goroutines.

## Paso 4: Probar la Espera Basica de Goroutines

Crea un archivo de prueba para verificar el comportamiento:

Crea `patient_demo.go`:

```go
package main

import "time"

func main() {
	println("Main starting...")

	go func() {
		time.Sleep(1 * time.Second)
		println("Goroutine 1 finished!")
	}()

	go func() {
		time.Sleep(2 * time.Second)
		println("Goroutine 2 finished!")
	}()

	println("Main finished, but Go will wait...")
}
```

Ejecuta con tu Go modificado:

```bash
./bin/go run patient_demo.go
```

**Salida esperada:**

```
Main starting...
Main finished, but Go will wait...
Goroutine 1 finished!
Goroutine 2 finished!
```

Go ahora espera a que todas las goroutines terminen.

## Lo que Aprendimos

- **Terminacion de Programas**: Como los programas Go finalizan y hacen limpieza
- **Seguimiento de Goroutines**: La funcion `gcount()` rastrea las goroutines activas
- **Planificacion Cooperativa**: `Gosched()` cede para permitir que otras goroutines se ejecuten
- **Modificacion del Runtime**: Como un pequenio cambio afecta a todos los programas Go
- **Compromisos de Disenio**: Beneficios e inconvenientes de la espera automatica

## Ideas de Extension

Prueba estas modificaciones adicionales:

1. Anadir un timeout: Esperar un maximo de 10 segundos a las goroutines
2. Anadir registro: Imprimir cuando comienza la espera y que goroutines quedan
3. Hacerlo configurable: Usar una variable de entorno para activar/desactivar
4. Anadir una advertencia: Detectar bucles infinitos en goroutines

## Limpieza

Para restaurar el comportamiento estandar de Go:

```bash
cd go/src/runtime
git checkout proc.go
cd ..
./make.bash
```

## Resumen

Has modificado con exito el runtime de Go para que sea "paciente" y espere a todas las goroutines.

```
Antes:   main() retorna → salida inmediata → goroutines abandonadas
Despues: main() retorna → espera a las goroutines → todas terminan → salida

Cambios: funcion main() en runtime/proc.go
Resultado: Ninguna goroutine se queda atras
```

Esta modificacion demuestra:

- Comprension profunda del runtime de Go
- Como funciona la terminacion de programas
- La relacion entre main() y las goroutines
- Compromisos reales en el disenio de lenguajes

Tu Go ahora es paciente.

---

*Continua con el [Ejercicio 8](08-goroutine-sleep-detective.es.md) o vuelve al [taller principal](../README.md)*
