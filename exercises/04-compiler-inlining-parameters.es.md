# Ejercicio 4: Parametros de Inlining del Compilador - Ajuste para el Control del Tamanio del Binario

> 📖 **¿Quieres aprender mas?** Lee [The IR](https://internals-for-interns.com/posts/the-go-ir/) en Internals for Interns para profundizar en la representacion intermedia de Go, incluyendo como se toman las decisiones de inlining de funciones.

En este ejercicio, exploraras y modificaras los parametros de inlining de Go para ver sus efectos dramaticos en el tamanio del binario. Esto te ensenara como el compilador de Go decide cuando hacer inline de funciones y como ajustar estos parametros puede cambiar significativamente tus programas compilados.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender el sistema de presupuesto de inlining de Go y sus parametros
- Saber donde se toman las decisiones de inlining en el compilador
- Modificar los umbrales de inlining para controlar el comportamiento de optimizacion
- Medir el impacto en el tamanio del binario

## Contexto: Inlining de Funciones en Go

El inlining de funciones es una optimizacion del compilador donde las llamadas a funciones se reemplazan por el cuerpo real de la funcion. Esto intercambia tamanio del binario por rendimiento:

**Beneficios:**

- Elimina la sobrecarga de las llamadas
- Permite optimizaciones adicionales en el punto de llamada
- Mejor utilizacion del pipeline de instrucciones

**Costes:**

- Mayor tamanio del binario
- Mayor uso de memoria (para el programa)

Go utiliza un sofisticado **sistema de presupuesto** para decidir cuando el inlining es rentable.

## Paso 1: Entender el Presupuesto de Inlining de Go

Examinemos los parametros actuales de inlining:

```bash
cd go/src/cmd/compile/internal/inline
```

Abre `inl.go` y busca los parametros clave alrededor de las lineas 49-85:

### Parametros Clave de Inlining

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

**Nota:** `inlineHotMaxBudget` es una `var`, no una `const`, porque se usa con PGO (Profile Guided Optimization) y puede modificarse en tiempo de ejecucion.

### Como Funciona el Sistema de Presupuesto

Cada sentencia/expresion de Go tiene un **coste**:

- Sentencias simples: 1 punto
- Llamadas a funciones: 57+ puntos
- Bucles, condiciones: 1 punto cada uno
- Expresiones complejas: Puntos variables

El compilador suma los costes y los compara con el presupuesto.

## Paso 2: Usar el Binario del Compilador de Go para Comparar Tamanios

En lugar de crear programas de juguete, usemos el propio binario del compilador de Go como sujeto de prueba. El compilador de Go (`bin/go`) es perfecto para demostrar los efectos del inlining porque:

- **Base de codigo grande** - Muestra diferencias de tamanio significativas
- **Codigo del mundo real** - Contiene los patrones reales que estamos optimizando
- **Relevancia para el taller** - Lo estamos compilando a lo largo de los ejercicios
- **Resultados dramaticos** - Lo suficientemente grande para mostrar un impacto significativo del inlining

### Probar Diferentes Configuraciones de Inlining en el Binario de Go

Recompilemos toda la cadena de herramientas de Go con diferentes configuraciones de inlining y comparemos los tamanios del binario `bin/go`:

```bash
cd go/src
```

### Compilacion Base - Configuracion por Defecto

Primero, compilemos con la configuracion de inlining por defecto y hagamos una copia de seguridad del binario:

```bash
# Build with default settings
./make.bash

# Copy the default Go binary for comparison
cp ../bin/go ../bin/go-default

# Check the size
ls -lh ../bin/go-default
wc -c ../bin/go-default
```

### Verificar el Impacto Actual del Inlining en la Compilacion del Compilador de Go

Podemos examinar como el inlining afecta al propio compilador de Go durante la compilacion:

```bash
# See inlining decisions when compiling the Go compiler
# This shows how inlining parameters affect the compiler's own build process
cd cmd/compile
../../bin/go build -gcflags="-m" . 2>&1 | grep "can inline" | wc -l
echo "Functions that can be inlined during Go compiler build"
```

## Paso 3: Modificar los Parametros de Inlining

Ahora modifiquemos los parametros de inlining para ver sus efectos.

### Experimento 1: Inlining Agresivo

Edita `go/src/cmd/compile/internal/inline/inl.go` alrededor de la linea 50:

```go
const (
    inlineMaxBudget       = 95    // Increased from 80
    inlineExtraCallCost   = 40    // Decreased from 57
    inlineBigFunctionMaxCost = 30 // Increased from 20
)
```

> **⚠️ Nota:** Ten cuidado de no aumentar estos valores demasiado. En Go 1.26.1, el runtime tiene restricciones estrictas de write barrier, y aumentar el presupuesto de inlining mas alla de ~95 hace que el compilador haga inline de funciones en contextos donde las write barriers estan prohibidas, rompiendo la compilacion. Esto en si mismo es una gran leccion sobre el delicado equilibrio de los parametros del compilador.

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

Ahora prueba con configuraciones conservadoras. Edita los parametros:

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

## Paso 4: Analisis Exhaustivo del Tamanio del Binario

Probemos configuraciones extremas de inlining para ver efectos dramaticos en el binario del compilador de Go:

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

### Experimento 4: Inlining Extremo - Demostracion del Punto de Ruptura

Probemos configuraciones extremadamente agresivas para ver que pasa cuando llevamos el inlining demasiado lejos:

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

**⚠️ Resultado esperado:** Esto fallara al compilar. Veras errores de "write barrier prohibited by caller". Esto ocurre porque el compilador hace inline de funciones del runtime en contextos donde las write barriers no estan permitidas, creando cadenas de llamadas ilegales.

Si falla (que es lo esperado), aprenderas que:
- El inlining extremo causa violaciones de write barrier en el runtime
- El runtime de Go tiene anotaciones `//go:nowritebarrierrec` que prohiben write barriers en ciertas cadenas de llamadas
- Cuando el inlining expone estas cadenas, el compilador rechaza correctamente la compilacion
- Los parametros por defecto estan cuidadosamente equilibrados por una buena razon

## Paso 5: Analizar los Resultados

Compara los tamanios del binario del compilador de Go:

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


## Que Hemos Modificado

### Funciones de los Parametros Clave

| Parametro | Proposito | Impacto |
|-----------|-----------|---------|
| `inlineMaxBudget` | Coste maximo para cualquier funcion inlined | Mayor = mas inlining |
| `inlineExtraCallCost` | Penalizacion por llamadas a funciones dentro de funciones inlined | Menor = mas agresivo |
| `inlineBigFunctionMaxCost` | Coste maximo al hacer inline en funciones grandes | Mayor = mas inlining en funciones grandes |
| `inlineBigFunctionNodes` | Umbral para la deteccion de funciones "grandes" | Menor = mas funciones consideradas "grandes" |

### Resultados Tipicos que Deberias Observar

Con el binario del compilador de Go, deberias observar diferencias de tamanio notables:

- **Sin Inlining**: Binario mas pequenio
- **Conservador**: Ligeramente mas pequenio que el por defecto
- **Por defecto**: Tamanio equilibrado
- **Agresivo**: Binario mas grande que el por defecto

**Ideas clave:**

- Incluso cambios modestos en los parametros de inlining producen diferencias medibles en el tamanio del binario
- El rango desde sin inlining hasta agresivo muestra el impacto de esta optimizacion
- Los valores mas agresivos estan limitados por restricciones del runtime (write barriers)

Los tamanios exactos dependen de tu sistema, pero deberias ver diferencias dramaticas similares.

## Lo que Aprendimos

- **Sistema de Presupuesto**: Como Go utiliza analisis basado en costes para las decisiones de inlining
- **Impacto de los Parametros**: Como diferentes configuraciones afectan el tamanio del binario y el rendimiento
- **Tecnicas de Medicion**: Uso de flags de depuracion para entender las decisiones del compilador
- **Compromisos**: La tension fundamental entre tamanio del binario y rendimiento
- **Ajuste del Compilador**: Como modificar el comportamiento del compilador para necesidades especificas

## Ideas de Extension

Prueba estos experimentos adicionales:

1. Crear un script para automatizar las pruebas con diferentes combinaciones de parametros
2. Probar con programas Go del mundo real (como compilar el propio Go)
3. Medir las diferencias en tiempo de compilacion con varias configuraciones
4. Experimentar con los parametros de PGO (Profile-Guided Optimization)
5. Analizar las diferencias en la salida de ensamblador entre llamadas con y sin inline

## Siguientes Pasos

Has aprendido como ajustar el comportamiento de inlining de Go y has visto su impacto real en el tamanio del binario y el rendimiento. En los proximos ejercicios, exploraremos la modificacion de la herramienta gofmt.

## Limpieza

Para restaurar los parametros originales de inlining y limpiar los binarios de prueba:

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

1. **El Inlining es un Compromiso**: Mas inlining = binarios mas grandes pero potencialmente ejecucion mas rapida
2. **Sistema de Presupuesto**: Go utiliza un sofisticado analisis de costes para tomar decisiones de inlining
3. **Impacto de los Parametros**: Pequenios cambios en los parametros pueden tener efectos significativos en la salida
4. **Herramientas de Depuracion**: Go proporciona excelentes herramientas para entender las decisiones del compilador
5. **Relevancia en el Mundo Real**: Estos parametros afectan a cada programa Go que compilas

El equipo del compilador de Go ha ajustado cuidadosamente estos valores por defecto mediante pruebas de rendimiento exhaustivas, pero ahora entiendes como ajustarlos para tus necesidades especificas.

---

*Continua con el [Ejercicio 5](05-gofmt-ast-transformation.es.md) o vuelve al [taller principal](../README.md)*
