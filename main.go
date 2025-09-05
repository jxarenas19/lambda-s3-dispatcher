package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	// "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type APIRequest struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

type OCRResult struct {
	Key       string `json:"key"`
	SourceURL string `json:"source_url,omitempty"`
	FullText  string `json:"full_text"`
}

type APIResponse struct {
	Key        string     `json:"key"`
	StatusCode int        `json:"status_code"`
	Result     *OCRResult `json:"result,omitempty"`
	Err        string     `json:"err,omitempty"`
}

type Output struct {
	Bucket       string        `json:"bucket"`
	Prefix       string        `json:"prefix"`
	Processed    int           `json:"processed"`
	APIResponses []APIResponse `json:"api_responses"`
	Errors       []string      `json:"errors"`
}

var (
	s3cli      *s3.Client
	presigner  *s3.PresignClient
	httpClient *http.Client

	bucket string
	prefix string
	apiURL string
	limit  int
)

func init() {
	// ENV obligatorios: BUCKET_NAME, API_URL
	bucket = os.Getenv("BUCKET_NAME")
	prefix = os.Getenv("PREFIX") // puede ser vacío
	apiURL = os.Getenv("API_URL")

	// Concurrencia (opcional, default 10)
	limit = 10
	if v := os.Getenv("MAX_CONCURRENCY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	httpClient = &http.Client{Timeout: 10 * time.Second}

	// AWS SDK v2
	awsProfile := "default"
	fmt.Printf("Usando AWS profile: %s\n", awsProfile)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithSharedConfigProfile(awsProfile),
	)
	if err != nil {
		panic(err)
	}
	s3cli = s3.NewFromConfig(cfg)
	presigner = s3.NewPresignClient(s3cli)
}

func Handler(ctx context.Context) (Output, error) {
	innitialTime := time.Now()
	formattedInnitialTime := innitialTime.Format("2006-01-02 15:04:05")
	fmt.Println("Formatted Time:", formattedInnitialTime)
	if bucket == "" || apiURL == "" {
		return Output{}, errors.New("faltan envs: BUCKET_NAME o API_URL")
	}

	fmt.Printf("Conectando a bucket: %s, región: %s\n", bucket, os.Getenv("AWS_REGION"))

	keys, err := listKeys(ctx, bucket, prefix)
	if err != nil {
		return Output{}, fmt.Errorf("listKeys: %w", err)
	}
	keys = keys
	out := Output{
		Bucket:    bucket,
		Prefix:    prefix,
		Processed: len(keys),
	}
	if len(keys) == 0 {
		return out, nil
	}

	results := make(chan APIResponse, len(keys))
	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup

	for _, key := range keys {
		// corta si el contexto fue cancelado
		select {
		case <-ctx.Done():
			out.Errors = append(out.Errors, "context cancelled/timeout")
			break
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // adquiere cupo
		go func(k string) {
			defer wg.Done()
			resp := processOnePresigned(ctx, k)
			// Libera el cupo ANTES de enviar al canal para evitar bloqueo circular
			<-sem
			results <- resp
		}(key)
	}

	// cierra el canal al terminar todas
	go func() {
		wg.Wait()
		close(results)
	}()

	// recolecta
	for r := range results {
		if r.Err != "" || r.StatusCode >= 400 {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %s", r.Key, r.Err))
		}
		out.APIResponses = append(out.APIResponses, r)
	}
	dur := time.Since(innitialTime)
	fmt.Printf("La función tardó %v\n", dur)
	return out, nil
}

func listKeys(ctx context.Context, bucket, prefix string) ([]string, error) {
	var keys []string
	p := s3.NewListObjectsV2Paginator(s3cli, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, o := range page.Contents {
			key := aws.ToString(o.Key)
			if key == "" || strings.HasSuffix(key, "/") {
				continue
			}
			if aws.ToInt64(o.Size) <= 0 {
				continue
			}
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func processOnePresigned(ctx context.Context, key string) APIResponse {
	// Presign GET (válida 10 min)
	ps, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(10*time.Minute))
	if err != nil {
		return APIResponse{Key: key, Err: "presign: " + err.Error()}
	}

	// Llama a la API mock (chi) con JSON {key, url}
	body, _ := json.Marshal(APIRequest{Key: key, URL: ps.URL})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL+"/ocr", strings.NewReader(string(body)))
	if err != nil {
		return APIResponse{Key: key, Err: "build req: " + err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return APIResponse{Key: key, Err: "http: " + err.Error()}
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	apiResp := APIResponse{Key: key, StatusCode: resp.StatusCode}

	if resp.StatusCode == 200 {
		var ocrResult OCRResult
		if err := json.Unmarshal(b, &ocrResult); err == nil {
			apiResp.Result = &ocrResult
		} else {
			apiResp.Err = "parse json: " + err.Error()
		}
	} else {
		apiResp.Err = string(b)
	}

	return apiResp
}

func main() {
	// Uncomment para testing local:
	testHandler()
	// lambda.Start(Handler)
}
