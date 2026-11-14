package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/faber-numeris/sem/internal/core/domain"
	"github.com/faber-numeris/sem/internal/ports/inbound"
	"github.com/faber-numeris/sem/internal/ports/outbound"
)

type Service struct {
	repo        outbound.Repository
	instanceURL string
}

var _ inbound.EmulatorService = (*Service)(nil)

func NewService(repo outbound.Repository, instanceURL string) *Service {
	if instanceURL == "" {
		instanceURL = "http://localhost:8080"
	}
	return &Service{
		repo:        repo,
		instanceURL: strings.TrimRight(instanceURL, "/"),
	}
}

func randomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func normalizeVersion(version string) string {
	v := strings.TrimSpace(version)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	return v
}

var supportedVersions = []domain.APIVersion{
	{Label: "Winter '24", URL: "/services/data/v59.0", Version: "59.0"},
	{Label: "Spring '24", URL: "/services/data/v60.0", Version: "60.0"},
	{Label: "Summer '24", URL: "/services/data/v61.0", Version: "61.0"},
	{Label: "Winter '25", URL: "/services/data/v62.0", Version: "62.0"},
}

func (s *Service) ListVersions() []domain.APIVersion {
	out := make([]domain.APIVersion, len(supportedVersions))
	copy(out, supportedVersions)
	return out
}

func (s *Service) ValidateVersion(version string) error {
	norm := normalizeVersion(version)
	for _, v := range supportedVersions {
		if v.Version == norm {
			return nil
		}
	}
	return domain.ErrInvalidVersion
}

func (s *Service) ListResources(version string) (map[string]string, error) {
	if err := s.ValidateVersion(version); err != nil {
		return nil, err
	}
	v := normalizeVersion(version)
	base := fmt.Sprintf("/services/data/v%s", v)
	return map[string]string{
		"sobjects":  base + "/sobjects",
		"search":    base + "/search",
		"query":     base + "/query",
		"queryAll":  base + "/queryAll",
		"limits":    base + "/limits",
		"tooling":   base + "/tooling",
		"jobs":      base + "/jobs",
		"composite": base + "/composite",
	}, nil
}
