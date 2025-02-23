package search

import (
	"context"
	"crypto/tls"
	"fluencybe/internal/core/config"
	"fluencybe/pkg/logger"
	"fmt"
	"net/http"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

func NewOpenSearchClient(cfg config.OpenSearchConfig, log *logger.PrettyLogger) (*opensearch.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}

	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{"https://" + cfg.Host + ":" + cfg.Port},
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: transport,
	})
	if err != nil {
		log.Critical("OPENSEARCH_CONNECTION_ERROR", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to create OpenSearch client")
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	// Ping OpenSearch
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := (opensearchapi.InfoRequest{}).Do(ctx, client); err != nil {
		log.Critical("OPENSEARCH_PING_ERROR", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to ping OpenSearch")
		return nil, fmt.Errorf("failed to ping OpenSearch: %w", err)
	}

	log.Info("OPENSEARCH_CONNECTION_SUCCESS", nil, "OpenSearch connection established successfully")
	return client, nil
}
