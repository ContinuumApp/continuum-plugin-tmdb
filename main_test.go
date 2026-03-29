package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	pluginv1 "github.com/ContinuumApp/continuum-plugin-sdk/pkg/pluginproto/continuum/plugin/v1"
	"github.com/ContinuumApp/continuum-plugin-tmdb/provider"
)

func TestRuntimeServerConfigure_ConfiguresTMDBProvider(t *testing.T) {
	server := &runtimeServer{}

	_, err := server.Configure(context.Background(), &pluginv1.ConfigureRequest{
		Config: []*pluginv1.ConfigEntry{
			{
				Key: "connection",
				Value: mustStruct(t, map[string]any{
					"api_key": "tmdb-api-key",
				}),
			},
		},
	})
	if err != nil {
		t.Fatalf("Configure() returned error: %v", err)
	}

	provider, err := server.providerForRequest()
	if err != nil {
		t.Fatalf("providerForRequest() returned error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider to be configured")
	}
	if server.config.APIKey != "tmdb-api-key" {
		t.Fatalf("config.APIKey = %q, want tmdb-api-key", server.config.APIKey)
	}
}

func TestRuntimeServerConfigure_RequiresTMDBAPIKey(t *testing.T) {
	server := &runtimeServer{}

	_, err := server.Configure(context.Background(), &pluginv1.ConfigureRequest{})
	if err != nil {
		t.Fatalf("Configure() returned error: %v", err)
	}

	_, err = server.providerForRequest()
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("providerForRequest() error code = %v, want %v", status.Code(err), codes.FailedPrecondition)
	}
}

func mustStruct(t *testing.T, value map[string]any) *structpb.Struct {
	t.Helper()

	result, err := structpb.NewStruct(value)
	if err != nil {
		t.Fatalf("structpb.NewStruct() returned error: %v", err)
	}
	return result
}

func TestResolveImageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		variant string
		wantURL string
	}{
		// poster variants
		{name: "poster card", path: "tmdb://poster/poster.jpg", variant: "card", wantURL: "https://image.tmdb.org/t/p/w300/poster.jpg"},
		{name: "poster featured", path: "tmdb://poster/poster.jpg", variant: "featured", wantURL: "https://image.tmdb.org/t/p/w500/poster.jpg"},
		{name: "poster full", path: "tmdb://poster/poster.jpg", variant: "full", wantURL: "https://image.tmdb.org/t/p/w780/poster.jpg"},
		{name: "poster original", path: "tmdb://poster/poster.jpg", variant: "original", wantURL: "https://image.tmdb.org/t/p/original/poster.jpg"},
		{name: "poster empty variant", path: "tmdb://poster/poster.jpg", variant: "", wantURL: "https://image.tmdb.org/t/p/original/poster.jpg"},
		// backdrop variants
		{name: "backdrop featured", path: "tmdb://backdrop/backdrop.jpg", variant: "featured", wantURL: "https://image.tmdb.org/t/p/w1280/backdrop.jpg"},
		{name: "backdrop card", path: "tmdb://backdrop/backdrop.jpg", variant: "card", wantURL: "https://image.tmdb.org/t/p/w300/backdrop.jpg"},
		// still variants
		{name: "still card", path: "tmdb://still/still.jpg", variant: "card", wantURL: "https://image.tmdb.org/t/p/w300/still.jpg"},
		// logo variants
		{name: "logo featured", path: "tmdb://logo/logo.png", variant: "featured", wantURL: "https://image.tmdb.org/t/p/w500/logo.png"},
		// profile variants
		{name: "profile card", path: "tmdb://profile/person.jpg", variant: "card", wantURL: "https://image.tmdb.org/t/p/w185/person.jpg"},
		// empty path
		{name: "empty path", path: "", variant: "card", wantURL: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Each sub-test gets its own mock server to avoid shared-state issues
			// with the client's sync.Once configuration cache.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/configuration" {
					_ = json.NewEncoder(w).Encode(map[string]any{
						"images": map[string]any{
							"secure_base_url": "https://image.tmdb.org/t/p/",
						},
					})
					return
				}
				t.Errorf("unexpected path: %s", r.URL.Path)
				http.NotFound(w, r)
			}))
			t.Cleanup(server.Close)

			client := provider.NewClient("test-key", 1000)
			client.SetBaseURL(server.URL)
			p := provider.NewProviderWithClient(client)

			rs := &runtimeServer{provider: p}
			ms := &metadataServer{runtime: rs}

			resp, err := ms.ResolveImageURL(context.Background(), &pluginv1.ResolveImageURLRequest{
				Path:    tc.path,
				Variant: tc.variant,
			})
			if err != nil {
				t.Fatalf("ResolveImageURL() error = %v", err)
			}
			if resp.GetUrl() != tc.wantURL {
				t.Fatalf("URL = %q, want %q", resp.GetUrl(), tc.wantURL)
			}
		})
	}
}
