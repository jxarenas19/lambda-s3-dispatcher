// test_local.go
package main

import (
	"context"
	"encoding/json"
	"os"
)

func testHandler() {
	// Setea envs para la prueba local
	os.Setenv("BUCKET_NAME", "tu-bucket")
	os.Setenv("PREFIX", "imagenes/")
	os.Setenv("API_URL", "http://localhost:8080") // tu API mock/real
	os.Setenv("AWS_REGION", "us-east-1")          // región
	// Si usas perfiles locales: export AWS_PROFILE=default (desde la terminal)

	// Caso 1: presigned
	Handler(context.Background())

}

func asJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
