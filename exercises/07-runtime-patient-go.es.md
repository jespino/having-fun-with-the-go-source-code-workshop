# Ejercicio 7: Go Paciente - Haciendo que Go Espere a las Goroutines

> 📖 **¿Quieres aprender más?** Lee [The Bootstrap](https://internals-for-interns.com/es/posts/understanding-go-runtime/) y [The Scheduler](https://internals-for-interns.com/es/posts/go-runtime-scheduler/) en Internals for Interns para profundizar en el arranque del runtime de Go y la planificación de goroutines.

En este ejercicio, modificarás el runtime de Go para que espere a que todas las goroutines terminen antes de que el programa finalice. Actualmente, cuando `main()` retorna, Go termina inmediatamente incluso si hay goroutines todavía en ejecución. Haremos que Go sea "paciente" esperando a que todas las goroutines terminen.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Entender el proceso de terminación de programas en Go
- Saber cómo contar las goroutines activas
- Modificar la función principal del runtime para cambiar el comportamiento del programa
- Entender los compromisos de la espera automática de goroutines

## Contexto: Comportamiento Actual de Terminación de Go

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

## Paso 1: Entender la Función Main del Runtime

La función `main()` del runtime de Go en `runtime/proc.go` es la responsable de ejecutar la función `main()` de tu programa. Examinemos cómo funciona:

```bash
cd go/src/runtime
```

Abre `proc.go` y busca la función `main()`. Cerca del principio (alrededor de las líneas 136-137), verás cómo el runtime se enlaza con el main de tu programa:

```go
//go:linkname main_main main.main
func main_main()
```

Esta directiva `//go:linkname` le dice al linker que conecte la función `main_main` del runtime con la función `main.main` de tu programa. Así es como el runtime puede llamar a código de tu paquete main.

Más abajo en la misma función `main()` (alrededor de la línea 289), verás dónde se llama:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()

... // tear-down process continues
```

**Cómo funciona:**

1. Se ejecuta el proceso de bootstrap del runtime de Go
2. La función `main()` del runtime se ejecuta primero
3. Un poco más de proceso de bootstrap
4. Se llama a `main_main` (que es la función `main()` de tu programa vía linkname)
5. Tu función `main()` se ejecuta - **la responsabilidad se delega a tu código**
6. Cuando tu `main()` retorna, el control vuelve a la función `main()` del runtime
7. El runtime continúa con el **proceso de desmontaje** del programa (limpieza y salida)

Actualmente, el desmontaje comienza inmediatamente después de que tu `main()` retorna, sin esperar a otras goroutines.

## Paso 2: Añadir la Lógica de Espera de Goroutines

Añadiremos código para esperar hasta que solo quede 1 goroutine (la propia goroutine principal).

**Edita `runtime/proc.go`:**

Busca la sección alrededor de las líneas 289-290 donde se llama a `main_main`:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()
```

Añade la lógica de espera justo después de la llamada a `fn()`:

```go
fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
fn()

// Wait until only 1 goroutine is running (the main goroutine)
for gcount(false) > 1 {
	Gosched()
}
```

### Entendiendo el Código

- **`gcount(false)`** - Función del runtime que devuelve el número de goroutines activas (el argumento `false` excluye las goroutines del sistema del conteo)
- **`gcount(false) > 1`** - Mientras haya más goroutines aparte de la principal ejecutándose
- **`Gosched()`** - Cede el procesador, permitiendo que otras goroutines se ejecuten
- **El bucle termina** - Cuando solo queda la goroutine principal (conteo = 1)

## Paso 3: Recompilar la Cadena de Herramientas de Go

```bash
cd go/src
./make.bash
```

Esto recompila el runtime con tu lógica de espera paciente de goroutines.

## Paso 4: Probar la Espera Básica de Goroutines

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

¡Go ahora espera a que todas las goroutines terminen!

## Lo que Aprendimos

- **Terminación de Programas**: Cómo los programas Go finalizan y hacen limpieza
- **Seguimiento de Goroutines**: La función `gcount()` rastrea las goroutines activas
- **Planificación Cooperativa**: `Gosched()` cede para permitir que otras goroutines se ejecuten
- **Modificación del Runtime**: Cómo un pequeño cambio afecta a todos los programas Go
- **Compromisos de Diseño**: Beneficios e inconvenientes de la espera automática

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir un timeout: Esperar un máximo de 10 segundos a las goroutines
2. Añadir registro: Imprimir cuando comienza la espera y qué goroutines quedan
3. Hacerlo configurable: Usar una variable de entorno para activar/desactivar
4. Añadir una advertencia: Detectar bucles infinitos en goroutines

## Limpieza

Para restaurar el comportamiento estándar de Go:

```bash
cd go/src/runtime
git checkout proc.go
cd ..
./make.bash
```

## Resumen

¡Has modificado con éxito el runtime de Go para que sea "paciente" y espere a todas las goroutines!

```
Antes:   main() retorna → salida inmediata → goroutines abandonadas
Después: main() retorna → espera a las goroutines → todas terminan → salida

Cambios: función main() en runtime/proc.go
Resultado: ¡Ninguna goroutine se queda atrás!
```

Esta modificación demuestra:

- Comprensión profunda del runtime de Go
- Cómo funciona la terminación de programas
- La relación entre main() y las goroutines
- Compromisos reales en el diseño de lenguajes

Tu Go ahora es paciente.

---

*Continúa con el [Ejercicio 8](08-goroutine-sleep-detective.es.md) o vuelve al [taller principal](../README.md)*
