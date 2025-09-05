# 🧪 Lambda S3 Dispatcher - PRUEBA DE CONCEPTO

> ⚠️ **PROYECTO DE PRUEBA** - Este es un código experimental para demostrar integración S3 + API OCR

## ¿Qué hace?

Un dispatcher simple que:

1. **Lee imágenes** de un bucket S3 
2. **Genera URLs prefirmadas** (presigned URLs) para cada imagen
3. **Envía las URLs** a una API de OCR externa
4. **Recolecta y estructura** las respuestas

## Flujo básico


```
S3 Bucket → Lista archivos → Presigned URLs → API OCR → Resultados JSON
```

## Variables de entorno requeridas

```bash
export BUCKET_NAME="tu-bucket-s3"
export API_URL="http://localhost:8080"  # URL de tu API OCR
export AWS_REGION="us-east-1"           # Región del bucket
```

## Variables opcionales

```bash
export PREFIX="imagenes/"               # Filtrar archivos por prefijo
export MAX_CONCURRENCY="10"            # Límite de requests paralelos (default: 10)
export AWS_PROFILE="tu-perfil"         # Perfil AWS específico
```

## Uso

### Ejecución local de prueba
```bash
go run .
```

### Como Lambda (comentar testHandler, descomentar lambda.Start)
```go
func main() {
    // testHandler()  // ← comentar esta línea
    lambda.Start(Handler)  // ← descomentar esta línea
}
```

## Estructura de respuesta

```json
{
  "bucket": "test-images-bucket",
  "prefix": "imagenes/",
  "processed": 30,
  "api_responses": [
    {
      "key": "imagenes/foto1.png",
      "status_code": 200,
      "result": {
        "key": "imagenes/foto1.png",
        "source_url": "https://...",
        "full_text": "Texto extraído de la imagen"
      }
    }
  ],
  "errors": []
}
```

## Requisitos

- **Go 1.19+**
- **Credenciales AWS** configuradas
- **API OCR** corriendo que acepte `POST /ocr` con formato:
  ```json
  {"key": "archivo.png", "url": "https://presigned-url..."}
  ```

## ⚠️ Nota importante

Este es un **código de prueba** para validar la arquitectura. En producción necesitarías:

- Manejo de errores más robusto
- Logging estructurado
- Métricas y monitoring
- Timeouts configurables
- Rate limiting
- Tests unitarios

---

*Desarrollado como POC para integración S3 + OCR API*