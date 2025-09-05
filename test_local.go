// test_local.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	out1, err := Handler(context.Background())
	fmt.Println("PRESIGNED =>", asJSON(out1), "ERR:", err)
}

func asJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
