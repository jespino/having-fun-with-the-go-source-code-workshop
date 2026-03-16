# Ejercicio 4: Parámetros de Inlining del Compilador - Ajuste para el Control del Tamaño del Binario

> 📖 **¿Quieres aprender más?** Lee [The IR](https://internals-for-interns.com/es/posts/the-go-ir/) en Internals for Interns para profundizar en la representación intermedia de Go, incluyendo cómo se toman las decisiones de inlining de funciones.

En este ejercicio, explorarás y modificarás los parámetros de inlining de Go para ver sus efectos dramáticos en el tamaño del binario. Esto te enseñará cómo el compilador de Go decide cuándo hacer inline de funciones y cómo ajustar estos parámetros puede cambiar significativamente tus programas compilados.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Entender el sistema de presupuesto de inlining de Go y sus parámetros
- Saber dónde se toman las decisiones de inlining en el compilador
- Modificar los umbrales de inlining para controlar el comportamiento de optimización
- Medir el impacto en el tamaño del binario

## Contexto: Inlining de Funciones en Go

El inlining de funciones es una optimización del compilador donde las llamadas a funciones se reemplazan por el cuerpo real de la función. Esto intercambia tamaño del binario por rendimiento:

**Beneficios:**

- Elimina la sobrecarga de las llamadas
- Permite optimizaciones adicionales en el punto de llamada
- Mejor utilización del pipeline de instrucciones

**Costes:**

- Mayor tamaño del binario
- Mayor uso de memoria (para el programa)

Go utiliza un sofisticado **sistema de presupuesto** para decidir cuándo el inlining es rentable.

## Paso 1: Entender el Presupuesto de Inlining de Go

Examinemos los parámetros actuales de inlining:

```bash
cd go/src/cmd/compile/internal/inline
```

Abre `inl.go` y busca los parámetros clave alrededor de las líneas 49-85:

### Parámetros Clave de Inlining

De `go/src/cmd/compile/internal/inline/inl.go:49-85`:

```go
const (
    inlineMaxBudget       = 80
    inlineExtraAppendCost = 0
    inlineExtraCallCost   = 57              // benchmarked to provide most benefit
    inlineParamCallCost   = 17              // calling a parameter costs less
    inlineExtraPanicCost  = 1               // do not penalize inlining panics
    inlineExtraThrowCost  = inlineMaxBudget // inlining runtime.throw does not help

    inlineBigFunctionNodes      = 5000                 // Functions with this many nodes are "big"
    inlineBigFunctionMaxCost    = 20                   // Max cost when inlining into a "big" function
    inlineClosureCalledOnceCost = 10 * inlineMaxBudget // if a closure is called once, inline it
)

var (
    // ...
    // Budget increased due to hotness (PGO).
    inlineHotMaxBudget int32 = 2000
)
```

**Nota:** `inlineHotMaxBudget` es una `var`, no una `const`, porque se usa con PGO (Profile Guided Optimization) y puede modificarse en tiempo de ejecución.

### Cómo Funciona el Sistema de Presupuesto

Cada sentencia/expresión de Go tiene un **coste**:

- Sentencias simples: 1 punto
- Llamadas a funciones: 57+ puntos
- Bucles, condiciones: 1 punto cada uno
- Expresiones complejas: Puntos variables

El compilador suma los costes y los compara con el presupuesto.

## Paso 2: Usar el Binario del Compilador de Go para Comparar Tamaños

En lugar de crear programas de juguete, usemos el propio binario del compilador de Go como sujeto de prueba. El compilador de Go (`bin/go`) es perfecto para demostrar los efectos del inlining porque:

- **Base de código grande** - Muestra diferencias de tamaño significativas
- **Código del mundo real** - Contiene los patrones reales que estamos optimizando
- **Relevancia para el taller** - Lo estamos compilando a lo largo de los ejercicios
- **Resultados dramáticos** - Lo suficientemente grande para mostrar un impacto significativo del inlining

### Probar Diferentes Configuraciones de Inlining en el Binario de Go

Recompilemos toda la cadena de herramientas de Go con diferentes configuraciones de inlining y comparemos los tamaños del binario `bin/go`:

```bash
cd go/src
```

### Compilación Base - Configuración por Defecto

Primero, compilemos con la configuración de inlining por defecto y hagamos una copia de seguridad del binario:

```bash
# Build with default settings
./make.bash

# Copy the default Go binary for comparison
cp ../bin/go ../bin/go-default

# Check the size
ls -lh ../bin/go-default
wc -c ../bin/go-default
```

### Verificar el Impacto Actual del Inlining en la Compilación del Compilador de Go

Podemos examinar cómo el inlining afecta al propio compilador de Go durante la compilación:

```bash
# See inlining decisions when compiling the Go compiler
# This shows how inlining parameters affect the compiler's own build process
cd cmd/compile
../../bin/go build -gcflags="-m" . 2>&1 | grep "can inline" | wc -l
echo "Functions that can be inlined during Go compiler build"
```

## Paso 3: Modificar los Parámetros de Inlining

¡Ahora modifiquemos los parámetros de inlining para ver sus efectos!

### Experimento 1: Inlining Agresivo

Edita `go/src/cmd/compile/internal/inline/inl.go` alrededor de la línea 50:

```go
const (
    inlineMaxBudget       = 95    // Increased from 80
    inlineExtraCallCost   = 40    // Decreased from 57
    inlineBigFunctionMaxCost = 30 // Increased from 20
)
```

> **⚠️ Nota:** ¡Ten cuidado de no aumentar estos valores demasiado! En Go 1.26.1, el runtime tiene restricciones estrictas de write barrier, y aumentar el presupuesto de inlining más allá de ~95 hace que el compilador haga inline de funciones en contextos donde las write barriers están prohibidas, rompiendo la compilación. Esto en sí mismo es una gran lección sobre el delicado equilibrio de los parámetros del compilador.

**Recompila el compilador:**

```bash
cd go/src
./make.bash
```

**Prueba el inlining agresivo en el binario de Go:**

```bash
# Copy the aggressively-inlined Go binary
cp ../bin/go ../bin/go-aggressive

# Compare sizes
echo "Default size: $(wc -c < ../bin/go-default)"
echo "Aggressive size: $(wc -c < ../bin/go-aggressive)"

# Calculate size difference
default_size=$(wc -c < ../bin/go-default)
aggressive_size=$(wc -c < ../bin/go-aggressive)
echo "Size difference: $(($aggressive_size - $default_size)) bytes"
echo "Percentage increase: $(echo "scale=2; ($aggressive_size - $default_size) * 100 / $default_size" | bc)%"
```

### Experimento 2: Inlining Conservador

Ahora prueba con configuraciones conservadoras. Edita los parámetros:

```go
const (
    inlineMaxBudget       = 40    // Decreased from 80
    inlineExtraCallCost   = 100   // Increased from 57
    inlineBigFunctionMaxCost = 5  // Decreased from 20
)
```

**Recompila y prueba:**

```bash
cd go/src
./make.bash

# Copy the conservatively-inlined Go binary
cp ../bin/go ../bin/go-conservative

# Compare all three Go binaries
echo "Conservative size: $(wc -c < ../bin/go-conservative)"
echo "Default size: $(wc -c < ../bin/go-default)"
echo "Aggressive size: $(wc -c < ../bin/go-aggressive)"
```

## Paso 4: Análisis Exhaustivo del Tamaño del Binario

Probemos configuraciones extremas de inlining para ver efectos dramáticos en el binario del compilador de Go:

### Experimento 3: Sin Inlining

Para comparar, desactivemos el inlining por completo:

```go
const (
    inlineMaxBudget       = 0     // No inlining budget
    inlineExtraCallCost   = 1000  // Prohibitive call cost
    inlineBigFunctionMaxCost = 0  // No big function inlining
)
```

```bash
cd go/src
./make.bash

# Copy the no-inlining Go binary
cp ../bin/go ../bin/go-no-inline
```

### Experimento 4: Inlining Extremo - Demostración del Punto de Ruptura

Probemos configuraciones extremadamente agresivas para ver qué pasa cuando llevamos el inlining demasiado lejos:

```go
const (
    inlineMaxBudget       = 500   // Very high budget
    inlineExtraCallCost   = 5     // Very low call cost
    inlineBigFunctionMaxCost = 200 // Very high big function budget
)
```

```bash
cd go/src
./make.bash
```

**⚠️ Resultado esperado:** ¡Esto fallará al compilar! Verás errores de "write barrier prohibited by caller". Esto ocurre porque el compilador hace inline de funciones del runtime en contextos donde las write barriers no están permitidas, creando cadenas de llamadas ilegales.

Si falla (que es lo esperado), aprenderás que:
- El inlining extremo causa violaciones de write barrier en el runtime
- El runtime de Go tiene anotaciones `//go:nowritebarrierrec` que prohíben write barriers en ciertas cadenas de llamadas
- Cuando el inlining expone estas cadenas, el compilador rechaza correctamente la compilación
- Los parámetros por defecto están cuidadosamente equilibrados por una buena razón

## Paso 5: Analizar los Resultados

Compara los tamaños del binario del compilador de Go:

```bash
cd go

echo "=== GO COMPILER BINARY SIZE COMPARISON ==="
echo "No Inlining:  $(wc -c < bin/go-no-inline) bytes"
echo "Conservative: $(wc -c < bin/go-conservative) bytes"
echo "Default:      $(wc -c < bin/go-default) bytes"
echo "Aggressive:   $(wc -c < bin/go-aggressive) bytes"

echo ""
echo "=== SIZE DIFFERENCES ==="
no_inline_size=$(wc -c < bin/go-no-inline)
conservative_size=$(wc -c < bin/go-conservative)
default_size=$(wc -c < bin/go-default)
aggressive_size=$(wc -c < bin/go-aggressive)

echo "No-inline vs Default: $(($default_size - $no_inline_size)) bytes difference"
echo "Default vs Aggressive: $(($aggressive_size - $default_size)) bytes difference"
echo "Full Range (No-inline to Aggressive): $(($aggressive_size - $no_inline_size)) bytes difference"

# Calculate percentages
echo ""
echo "=== PERCENTAGE DIFFERENCES ==="
echo "Aggressive vs Default: $(echo "scale=2; ($aggressive_size - $default_size) * 100 / $default_size" | bc)%"
echo "Default vs No-inline: $(echo "scale=2; ($default_size - $no_inline_size) * 100 / $no_inline_size" | bc)%"
```


## Qué Hemos Modificado

### Funciones de los Parámetros Clave

| Parámetro | Propósito | Impacto |
|-----------|-----------|---------|
| `inlineMaxBudget` | Coste máximo para cualquier función inlined | Mayor = más inlining |
| `inlineExtraCallCost` | Penalización por llamadas a funciones dentro de funciones inlined | Menor = más agresivo |
| `inlineBigFunctionMaxCost` | Coste máximo al hacer inline en funciones grandes | Mayor = más inlining en funciones grandes |
| `inlineBigFunctionNodes` | Umbral para la detección de funciones "grandes" | Menor = más funciones consideradas "grandes" |

### Resultados Típicos que Deberías Observar

Con el binario del compilador de Go, deberías observar diferencias de tamaño notables:

- **Sin Inlining**: Binario más pequeño
- **Conservador**: Ligeramente más pequeño que el por defecto
- **Por defecto**: Tamaño equilibrado
- **Agresivo**: Binario más grande que el por defecto

**Ideas clave:**

- Incluso cambios modestos en los parámetros de inlining producen diferencias medibles en el tamaño del binario
- El rango desde sin inlining hasta agresivo muestra el impacto de esta optimización
- Los valores más agresivos están limitados por restricciones del runtime (write barriers)

Los tamaños exactos dependen de tu sistema, pero deberías ver diferencias dramáticas similares.

## Lo que Aprendimos

- **Sistema de Presupuesto**: Cómo Go utiliza análisis basado en costes para las decisiones de inlining
- **Impacto de los Parámetros**: Cómo diferentes configuraciones afectan el tamaño del binario y el rendimiento
- **Técnicas de Medición**: Uso de flags de depuración para entender las decisiones del compilador
- **Compromisos**: La tensión fundamental entre tamaño del binario y rendimiento
- **Ajuste del Compilador**: Cómo modificar el comportamiento del compilador para necesidades específicas

## Ideas de Extensión

Prueba estos experimentos adicionales:

1. Crear un script para automatizar las pruebas con diferentes combinaciones de parámetros
2. Probar con programas Go del mundo real (¡como compilar el propio Go!)
3. Medir las diferencias en tiempo de compilación con varias configuraciones
4. Experimentar con los parámetros de PGO (Profile-Guided Optimization)
5. Analizar las diferencias en la salida de ensamblador entre llamadas con y sin inline

## Siguientes Pasos

Has aprendido cómo ajustar el comportamiento de inlining de Go y has visto su impacto real en el tamaño del binario y el rendimiento. En los próximos ejercicios, exploraremos la modificación de la herramienta gofmt.

## Limpieza

Para restaurar los parámetros originales de inlining y limpiar los binarios de prueba:

```bash
cd go/src/cmd/compile/internal/inline
git checkout inl.go
cd ../../../../

# Rebuild with original parameters
cd src
./make.bash

# Clean up test binaries
rm -f ../bin/go-default ../bin/go-aggressive ../bin/go-conservative ../bin/go-no-inline
```

## Conclusiones Clave

1. **El Inlining es un Compromiso**: Más inlining = binarios más grandes pero potencialmente ejecución más rápida
2. **Sistema de Presupuesto**: Go utiliza un sofisticado análisis de costes para tomar decisiones de inlining
3. **Impacto de los Parámetros**: Pequeños cambios en los parámetros pueden tener efectos significativos en la salida
4. **Herramientas de Depuración**: Go proporciona excelentes herramientas para entender las decisiones del compilador
5. **Relevancia en el Mundo Real**: Estos parámetros afectan a cada programa Go que compilas

El equipo del compilador de Go ha ajustado cuidadosamente estos valores por defecto mediante pruebas de rendimiento exhaustivas, pero ahora entiendes cómo ajustarlos para tus necesidades específicas.

---

*Continúa con el [Ejercicio 5](05-gofmt-ast-transformation.es.md) o vuelve al [taller principal](../README.md)*
