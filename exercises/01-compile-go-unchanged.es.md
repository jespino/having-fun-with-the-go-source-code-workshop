# Ejercicio 1: Compilar Go sin Cambios

En este ejercicio, aprenderas a compilar la toolchain de Go desde el codigo fuente sin realizar ninguna modificacion. Esta es una habilidad esencial antes de empezar a hacer cambios en el lenguaje.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Entender el proceso de compilacion de Go y el concepto de bootstrap
- Compilar Go desde el codigo fuente exitosamente
- Saber como explorar la estructura del codigo fuente de Go
- Saber como probar tu compilacion personalizada de Go

## Paso 1: Entender el Proceso de Bootstrap

Go esta escrito en Go. Esto crea un problema del "huevo y la gallina": como compilas Go sin tener Go? La solucion es el bootstrapping:

1. El equipo de Go proporciona binarios precompilados
2. Estos binarios compilan el codigo fuente actual de Go
3. La version recien compilada se puede usar para el desarrollo

Verifiquemos que tienes Go instalado (necesario para el bootstrapping):

```bash
go version
# Debe mostrar version 1.24 o superior
```

**⚠️ Importante**: Debes tener Go 1.24 o superior instalado para compilar Go 1.26.1. Si no tienes Go instalado o tu version es demasiado antigua, instala la ultima version desde <https://golang.org/dl/>

## Paso 2: Navegar al Directorio del Codigo Fuente de Go

```bash
cd go/src
pwd
# Deberia mostrar: /ruta/al/taller/go/src

# Verifica que estas en la version correcta de Go
git describe --tags
# Deberia mostrar: go1.26.1
```

## Paso 3: Iniciar el Proceso de Compilacion

Go proporciona diferentes scripts para la compilacion. Empezaremos con `make.bash` que compila la toolchain, y mientras tanto exploraremos el codigo fuente.

**En sistemas tipo Unix (Linux, macOS):**

```bash
./make.bash
```

**En Windows:**

```cmd
make.bat
```

Este script hara lo siguiente:

1. Compilar la toolchain de Go (compilador, enlazador, runtime, biblioteca estandar)
2. Tardara aproximadamente entre 2 y 10 minutos dependiendo de tu sistema

**Nota:** La primera vez que ejecutes esto, tardara mas ya que necesita compilar todo desde cero.

### Que pasa con `all.bash` y `run.bash`?

Quizas te preguntes sobre otros scripts en el directorio `src/`:

- **`make.bash`**: Compila solo la toolchain de Go (lo que estamos usando)
- **`run.bash`**: Ejecuta la suite completa de tests (requiere que Go este compilado primero)
- **`all.bash`**: Script de conveniencia que ejecuta `make.bash` + `run.bash` + muestra informacion de la compilacion

Para este taller, `make.bash` es perfecto porque:

- Un tiempo de compilacion mas corto significa menos espera
- Solo necesitamos una compilacion funcional de Go para nuestros experimentos
- Podemos ejecutar los tests mas adelante si es necesario con `run.bash`

## Paso 4: Explorar el Codigo Fuente Mientras se Compila

Mientras la compilacion esta en marcha, abre una **nueva terminal** o **IDE** y exploremos la estructura del codigo fuente de Go. Este es un buen momento para entender lo que estamos compilando.

**En tu nueva terminal:**

```bash
cd /path/to/workshop/go  # Navega al directorio del codigo fuente de Go
ls -la
```

### Estructura del Repositorio

Directorios clave que deberias ver:

- **`src/`**: Contiene el codigo fuente de Go
  - `src/cmd/`: Herramientas de linea de comandos (go, gofmt, etc.)
  - `src/runtime/`: Sistema de runtime de Go
  - `src/go/`: Paquetes del lenguaje Go (parser, AST, etc.)
- **`test/`**: Archivos de test del lenguaje Go
- **`api/`**: Datos de compatibilidad de la API
- **`doc/`**: Documentacion

### Examinar la Estructura del Compilador de Go

El compilador de Go se encuentra en `src/cmd/compile/`. Vamos a explorarlo:

```bash
cd src/cmd/compile
ls -la
```

Archivos y directorios clave:

- **`main.go`**: Punto de entrada del compilador
- **`internal/`**: Paquetes internos del compilador
  - `internal/syntax/`: Lexer/parser (scanner, parser)
  - `internal/types2/`: Verificador de tipos
  - `internal/ir/`: Representacion intermedia
  - `internal/gc/`: Generacion de codigo

