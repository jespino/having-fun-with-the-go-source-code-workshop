# Ejercicio 0: Introduccion y Configuracion del Entorno

Bienvenido al taller de codigo fuente de Go. En este ejercicio introductorio, prepararas tu entorno de desarrollo y te familiarizaras con el repositorio del codigo fuente de Go.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, seras capaz de:

- Tener un entorno de desarrollo de Go funcional
- Saber como obtener el codigo fuente de Go

## Requisitos Previos

Asegurate de tener lo siguiente instalado:

- Git
- Al menos 4GB de espacio libre en disco

## Paso 1: Instalar o Actualizar Go

**⚠️ Importante**: Necesitas una instalacion existente de Go (version 1.24 o superior) para compilar Go desde el codigo fuente. Esto se conoce como "bootstrapping": usar un compilador de Go existente para compilar el nuevo.

### Verifica tu Version Actual de Go

```bash
go version
# Deberia mostrar: go version go1.24 o superior
```

### Si Go no esta Instalado o es una Version Antigua

Si no tienes Go instalado, o tu version es anterior a la 1.24:

1. **Descarga Go**: Visita <https://go.dev/dl/> y descarga el instalador apropiado para tu sistema operativo
2. **Instala Go**: Sigue la guia oficial de instalacion para tu plataforma
3. **Verifica la instalacion**: Abre una nueva terminal y ejecuta:

   ```bash
   go version
   # Deberia mostrar: go version go1.24 o superior
   ```

**Ayuda con la instalacion**: Si necesitas instrucciones detalladas de instalacion, consulta la [guia oficial de instalacion de Go](https://go.dev/doc/install).

## Paso 2: Clonar el Codigo Fuente de Go

Vamos a clonar el repositorio oficial de Go. Usamos `--depth 1` para evitar descargar todo el historial, lo que hace que la clonacion sea mucho mas rapida:

Para mantener la consistencia a lo largo del taller, usaremos la version 1.26.1 de Go.

```bash
git clone --depth 1 --branch go1.26.1 https://go.googlesource.com/go
cd go
```

## Paso 3: Verificar la Version Go 1.26.1

Verifica que estas en la version correcta:

```bash
git describe --tags
# Deberia mostrar: go1.26.1
```

## Lo que Hemos Logrado

- Instalamos o verificamos Go 1.24+ para el bootstrapping
- Clonamos el repositorio oficial de Go en la version 1.26.1
- El entorno esta listo para compilar Go desde el codigo fuente

## Siguientes Pasos

Perfecto. Tu entorno esta configurado y listo. En el [Ejercicio 1: Compilar Go sin Cambios](./01-compile-go-unchanged.es.md), compilaras la toolchain de Go desde el codigo fuente y exploraras la estructura del compilador de Go mientras se compila.
