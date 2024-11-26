# OS Go Concurrency
Este proyecto tiene como objetivo demostrar cómo la concurrencia puede optimizar la solución de problemas a través del desarrollo de una librería que implementa patrones concurrentes. Ubicada en la ruta `pkg/pipeline`, esta librería ofrece algoritmos de sincronización como lo son: worker pool, pipeline, fan-in y fan-out. 

Estas herramientas permiten distribuir tareas eficientemente entre múltiples procesos, maximizando el uso de los recursos del sistema y mejorando significativamente el rendimiento. La librería es versátil y puede aplicarse a una variedad de casos de uso, proporcionando un enfoque flexible para abordar problemas que requieren procesamiento paralelo.

En este proyecto, la librería se utilizó para dar realidad a una simulación de problema real: el procesamiento eficiente de productos, buscando principalmente medir como un enfoque concurrente puede optimizar tiempo en comparación con un enfoque secuencial.

## Requerimientos

- Docker
- Docker Compose

## Descripción del Proyecto

La aplicación está desarrollada en Go 1.23 y utiliza un entorno de contenedores para gestionar servicios de telemetría y monitoreo, como Jaeger, Prometheus y Grafana. La concurrencia es una característica clave que permite manejar grandes volúmenes de productos de manera eficiente.

## Configuración Inicial

Antes de iniciar, asegúrate de que todos los servicios necesarios están bien configurados. Revisa el archivo `config/app.json` que contiene parámetros esenciales de configuración:

```json
{
  "workers": 4  // Define la cantidad de workers en el pool concurrente
}
```

Con estos vamos a poder establecer la cantidad "hilos" que vamos a tener a disposición para procesar nuestro workload por partes.

## Servicios Incluidos

### OpenTelemetry Collector

- **Descripción**: Agrega soporte para recopilar, procesar y exportar datos de telemetría.
- **Conexiones Internas**: Recibe datos de telemetría desde la aplicación y puede enviar datos a otros servicios de observabilidad.

### Jaeger

- **Descripción**: Sistema de seguimiento distribuido para monitorear transacciones entre servicios dentro de tu aplicación.
- **Interfaz Web**: [http://localhost:16686](http://localhost:16686)

### Prometheus

- **Descripción**: Sistema de monitoreo y de alerta. Consulta tiempo-series de datos.
- **Interfaz Web**: [http://localhost:9090](http://localhost:9090)

### Grafana

- **Descripción**: Plataforma de análisis y monitoreo de métricas.
- **Interfaz Web**: [http://localhost:3000](http://localhost:3000)
- **Nota**: Configura dashboards para visualizar las métricas recolectadas por Prometheus.

## Cómo Ejecutar

1. Clona el repositorio a tu máquina local.
2. Asegúrate de que Docker y Docker Compose estén instalados y funcionando.
3. Ejecuta el siguiente comando para construir y levantar todos los servicios:

   ```bash
   docker-compose up --build
   ```

4. La aplicación y todos los servicios estarán activos y en ejecución.

## Endpoints de la Aplicación

A continuación, se presentan ejemplos de cómo interactuar con la API usando cURL.

- **Verificar servicio**:
  ```bash
  curl -X GET http://localhost:8080/ping
  ```

- **Procesar productos secuencialmente**:
  ```bash
  curl -X POST http://localhost:8080/products -H "Content-Type: application/json" -d '{
    "products": [
      {
        "product_id": "001",
        "name": "Widget 1"
      },
      {
        "product_id": "002",
        "name": "Widget 2"
      },
      {
        "product_id": "003",
        "name": "Widget 3"
      },
      {
        "product_id": "004",
        "name": "Widget 4"
      },
      {
        "product_id": "005",
        "name": "Widget 5"
      }
    ]
  }'
  ```

- **Procesar productos concurrentemente**:
  ```bash
  curl -X POST http://localhost:8080/products?mode=concurrent -H "Content-Type: application/json" -d '{
    "products": [
      {
        "product_id": "001",
        "name": "Widget 1"
      },
      {
        "product_id": "002",
        "name": "Widget 2"
      },
      {
        "product_id": "003",
        "name": "Widget 3"
      },
      {
        "product_id": "004",
        "name": "Widget 4"
      },
      {
        "product_id": "005",
        "name": "Widget 5"
      }
    ]
  }'
  ```
Vemos que este último endpoint tiene el query param `?mode=concurrent` 

## Notas Adicionales

- Ajusta el número de `workers` en `config/app.json` para optimizar el rendimiento de acuerdo al entorno de ejecución.
- Asegúrate de monitorear los logs de cada servicio usando la interfaz adecuada (Grafana, Prometheus, Jaeger) para asegurar que el sistema se comporte según lo esperado.