# Ejercicio 6: Pasada SSA - Deteccion de Divisiones por Potencias de Dos

> 📖 **¿Quieres aprender mas?** Lee [The SSA Phase](https://internals-for-interns.com/posts/the-go-ssa/) en Internals for Interns para profundizar en las pasadas de optimizacion SSA de Go.

En este ejercicio, aprenderas como funcionan las pasadas del compilador SSA (Static Single Assignment) de Go creando una pasada de optimizacion personalizada que detecta operaciones de division por potencias de dos.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender la arquitectura de pasadas del compilador SSA de Go
- Saber como recorrer bloques y valores SSA
- Crear una pasada de analisis personalizada desde cero
- Integrar tu pasada en el pipeline del compilador
- Usar los volcados SSA para verificar que tu pasada funciona

## Contexto: Pasadas del Compilador SSA

El compilador de Go transforma tu codigo a traves de multiples pasadas:

1. **Parseo** - Convertir el codigo fuente a AST
2. **Verificacion de Tipos** - Comprobar que los tipos son correctos
3. **Generacion de IR** - Convertir a forma IR (representacion intermedia)
3. **Generacion de SSA** - Convertir a forma SSA (Static Single Assignment)
4. **Pasadas de Optimizacion** - Transformar el SSA (nuestro enfoque)
5. **Generacion de Codigo** - Producir codigo maquina

Vamos a trabajar con la forma SSA para conocer la posibilidad de optimizar potencias de dos.

## Paso 1: Entender la Estructura de las Pasadas SSA

Las pasadas SSA se registran en `compile.go` y operan sobre funciones. Examinemos la estructura:

```bash
cd go/src/cmd/compile/internal/ssa
```

Abre `compile.go` y busca `var passes` (alrededor de la linea 457). Veras:

```go
var passes = [...]pass{
	{name: "number lines", fn: numberLines, required: true},
	{name: "early phielim and copyelim", fn: copyelim},
	// ... many more passes
}
```

Cada pasada tiene:

- **name** - Se muestra en la salida de depuracion
- **fn** - Funcion que realiza la transformacion
- **required** - Si la pasada debe ejecutarse obligatoriamente

## Paso 2: Crear la Pasada de Deteccion de Potencias de Dos

Crea un nuevo archivo para nuestra pasada de deteccion:

```bash
cd go/src/cmd/compile/internal/ssa
```

**Crea `powoftwodetector.go`:**

```go
package ssa

import (
	"fmt"
	"math/bits"
)

func detectDivByPowerOfTwo(f *Func) {
	count := 0

	for _, b := range f.Blocks {
		for _, v := range b.Values {
			// Check for division operations
			if v.Op == OpDiv64 || v.Op == OpDiv32 || v.Op == OpDiv16 || v.Op == OpDiv8 ||
				v.Op == OpDiv64u || v.Op == OpDiv32u || v.Op == OpDiv16u || v.Op == OpDiv8u {

				// Check if the divisor (second argument) is a constant
				if len(v.Args) >= 2 {
					divisor := v.Args[1]

					// Check if it's a constant value
					if divisor.Op == OpConst64 || divisor.Op == OpConst32 ||
						divisor.Op == OpConst16 || divisor.Op == OpConst8 {

						constValue := divisor.AuxInt

						// Check if the constant is a power of two
						if isPowerOfTwo(constValue) {
							count++
							if f.pass.debug > 0 {
								fmt.Printf("  [PowerOfTwo] Found division by power of 2: %v / %d (could be >> %d) at %v\n",
									v.Args[0], constValue, bits.TrailingZeros64(uint64(constValue)), v.Pos)
							}
						}
					}
				}
			}
		}
	}

	if count > 0 {
		fmt.Printf("[PowerOfTwo Detector] Function %s: found %d division(s) by power of 2\n", f.Name, count)
	}
}
```

### Entendiendo el Codigo

- **`f *Func`** - La funcion SSA que se esta analizando
- **`f.Blocks`** - Todos los bloques basicos de la funcion
- **`b.Values`** - Todos los valores SSA (operaciones) en un bloque
- **`v.Op`** - El tipo de operacion (division, suma, etc.)
- **`v.Args`** - Los operandos de la operacion
- **`divisor.AuxInt`** - El valor de la constante
- **`isPowerOfTwo()`** - Funcion auxiliar que ya existe en `rewrite.go`
- **`bits.TrailingZeros64()`** - Calcula cuantos bits hay que desplazar

## Paso 3: Registrar la Pasada en el Compilador

**Edita `compile.go`:**

Busca el array `var passes` (alrededor de la linea 457) y anade tu pasada como la **primera** entrada:

```go
var passes = [...]pass{
	{name: "detect div by power of two", fn: detectDivByPowerOfTwo, required: true},
	{name: "number lines", fn: numberLines, required: true},
	// ... rest of the passes
```

Esto ejecuta tu detector al principio del pipeline, antes de que otras optimizaciones puedan eliminar la division.

## Paso 4: Recompilar el Compilador

```bash
cd go/src
./make.bash
```

Esto compila tu nueva pasada dentro del compilador de Go.

## Paso 5: Crear Programas de Prueba

Crea `test_divisions.go`:

```go
package main

func testDivisions() int {
	x := 100

	// These should be detected (powers of 2)
	a := x / 2   // 2 = 2^1, could be >> 1
	b := x / 4   // 4 = 2^2, could be >> 2
	c := x / 8   // 8 = 2^3, could be >> 3
	d := x / 16  // 16 = 2^4, could be >> 4

	// These should NOT be detected (not powers of 2)
	e := x / 3
	f := x / 5
	g := x / 7

	return a + b + c + d + e + f + g
}

func main() {
	result := testDivisions()
	println("Result:", result)
}
```

## Paso 6: Ejecutar y Ver la Deteccion

```bash
../go/bin/go build test_divisions.go
```

**Salida esperada:**
```
[PowerOfTwo Detector] Function main.testDivisions: found 4 division(s) by power of 2
```

Tu detector encontro las 4 divisiones por potencias de 2.

## Paso 7: Probar con Salida de Depuracion

Para obtener informacion detallada sobre cada deteccion:

```bash
GOSSAFUNC=testDivisions ../go/bin/go build -gcflags="-d=ssa/detect_div_by_power_of_two/debug=1" test_divisions.go
```

**Salida esperada:**
```
  [PowerOfTwo] Found division by power of 2: v10 / 2 (could be >> 1) at test_divisions.go:6
  [PowerOfTwo] Found division by power of 2: v14 / 4 (could be >> 2) at test_divisions.go:7
  [PowerOfTwo] Found division by power of 2: v18 / 8 (could be >> 3) at test_divisions.go:8
  [PowerOfTwo] Found division by power of 2: v22 / 16 (could be >> 4) at test_divisions.go:9
[PowerOfTwo Detector] Function main.testDivisions: found 4 division(s) by power of 2
```

Esto muestra las ubicaciones exactas y las cantidades de desplazamiento.

## Lo que Aprendimos

- **Arquitectura de Pasadas SSA**: Como crear y registrar pasadas del compilador
- **Recorrido SSA**: Como navegar por bloques y valores para analizar codigo
- **Deteccion de Operaciones**: Como identificar operaciones SSA especificas
- **Analisis vs Transformacion**: Nuestra pasada analiza pero no modifica (todavia)

## Ideas de Extension

Prueba estas mejoras adicionales:

1. **Implementar la optimizacion real**: Reemplazar las divisiones por desplazamientos
2. **Detectar multiplicaciones por potencias de 2**: Podrian usar desplazamientos a la izquierda
3. **Contar el total de optimizaciones**: Llevar la cuenta a lo largo de toda la compilacion
4. **Reportar ganancias de eficiencia**: Estimar el ahorro de ciclos por la optimizacion

## Limpieza

Para eliminar tu pasada personalizada:

```bash
cd go/src/cmd/compile/internal/ssa
rm powoftwodetector.go
# Edit compile.go and remove your pass from the passes array
cd ../../src
./make.bash
```

## Resumen

Has creado con exito una pasada personalizada del compilador SSA que detecta oportunidades de optimizacion.

```
Nombre de la pasada: "detect div by power of two"
Entrada:             Representacion SSA de la funcion
Analisis:            Encuentra operaciones x / (potencia de 2)
Salida:              Reporta optimizaciones potenciales
Ubicacion:           Al principio del pipeline del compilador

Ejemplo:             x / 8  →  Reporta: "could be >> 3"
```

Esto demuestra como la infraestructura del compilador de Go permite pasadas personalizadas de analisis y optimizacion. Las optimizaciones reales usan los mismos patrones, solo que modifican el SSA en lugar de limitarse a reportar.

---

*Continua con el [Ejercicio 7](07-runtime-patient-go.es.md) o vuelve al [taller principal](../README.md)*