## Paso 5: Entender la Salida de la Compilacion

**Vuelve a tu terminal original** donde la compilacion esta en ejecucion. A medida que avanza la compilacion, deberias ver una salida como esta:

```
Building Go cmd/dist using /usr/local/go. (go1.26.1 darwin/amd64)
Building Go toolchain1 using /usr/local/go.
Building Go bootstrap cmd/go (go_bootstrap) using Go toolchain1.
Building Go toolchain2 using go_bootstrap and Go toolchain1.
Building Go toolchain3 using go_bootstrap and Go toolchain2.
Building packages and commands for darwin/amd64.
```

Esto muestra el proceso de bootstrap en multiples etapas:

- El compilador se compila con la version de Go instalada en tu sistema (toolchain1)
- Luego el compilador se vuelve a compilar usando la toolchain1 para producir la toolchain2
- Finalmente la toolchain3 se genera usando la toolchain2
- La toolchain3 y la toolchain2 deberian ser identicas

## Paso 6: Localizar tu Binario de Go Compilado

Despues de una compilacion exitosa, tu nuevo binario de Go estara en:

```bash
ls -la /path/to/workshop/go/bin
```

Deberias ver:

- `go` - El comando principal de Go
- `gofmt` - Formateador de Go
- Otras herramientas de Go

## Paso 7: Probar tu Compilacion Personalizada de Go

Probemos tu Go recien compilado:

```bash
# Verificar la version de tu Go compilado
../bin/go version
```

Crea un archivo hello.go en un directorio temporal, por ejemplo `/tmp`.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello from my custom Go build!")
}
```

```bash
# Compilar y ejecutar con tu Go personalizado
/path/to/workshop/go/bin/go run /tmp/hello.go
```

## ⚠️ Solucion de Problemas

### Interferencia de GOROOT

Si al ejecutar `../bin/go run /tmp/hello.go` (o la ruta completa al binario) obtienes resultados inesperados o se usa el Go del sistema en lugar del que acabas de compilar, puede que necesites eliminar la variable de entorno `GOROOT` primero:

```bash
unset GOROOT
/path/to/workshop/go/bin/go run /tmp/hello.go
```

Esto ocurre porque `GOROOT` puede estar configurada por la instalacion de Go de tu sistema, apuntando al nuevo binario hacia la biblioteca estandar y herramientas incorrectas. Al eliminarla, el binario detecta automaticamente su propio directorio raiz basandose en su ubicacion.

## Lo que Aprendimos

- **Proceso de Bootstrap**: Go se compila a si mismo usando una instalacion existente de Go
- **Estructura del Codigo Fuente de Go**: Un codigo bien organizado con separacion clara (cmd/, runtime/, etc.)
- **Proceso de Compilacion**: `./make.bash` compila todo

## Siguientes Pasos

Felicidades. Ahora tienes una toolchain de Go funcional compilada desde el codigo fuente.

Puedes continuar con cualquiera de los siguientes ejercicios para aprender sobre diferentes partes de Go:

- [Ejercicio 2: Anadir el Operador Flecha "=>" para Goroutines](./02-scanner-arrow-operator.es.md) - Modificaciones del scanner
- [Ejercicio 3: Multiples palabras clave "go" - Mejora del Parser](./03-parser-multiple-go.es.md) - Modificaciones del parser
- [Ejercicio 4: Parametros de Inlining - Experimentos de Inlining de Funciones](./04-compiler-inlining-parameters.md) - Parametros del compilador
- [Ejercicio 5: Transformacion de gofmt - "hello" a "helo"](./05-gofmt-ast-transformation.md) - Transformaciones del AST
- [Ejercicio 6: Pase SSA - Deteccion de Division por Potencias de Dos](./06-ssa-power-of-two-detector.md) - Pases del compilador SSA
- [Ejercicio 7: Go Paciente - Hacer que Go Espere a las Goroutines](./07-runtime-patient-go.md) - Modificaciones del runtime
- [Ejercicio 8: Detective de Goroutines Dormidas - Monitoreo del Estado del Runtime](./08-goroutine-sleep-detective.md) - Monitoreo del scheduler
- [Ejercicio 9: Select Predecible - Eliminar la Aleatoriedad del Select de Go](./09-predictable-select.md) - Comportamiento del select
- [Ejercicio 10: Stack Traces al Estilo Java - Hacer que los Panics de Go se Vean Familiares](./10-java-style-stack-traces.md) - Formato de errores

O vuelve al [taller principal](../README.md) para elegir un ejercicio.
