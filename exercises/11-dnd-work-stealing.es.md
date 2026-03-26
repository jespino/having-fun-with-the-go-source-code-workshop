# Ejercicio 11: D&D Work Stealing - Tirando Dados por Goroutines

> **¿Quieres aprender más?** Lee [The Scheduler](https://internals-for-interns.com/posts/go-runtime-scheduler/) en Internals for Interns para una inmersión profunda en el runtime de Go y la planificación de goroutines.

En este ejercicio, añadirás una tirada de dado d20 al algoritmo de work stealing del planificador de Go. Cuando un procesador (P) intenta robar goroutines de la cola de ejecución de otro P, primero debe sacar más de 10 en un dado de veinte caras. Las tiradas fallidas significan que el robo se bloquea, haciendo visible y entretenida la distribución de trabajo del planificador.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Comprender cómo el planificador de work stealing de Go distribuye goroutines entre procesadores
- Saber dónde se encuentra la función `stealWork` y cómo itera sobre los demás P's
- Modificar la lógica de robo para añadir una puerta aleatoria
- Observar los intentos de work stealing en tiempo real

## Contexto: Work Stealing

El planificador de Go utiliza work stealing para equilibrar la carga entre procesadores. Cuando un P se queda sin goroutines que ejecutar, mira las colas de otros P's y roba la mitad de su trabajo:

```
Antes (comportamiento actual):
  P0: [g1, g2, g3, g4]    P1: []  (idle)
  P1 intenta robar de P0 → siempre tiene éxito
  P0: [g1, g2]             P1: [g3, g4]

Después (nuestra modificación):
  P0: [g1, g2, g3, g4]    P1: []  (idle)
  P1 tira d20 para robar de P0 → sacó 7, ¡falló!
  P1 tira d20 para robar de P0 → sacó 16, ¡robó!
  P0: [g1, g2]             P1: [g3, g4]
```

## Paso 1: Entender el Mecanismo de Robo

La lógica de work stealing se encuentra en la función `stealWork` en `proc.go`:

```bash
cd go/src/runtime
grep -n "func stealWork" proc.go
```

Encontrarás `stealWork` alrededor de la línea 3828. Esta función es llamada por `findRunnable` cuando un P no tiene trabajo local. Itera sobre los demás P's en orden aleatorio, intentando robar goroutines de sus colas de ejecución.

## Paso 2: Encontrar el Intento de Robo

Dentro de `stealWork`, busca el intento de robo real alrededor de la línea 3883:

```bash
grep -n "runqsteal" proc.go | head -5
```

Verás este bloque de código (alrededor de las líneas 3883-3887):

```go
// go/src/runtime/proc.go:3883-3887
// Don't bother to attempt to steal if p2 is idle.
if !idlepMask.read(enum.position()) {
    if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
        return gp, false, now, pollUntil, ranTimer
    }
}
```

Variables clave en este punto:
- **`pp`** - El P actual (el ladrón), tipo `*p`, tiene el campo `pp.id` (int32)
- **`p2`** - El P objetivo (la víctima), tipo `*p`, tiene el campo `p2.id`
- **`runqsteal(pp, p2, ...)`** - Mueve goroutines de la cola de p2 a la cola de pp

## Paso 3: Añadir la Tirada de Dado de D&D

Reemplaza las líneas 3883-3887 con nuestra versión con puerta de dado:

```go
// go/src/runtime/proc.go:3883-3887
// Don't bother to attempt to steal if p2 is idle.
if !idlepMask.read(enum.position()) {
    if mainStarted && gogetenv("GODND") != "" {
        // D&D Work Stealing: Roll a d20 to attempt stealing!
        roll := cheaprandn(20) + 1 // Roll 1-20
        if roll > 10 {
            if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
                println("🎲 [P", pp.id, "] Rolling to steal from P", p2.id, "... rolled", roll, ". Stole!")
                return gp, false, now, pollUntil, ranTimer
            }
        } else {
            println("🎲 [P", pp.id, "] Rolling to steal from P", p2.id, "... rolled", roll, ". Failed!")
        }
    } else {
        if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
            return gp, false, now, pollUntil, ranTimer
        }
    }
}
```

### Entendiendo el Código

- **`mainStarted`** - Un booleano que ya existe en `proc.go` y cambia a `true` al inicio de la goroutine principal — antes de que se ejecuten `sysmon`, GC e `init()`. Elimina el ruido más temprano del planificador, pero algunos prints pre-main seguirán apareciendo (ver la nota más abajo)
- **`gogetenv("GODND")`** - El lector interno de variables de entorno del runtime (equivalente a `os.Getenv` — el runtime no puede importar `os`). Todo queda protegido detrás de `GODND=1` para que el planificador se comporte normalmente a menos que lo actives
- **`cheaprandn(20) + 1`** - Tira 1-20. `cheaprandn(20)` devuelve 0-19, el `+ 1` lo desplaza al rango correcto de un d20
- **`roll > 10`** - 50% de probabilidad de éxito (tiradas 11-20 tienen éxito, tiradas 1-10 fallan)
- **`println(...)`** - La función print integrada del runtime, escribe a stderr mediante syscall directo, no necesita imports
- **`pp.id` / `p2.id`** - Los campos de ID del procesador (int32), definidos en el struct `p` en `runtime2.go`
- Solo imprimimos "Stole!" cuando `runqsteal` devuelve non-nil (la cola objetivo podría haberse vaciado entre la comprobación de idle y el robo)

> **🐉 Dato curioso: ¡El planificador ya tira con ventaja — multiplicada por cuatro!**
>
> Mira el bucle exterior en `stealWork`: `const stealTries = 4`. El planificador no intenta robar solo una vez — hace un bucle **4 veces** sobre todos los P's, reordenándolos con `cheaprand()` en cada pasada. Así que tu puerta de d20 no se tira solo una vez por intento de robo — un P determinado tiene hasta 4 oportunidades por objetivo. En términos de D&D, eso es tirar con ventaja... al cuadrado.
>
> Y la 4ª pasada es especial: establece `stealTimersOrRunNextG = true`, lo que desbloquea el robo de la goroutine `runnext` de la víctima — la que estaba a punto de ejecutar. El comentario del código fuente dice literalmente *"stealing from the other P's runnext should be the last resort."* Así que la pasada final es la ronda desesperada, sin guantes, donde todo vale.
>
> El planificador de Go ya estaba jugando a D&D antes de que llegaras. Tú solo estás haciendo visibles los dados.

## Paso 4: Recompilar el Toolchain de Go

```bash
cd ../  # volver a go/src
./make.bash
```

## Paso 5: Probar el Planificador D&D

Crea el archivo `dnd_steal_demo.go`:

```bash
# Crear el archivo
touch /tmp/dnd_steal_demo.go
```

```go
package main

import (
    "runtime"
    "sync"
)

func busyWork(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    sum := 0
    for i := 0; i < 1_000_000; i++ {
        sum += i
    }
    println("Goroutine", id, "finished")
}

func main() {
    runtime.GOMAXPROCS(4) // 4 P's para que el robo sea visible
    println("=== D&D Work Stealing Demo ===")
    println()

    var wg sync.WaitGroup
    for i := 1; i <= 20; i++ {
        wg.Add(1)
        go busyWork(i, &wg)
    }
    wg.Wait()
    println()
    println("=== All goroutines completed! ===")
}
```

Compila y ejecuta con el modo D&D activado:

```bash
../go/bin/go build -o dnd_steal_demo /tmp/dnd_steal_demo.go
GODND=1 ./dnd_steal_demo
```

**¿Por qué `GODND=1`?** Sin esta variable, el planificador funciona normalmente — sin tiradas de dados, sin prints. Esto mantiene `./make.bash` y `go build` limpios y silenciosos. Ten en cuenta que algunos prints pre-main y post-main seguirán apareciendo incluso con `GODND=1` — consulta las notas más abajo para saber por qué.

**¿Por qué compilar primero y luego ejecutar?** El comando `go build` en sí mismo usa tu Go modificado. Compilar por separado (sin `GODND=1`) mantiene silencioso el work stealing del compilador.

Salida esperada (varía en cada ejecución):

```
🎲 [P 3 ] Rolling to steal from P 0 ... rolled 13 . Stole!
=== D&D Work Stealing Demo ===

🎲 [P 2 ] Rolling to steal from P 0 ... rolled 15 . Stole!
🎲 [P 3 ] Rolling to steal from P 0 ... rolled 12 . Stole!
🎲 [P 1 ] Rolling to steal from P 0 ... rolled 11 . Stole!
Goroutine 15 finished
🎲 [P 0 ] Rolling to steal from P 2 ... rolled 7 . Failed!
🎲 [P 0 ] Rolling to steal from P 3 ... rolled 20 . Stole!
Goroutine 1 finished
Goroutine 12 finished
...
=== All goroutines completed! ===
🎲 [P %
```

Verás las tiradas de dados intercaladas con los mensajes de finalización de goroutines. Las tiradas de 1-10 fallan, las tiradas de 11-20 tienen éxito. ¡Observa cómo un P sin trabajo sigue reintentando con diferentes objetivos hasta que saca una tirada suficientemente alta!

> **📖 ¿Por qué hay tiradas de dados antes de `=== D&D Work Stealing Demo ===`?**
>
> Tú no las escribiste — lo hizo el runtime de Go. Entre que `mainStarted` cambia a `true` y tu primer `println`, el runtime inicia `sysmon`, activa el GC y ejecuta todas las funciones `init()` — cada una lanzando goroutines que los P's ociosos intentan robar inmediatamente. Has instrumentado el planificador en sí, así que ahora estás viendo actividad que siempre estuvo ahí, solo que en silencio.

> **📖 ¿Por qué hay un `🎲 [P %` truncado después de `=== All goroutines completed! ===`?**
>
> La misma historia, pero al revés. Después de que `wg.Wait()` retorna, los P's ociosos siguen girando en `stealWork` buscando más trabajo. Uno empieza un `println` justo cuando `main()` retorna y el proceso termina — el print nunca se completa. El `%` es tu shell indicando que la salida no terminó con un salto de línea. El planificador tampoco se detiene por ti a la salida.

> ¡Bienvenido al interior de Go!


## Entendiendo lo que Hicimos

1. **Encontramos la Lógica de Robo**: Localizamos `stealWork` en `proc.go` donde los P's roban goroutines entre sí
2. **Añadimos una Puerta de Dado**: Usamos `cheaprandn(20) + 1` para generar una tirada de d20 (1-20) antes de cada intento de robo
3. **Registramos las Tiradas**: Añadimos llamadas a `println()` mostrando qué P está robando de cuál, el resultado de la tirada y si tuvo éxito
4. **Observamos el Efecto**: Vimos los intentos de work stealing en tiempo real, con algunos fallando debido a tiradas bajas

## Lo que Aprendimos

- **Work Stealing**: Cómo Go distribuye goroutines entre procesadores cuando las colas están desequilibradas
- **Función `stealWork`**: El bucle principal que itera sobre los P's buscando trabajo que robar
- **`cheaprandn`**: El generador de números pseudoaleatorios rápido del runtime, usado en todo el planificador
- **Observabilidad del Planificador**: Cómo añadir logging al planificador sin romper su comportamiento
- **Identidad de P**: Cada procesador tiene un campo `id` único que lo identifica en las decisiones de planificación

## Ideas de Extensión

1. Hacer la dificultad configurable: la tirada debe superar 15 en lugar de 10 (robos más difíciles)
2. Añadir un "golpe crítico" con un 20 natural: robar TODAS las goroutines del objetivo, no solo la mitad
3. Añadir una "pifia" con un 1 natural: el P que roba cede un ciclo con `Gosched()`
4. Registrar e imprimir el total de tiradas, éxitos y fallos al finalizar el programa

## Limpieza

Para eliminar la tirada de dado:

```bash
cd go/src/runtime
git checkout proc.go
cd ../
./make.bash
```

## Resumen

Has convertido el planificador de work stealing de Go en un encuentro de RPG de mesa:

```
Antes:   P intenta robar -> siempre tiene éxito si el objetivo tiene trabajo
Después: P intenta robar -> debe sacar > 10 en un d20 primero

Cambios: runtime/proc.go función stealWork() (~14 líneas)
Resultado: ¡El planificador ahora juega a D&D!
```

---

*¡Felicidades por completar todos los ejercicios del taller! Vuelve al [taller principal](../README.md)*
