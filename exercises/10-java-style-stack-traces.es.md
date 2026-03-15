# Ejercicio 10: Stack Traces Estilo Java - Haciendo que los Panics de Go Resulten Familiares

En este ejercicio, modificarás el formato de los stack traces de Go para que se parezcan al estilo de Java. En lugar de los stack traces de Go, crearemos trazas al estilo Java.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Comprender cómo Go formatea los stack traces en el runtime
- Saber dónde se generan los mensajes de panic
- Modificar el formato de salida del runtime

## Contexto: Estilos de Stack Trace

Vamos a transformar el formato de stack trace de Go:

```
panic: Something went wrong

goroutine 1 [running]:
main.methodC()
        /Users/dev/project/main.go:15 +0x45
main.methodB()
        /Users/dev/project/main.go:11 +0x23
main.methodA()
        /Users/dev/project/main.go:7 +0x12
```

En este formato estilo Java:

```
Exception in thread "main" go.runtime.Panic: Something went wrong
    at main.methodC(main.go:15)
    at main.methodB(main.go:11)
    at main.methodA(main.go:7)
```

## Paso 1: Crear un Programa de Prueba

Crea un archivo `stack_trace_demo.go`:

```go
package main

import "fmt"

func methodC() {
    panic("Something went wrong")
}

func methodB() {
    methodC()
}

func methodA() {
    methodB()
}

func main() {
    fmt.Println("Starting the program...")
    methodA()
}
```

Ejecuta con el Go actual para ver el formato del stack trace:

```bash
go run stack_trace_demo.go
```

## Paso 2: Navegar a los Archivos del Runtime

```bash
cd go/src/runtime
```

Archivos clave que modificaremos:
- **`panic.go`** - Cabecera del mensaje de panic
- **`traceback.go`** - Formato de los frames del stack

## Paso 3: Modificar la Cabecera del Panic

**Edita `panic.go`:**

Encuentra la función `printpanics` alrededor de la línea 734. Busca:

```go
print("panic: ")
printpanicval(p.arg)
```

Cámbialo a:

```go
print("Exception in thread \"main\" go.runtime.Panic: ")
printpanicval(p.arg)
```

## Paso 4: Eliminar la Cabecera de Goroutine

**Edita `traceback.go`:**

Encuentra la función `goroutineheader` alrededor de la línea 1215. Añade una sentencia return al principio:

```go
func goroutineheader(gp *g) {
    return  // Add this line to skip printing goroutine info
    level, _, _ := gotraceback()
    // ... rest of original code below (now unreachable)
}
```

## Paso 5: Transformar el Formato de los Frames del Stack

**Continuando en `traceback.go`:**

Encuentra la función `traceback2` alrededor de la línea 945. Comenta la llamada a `gotraceback()` (alrededor de la línea 966):

```go
gp := u.g.ptr()
// level, _, _ := gotraceback()  // Comment this out
var cgoBuf [32]uintptr
```

Luego encuentra donde se imprimen los frames del stack (alrededor de las líneas 991-1008). Reemplaza toda esta sección:

```go
printFuncName(name)
print("(")
if iu.isInlined(uf) {
    print("...")
} else {
    argp := unsafe.Pointer(u.frame.argp)
    printArgs(f, argp, u.symPC())
}
print(")\n")
print("\t", file, ":", line)
if !iu.isInlined(uf) {
    if u.frame.pc > f.entry() {
        print(" +", hex(u.frame.pc-f.entry()))
    }
    if gp.m != nil && gp.m.throwing >= throwTypeRuntime && gp == gp.m.curg || level >= 2 {
        print(" fp=", hex(u.frame.fp), " sp=", hex(u.frame.sp), " pc=", hex(u.frame.pc))
    }
}
print("\n")
```

Con este formato estilo Java:

```go
// Extract just the filename (not full path)
fileName := file
for i := len(file) - 1; i >= 0; i-- {
    if file[i] == '/' || file[i] == '\\' {
        fileName = file[i+1:]
        break
    }
}
print("    at ", name, "(", fileName, ":", line, ")\n")
```

## Paso 6: Recompilar el Runtime de Go

```bash
cd ../  # back to go/src
./make.bash
```

## Paso 7: Probar los Stack Traces Estilo Java

```bash
../go/bin/go run stack_trace_demo.go
```

Deberías ver:

```
Starting the program...
Exception in thread "main" go.runtime.Panic: Something went wrong
    at main.methodC(stack_trace_demo.go:6)
    at main.methodB(stack_trace_demo.go:10)
    at main.methodA(stack_trace_demo.go:14)
    at main.main(stack_trace_demo.go:19)
```

## Entendiendo lo que Hicimos

1. **Cambiamos la Cabecera del Panic** (`panic.go` línea 747): Cambiamos `"panic: "` por `"Exception in thread \"main\" go.runtime.Panic: "`
2. **Eliminamos la Info de Goroutine** (`traceback.go` línea 1215): Añadimos un `return` anticipado en `goroutineheader()`
3. **Simplificamos los Frames del Stack** (`traceback.go` líneas 991-1008): Reemplazamos la salida de Go con el formato Java `"    at name(file:line)"`
4. **Eliminamos Info de Depuración**: Comentamos la llamada a `gotraceback()` y eliminamos offsets hexadecimales y punteros de frame
5. **Solo Nombre de Archivo**: Extraemos el nombre del archivo de la ruta completa usando un bucle

## Lo que Aprendimos

- **Formato del Runtime**: Cómo Go genera los stack traces
- **Manejo de Panics**: Dónde se originan los mensajes de panic
- **Control de Salida**: Modificar las sentencias print del runtime

## Ideas de Extensión

Prueba estas modificaciones adicionales:

1. Añadir color a la salida (rojo para "Exception")
2. Hacerlo configurable mediante una variable de entorno
3. Añadir formato estilo Python como otra opción
4. Incluir conversión de rutas de paquetes (github.com/user/pkg → github.com.user.pkg)

## Limpieza

Para restaurar el formato original de stack traces de Go:

```bash
cd go/src/runtime
git checkout panic.go traceback.go
cd ../
./make.bash
```

## Resumen

Has transformado los stack traces de Go en formato estilo Java:

```
// Before: Technical and verbose
goroutine 1 [running]:
main.methodC()
        /full/path/to/main.go:15 +0x45

// After: Clean and familiar
Exception in thread "main" go.runtime.Panic: ...
    at main.methodC(main.go:15)
```

---

*¡Felicidades por completar todos los ejercicios del taller! Vuelve al [taller principal](../README.md)*
