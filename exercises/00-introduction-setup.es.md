# Ejercicio 0: Introducción y Configuración del Entorno

Bienvenido al taller de código fuente de Go. En este ejercicio introductorio, prepararás tu entorno de desarrollo y te familiarizarás con el repositorio del código fuente de Go.

## Objetivos de Aprendizaje

Al finalizar este ejercicio, serás capaz de:

- Tener un entorno de desarrollo de Go funcional
- Saber cómo obtener el código fuente de Go

## Requisitos Previos

Asegúrate de tener lo siguiente instalado:

- Git
- Al menos 4GB de espacio libre en disco

## Paso 1: Instalar o Actualizar Go

**⚠️ Importante**: Necesitas una instalación existente de Go (versión 1.24 o superior) para compilar Go desde el código fuente. Esto se conoce como "bootstrapping": usar un compilador de Go existente para compilar el nuevo.

### Verifica tu Versión Actual de Go

```bash
go version
# Debería mostrar: go version go1.24 o superior
```

### Si Go no está Instalado o es una Versión Antigua

Si no tienes Go instalado, o tu versión es anterior a la 1.24:

1. **Descarga Go**: Visita <https://go.dev/dl/> y descarga el instalador apropiado para tu sistema operativo
2. **Instala Go**: Sigue la guía oficial de instalación para tu plataforma
3. **Verifica la instalación**: Abre una nueva terminal y ejecuta:

   ```bash
   go version
   # Debería mostrar: go version go1.24 o superior
   ```

**Ayuda con la instalación**: Si necesitas instrucciones detalladas de instalación, consulta la [guía oficial de instalación de Go](https://go.dev/doc/install).

## Paso 2: Clonar el Código Fuente de Go

Vamos a clonar el repositorio oficial de Go. Usamos `--depth 1` para evitar descargar todo el historial, lo que hace que la clonación sea mucho más rápida:

Para mantener la consistencia a lo largo del taller, usaremos la versión 1.26.1 de Go.

```bash
git clone --depth 1 --branch go1.26.1 https://go.googlesource.com/go
cd go
```

## Paso 3: Verificar la Versión Go 1.26.1

Verifica que estás en la versión correcta:

```bash
git describe --tags
# Debería mostrar: go1.26.1
```

## Lo que Hemos Logrado

- Instalamos o verificamos Go 1.24+ para el bootstrapping
- Clonamos el repositorio oficial de Go en la versión 1.26.1
- El entorno está listo para compilar Go desde el código fuente

## Siguientes Pasos

Perfecto. Tu entorno está configurado y listo. En el [Ejercicio 1: Compilar Go sin Cambios](./01-compile-go-unchanged.es.md), compilarás la toolchain de Go desde el código fuente y explorarás la estructura del compilador de Go mientras se compila.
