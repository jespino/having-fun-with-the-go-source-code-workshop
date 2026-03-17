# Ejercicio 6: Pasada SSA - Detección de Divisiones por Potencias de Dos

> 📖 **¿Quieres aprender más?** Lee [The SSA Phase](https://internals-for-interns.com/es/posts/the-go-ssa/) en Internals for Interns para profundizar en las pasadas de optimización SSA de Go.

En este ejercicio, aprenderás cómo funcionan las pasadas del compilador SSA (Static Single Assignment) de Go creando una pasada de optimización personalizada que detecta operaciones de división por potencias de dos.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Entender la arquitectura de pasadas del compilador SSA de Go
- Saber cómo recorrer bloques y valores SSA
- Crear una pasada de análisis personalizada desde cero
- Integrar tu pasada en el pipeline del compilador
- Usar los volcados SSA para verificar que tu pasada funciona

## Introducción: ¿Qué es SSA?

**Static Single Assignment (SSA)** es una representación del compilador donde cada variable se asigna exactamente una vez. En lugar de reutilizar variables como `x = 1; x = x + 2`, SSA crea nuevas versiones: `x1 = 1; x2 = x1 + 2`. Esta restricción elimina la ambigüedad — al analizar un valor, el compilador sabe que exactamente una definición lo produjo, lo que permite optimizaciones potentes.

El código SSA se organiza en dos estructuras clave:

- **Values (Valores)**: Computaciones individuales como `v3 = Add64 v1 v2`. Cada valor tiene una operación (Add64, Div32, Const64...), un tipo y referencias a sus entradas.
- **Blocks (Bloques)**: Secuencias de valores sin ramificaciones en el medio. Los bloques están conectados por aristas de flujo de control, formando el grafo de flujo de la función.

Cuando los caminos de flujo de control se fusionan (como después de un `if/else`), SSA usa **nodos PHI** para reconciliar diferentes valores: `v5 = Phi v3 v4` significa "v5 es v3 o v4, dependiendo de qué rama se ejecutó".

El compilador ejecuta más de 30 **pasadas** sobre el grafo SSA en secuencia. Cada pasada recorre los bloques y valores para analizarlos o transformarlos. Las pasadas se ejecutan antes y después del **lowering** — el paso que convierte las operaciones genéricas (como `Add64`) en instrucciones específicas de la arquitectura (como `AMD64ADDQ`).

## Contexto: Pasadas del Compilador SSA

El compilador de Go transforma tu código a través de múltiples pasadas:

1. **Parseo** - Convertir el código fuente a AST
2. **Verificación de Tipos** - Comprobar que los tipos son correctos
3. **Generación de IR** - Convertir a forma IR (representación intermedia)
3. **Generación de SSA** - Convertir a forma SSA (Static Single Assignment)
4. **Pasadas de Optimización** - Transformar el SSA (nuestro enfoque)
5. **Generación de Código** - Producir código máquina

Vamos a trabajar con la forma SSA para conocer la posibilidad de optimizar potencias de dos.

## Paso 1: Entender la Estructura de las Pasadas SSA

Las pasadas SSA se registran en `compile.go` y operan sobre funciones. Examinemos la estructura:

```bash
cd go/src/cmd/compile/internal/ssa
```

Abre `compile.go` y busca `var passes` (alrededor de la línea 457). Verás:

```go
var passes = [...]pass{
	{name: "number lines", fn: numberLines, required: true},
	{name: "early phielim and copyelim", fn: copyelim},
	// ... many more passes
}
```

Cada pasada tiene:

- **name** - Se muestra en la salida de depuración
- **fn** - Función que realiza la transformación
- **required** - Si la pasada debe ejecutarse obligatoriamente

## Paso 2: Crear la Pasada de Detección de Potencias de Dos

Crea un nuevo archivo para nuestra pasada de detección:

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

### Entendiendo el Código

- **`f *Func`** - La función SSA que se está analizando
- **`f.Blocks`** - Todos los bloques básicos de la función
- **`b.Values`** - Todos los valores SSA (operaciones) en un bloque
- **`v.Op`** - El tipo de operación (división, suma, etc.)
- **`v.Args`** - Los operandos de la operación
- **`divisor.AuxInt`** - El valor de la constante
- **`isPowerOfTwo()`** - Función auxiliar que ya existe en `rewrite.go`
- **`bits.TrailingZeros64()`** - Calcula cuántos bits hay que desplazar

## Paso 3: Registrar la Pasada en el Compilador

**Edita `compile.go`:**

Busca el array `var passes` (alrededor de la línea 457) y añade tu pasada como la **primera** entrada:

```go
var passes = [...]pass{
	{name: "detect div by power of two", fn: detectDivByPowerOfTwo, required: true},
	{name: "number lines", fn: numberLines, required: true},
	// ... rest of the passes
```

Esto ejecuta tu detector al principio del pipeline, antes de que otras optimizaciones puedan eliminar la división.

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

## Paso 6: Ejecutar y Ver la Detección

```bash
../go/bin/go build test_divisions.go
```

**Salida esperada:**
```
[PowerOfTwo Detector] Function main.testDivisions: found 4 division(s) by power of 2
```

Tu detector encontró las 4 divisiones por potencias de 2.

## Paso 7: Probar con Salida de Depuración

Para obtener información detallada sobre cada detección:

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

- **Arquitectura de Pasadas SSA**: Cómo crear y registrar pasadas del compilador
- **Recorrido SSA**: Cómo navegar por bloques y valores para analizar código
- **Detección de Operaciones**: Cómo identificar operaciones SSA específicas
- **Análisis vs Transformación**: Nuestra pasada analiza pero no modifica (¡todavía!)

## Ideas de Extensión

Prueba estas mejoras adicionales:

1. **Implementar la optimización real**: Reemplazar las divisiones por desplazamientos
2. **Detectar multiplicaciones por potencias de 2**: Podrían usar desplazamientos a la izquierda
3. **Contar el total de optimizaciones**: Llevar la cuenta a lo largo de toda la compilación
4. **Reportar ganancias de eficiencia**: Estimar el ahorro de ciclos por la optimización

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

¡Has creado con éxito una pasada personalizada del compilador SSA que detecta oportunidades de optimización!

```
Nombre de la pasada: "detect div by power of two"
Entrada:             Representación SSA de la función
Análisis:            Encuentra operaciones x / (potencia de 2)
Salida:              Reporta optimizaciones potenciales
Ubicación:           Al principio del pipeline del compilador

Ejemplo:             x / 8  →  Reporta: "could be >> 3"
```

Esto demuestra cómo la infraestructura del compilador de Go permite pasadas personalizadas de análisis y optimización. Las optimizaciones reales usan los mismos patrones, solo que modifican el SSA en lugar de limitarse a reportar.

---

*Continúa con el [Ejercicio 7](07-runtime-patient-go.es.md) o vuelve al [taller principal](../README.md)*
