package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ContinuumApp/continuum-plugin-tmdb/metadata"
)

func TestGetImagesReturnsRawPaths(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"images": map[string]any{
					"secure_base_url": serverURL(t, r) + "/images/",
				},
			})
		case "/movie/42":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 42,
				"images": map[string]any{
					"posters": []map[string]any{
						{"file_path": "/poster.jpg", "width": 2000, "height": 3000, "vote_average": 8.0},
					},
					"backdrops": []map[string]any{
						{"file_path": "/backdrop.jpg", "width": 3840, "height": 2160, "vote_average": 7.0},
					},
					"logos": []map[string]any{
						{"file_path": "/logo.png", "width": 1200, "height": 600, "vote_average": 6.0},
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	p := newTMDBTestProvider(server.URL)

	images, err := p.GetImages(context.Background(), metadata.ImageRequest{
		ProviderIDs: map[string]string{"tmdb": "42"},
		ContentType: "movie",
	})
	if err != nil {
		t.Fatalf("GetImages() error = %v", err)
	}
	if len(images) != 3 {
		t.Fatalf("len(images) = %d, want 3", len(images))
	}

	got := map[metadata.ImageType]string{}
	for _, img := range images {
		got[img.Type] = img.URL
	}

	if got[metadata.ImagePoster] != "/poster.jpg" {
		t.Fatalf("poster URL = %q", got[metadata.ImagePoster])
	}
	if got[metadata.ImageBackdrop] != "/backdrop.jpg" {
		t.Fatalf("backdrop URL = %q", got[metadata.ImageBackdrop])
	}
	if got[metadata.ImageLogo] != "/logo.png" {
		t.Fatalf("logo URL = %q", got[metadata.ImageLogo])
	}
}

func TestGetSeasonsReturnsRawPosterPath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"images": map[string]any{
					"secure_base_url": serverURL(t, r) + "/images/",
				},
			})
		case "/tv/77":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 77,
				"seasons": []map[string]any{
					{"season_number": 2, "poster_path": "/season-two.jpg"},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	p := newTMDBTestProvider(server.URL)

	seasons, err := p.GetSeasons(context.Background(), metadata.SeasonsRequest{
		ProviderIDs: map[string]string{"tmdb": "77"},
		ContentType: "series",
	})
	if err != nil {
		t.Fatalf("GetSeasons() error = %v", err)
	}
	if len(seasons) != 1 {
		t.Fatalf("len(seasons) = %d, want 1", len(seasons))
	}
	if seasons[0].PosterPath != "/season-two.jpg" {
		t.Fatalf("season poster = %q", seasons[0].PosterPath)
	}
}

func TestGetEpisodesReturnsRawStillPath(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"images": map[string]any{
					"secure_base_url": serverURL(t, r) + "/images/",
				},
			})
		case "/tv/77/season/2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": 2,
				"episodes": []map[string]any{
					{
						"id":             9001,
						"season_number":  2,
						"episode_number": 5,
						"name":           "Test Episode",
						"still_path":     "/still.jpg",
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	p := newTMDBTestProvider(server.URL)

	episodes, err := p.GetEpisodes(context.Background(), metadata.EpisodesRequest{
		ProviderIDs:  map[string]string{"tmdb": "77"},
		SeasonNumber: 2,
	})
	if err != nil {
		t.Fatalf("GetEpisodes() error = %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("len(episodes) = %d, want 1", len(episodes))
	}
	if episodes[0].StillPath != "/still.jpg" {
		t.Fatalf("episode still = %q", episodes[0].StillPath)
	}
}

func TestGetPersonDetail_UsesTMDBID(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/person/287":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":             287,
				"name":           "Brad Pitt",
				"biography":      "Biography",
				"birthday":       "1963-12-18",
				"place_of_birth": "Shawnee, Oklahoma, USA",
				"homepage":       "https://example.test/brad",
				"profile_path":   "/brad.jpg",
				"external_ids": map[string]any{
					"imdb_id": "nm0000093",
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	p := newTMDBTestProvider(server.URL)

	person, err := p.GetPersonDetail(context.Background(), metadata.PersonDetailRequest{
		ProviderIDs: map[string]string{"tmdb": "287"},
	})
	if err != nil {
		t.Fatalf("GetPersonDetail() error = %v", err)
	}
	if person == nil {
		t.Fatal("GetPersonDetail() returned nil person")
	}
	if person.Name != "Brad Pitt" {
		t.Fatalf("Name = %q, want Brad Pitt", person.Name)
	}
	if person.BirthDate != "1963-12-18" {
		t.Fatalf("BirthDate = %q, want 1963-12-18", person.BirthDate)
	}
	if person.Birthplace != "Shawnee, Oklahoma, USA" {
		t.Fatalf("Birthplace = %q", person.Birthplace)
	}
	if person.Homepage != "https://example.test/brad" {
		t.Fatalf("Homepage = %q", person.Homepage)
	}
	if person.PhotoPath != "/brad.jpg" {
		t.Fatalf("PhotoPath = %q, want /brad.jpg", person.PhotoPath)
	}
	if person.ProviderIDs["tmdb"] != "287" {
		t.Fatalf("ProviderIDs[tmdb] = %q, want 287", person.ProviderIDs["tmdb"])
	}
	if person.ProviderIDs["imdb"] != "nm0000093" {
		t.Fatalf("ProviderIDs[imdb] = %q, want nm0000093", person.ProviderIDs["imdb"])
	}
}

func TestGetPersonDetail_FindsTMDBPersonByIMDbID(t *testing.T) {
	t.Parallel()

	findSource := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/find/nm0000093":
			findSource = r.URL.Query().Get("external_source")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"person_results": []map[string]any{
					{"id": 287},
				},
			})
		case "/person/287":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":           287,
				"name":         "Brad Pitt",
				"profile_path": "/brad.jpg",
				"external_ids": map[string]any{
					"imdb_id": "nm0000093",
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.String())
		}
	}))
	defer server.Close()

	p := newTMDBTestProvider(server.URL)

	person, err := p.GetPersonDetail(context.Background(), metadata.PersonDetailRequest{
		ProviderIDs: map[string]string{"imdb": "nm0000093"},
	})
	if err != nil {
		t.Fatalf("GetPersonDetail() error = %v", err)
	}
	if person == nil {
		t.Fatal("GetPersonDetail() returned nil person")
	}
	if person.ProviderIDs["tmdb"] != "287" {
		t.Fatalf("ProviderIDs[tmdb] = %q, want 287", person.ProviderIDs["tmdb"])
	}
	if findSource != "imdb_id" {
		t.Fatalf("external_source = %q, want imdb_id", findSource)
	}
}

func newTMDBTestProvider(baseURL string) *Provider {
	client := NewClient(1000)
	client.SetBaseURL(baseURL)
	return NewProviderWithClient(client)
}

func serverURL(t *testing.T, r *http.Request) string {
	t.Helper()
	return "http://" + r.Host
}
