package service

import (
	"fmt"
	"strconv"
	"strings"

	"graphql-pkm/internal/ai"
	"graphql-pkm/internal/database"
	"graphql-pkm/internal/models"
)

const (
	settingAIEnabled      = "ai.enabled"
	settingAIProvider     = "ai.provider"
	settingAIBaseURL      = "ai.base_url"
	settingAIModel        = "ai.model"
	settingAIAPIKey       = "ai.api_key"
	settingAIEmbedBaseURL = "ai.embed_base_url"
	settingAIEmbedModel   = "ai.embed_model"
)

// AISettingsDefaults are the environment/default values used when no saved
// setting exists. Saved settings always take precedence.
type AISettingsDefaults struct {
	Enabled      bool
	Provider     string
	BaseURL      string
	APIKey       string
	Model        string
	EmbedBaseURL string
	EmbedAPIKey  string
	EmbedModel   string
}

type resolvedAISettings struct {
	Enabled      bool
	Provider     string
	BaseURL      string
	Model        string
	APIKey       string
	EmbedBaseURL string
	EmbedAPIKey  string
	EmbedModel   string
}

type AISettingsService struct {
	store    database.AppSettingsStore
	client   *ai.Client
	defaults AISettingsDefaults
}

func NewAISettingsService(store database.AppSettingsStore, client *ai.Client, defaults AISettingsDefaults) *AISettingsService {
	return &AISettingsService{
		store:    store,
		client:   client,
		defaults: defaults,
	}
}

// Reload resolves the effective settings and applies them to the shared AI
// client. It is called at startup and after every save.
func (s *AISettingsService) Reload() error {
	settings, err := s.resolve()
	if err != nil {
		return err
	}

	s.client.UpdateConfig(ai.ClientConfig{
		Enabled:      settings.Enabled,
		BaseURL:      settings.BaseURL,
		APIKey:       settings.APIKey,
		Model:        settings.Model,
		EmbedBaseURL: settings.EmbedBaseURL,
		EmbedAPIKey:  settings.EmbedAPIKey,
		EmbedModel:   settings.EmbedModel,
	})
	return nil
}

func (s *AISettingsService) GetAISettings() (*models.AISettings, error) {
	settings, err := s.resolve()
	if err != nil {
		return nil, err
	}
	return toAISettingsModel(settings), nil
}

func (s *AISettingsService) SaveAISettings(input models.SaveAISettingsInput) (*models.AISettings, error) {
	provider := strings.TrimSpace(input.Provider)
	if provider == "" {
		provider = "ollama"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")
	model := strings.TrimSpace(input.Model)
	embedBaseURL := strings.TrimRight(strings.TrimSpace(input.EmbedBaseURL), "/")
	embedModel := strings.TrimSpace(input.EmbedModel)

	if input.Enabled && baseURL == "" {
		return nil, fmt.Errorf("base URL is required when AI is enabled")
	}
	if input.Enabled && model == "" {
		return nil, fmt.Errorf("model is required when AI is enabled")
	}

	values := map[string]string{
		settingAIEnabled:      strconv.FormatBool(input.Enabled),
		settingAIProvider:     provider,
		settingAIBaseURL:      baseURL,
		settingAIModel:        model,
		settingAIEmbedBaseURL: embedBaseURL,
		settingAIEmbedModel:   embedModel,
	}

	// Omitted API key = keep, empty string = clear, value = replace.
	if input.APIKey != nil {
		values[settingAIAPIKey] = strings.TrimSpace(*input.APIKey)
	}

	if err := s.store.SetAppSettings(values); err != nil {
		return nil, err
	}
	if err := s.Reload(); err != nil {
		return nil, err
	}
	return s.GetAISettings()
}

func (s *AISettingsService) ListAIModels(input models.ListAIModelsInput) (*models.AIModelList, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")
	if baseURL == "" {
		message := "Base URL is required."
		return &models.AIModelList{Models: []string{}, Error: &message}, nil
	}

	var apiKey string
	if input.APIKey != nil {
		apiKey = strings.TrimSpace(*input.APIKey)
	} else {
		settings, err := s.resolve()
		if err != nil {
			return nil, err
		}
		apiKey = settings.APIKey
	}

	modelIDs, err := s.client.ListModels(baseURL, apiKey)
	if err != nil {
		message := err.Error()
		return &models.AIModelList{Models: []string{}, Error: &message}, nil
	}
	if modelIDs == nil {
		modelIDs = []string{}
	}
	return &models.AIModelList{Models: modelIDs}, nil
}

func (s *AISettingsService) resolve() (resolvedAISettings, error) {
	settings := resolvedAISettings{
		Enabled:      s.defaults.Enabled,
		Provider:     s.defaults.Provider,
		BaseURL:      s.defaults.BaseURL,
		Model:        s.defaults.Model,
		APIKey:       s.defaults.APIKey,
		EmbedBaseURL: s.defaults.EmbedBaseURL,
		EmbedAPIKey:  s.defaults.EmbedAPIKey,
		EmbedModel:   s.defaults.EmbedModel,
	}
	if settings.Provider == "" {
		settings.Provider = "ollama"
	}

	if value, found, err := s.store.GetAppSetting(settingAIEnabled); err != nil {
		return settings, err
	} else if found {
		if parsed, parseErr := strconv.ParseBool(strings.TrimSpace(value)); parseErr == nil {
			settings.Enabled = parsed
		}
	}
	if value, found, err := s.store.GetAppSetting(settingAIProvider); err != nil {
		return settings, err
	} else if found && strings.TrimSpace(value) != "" {
		settings.Provider = strings.TrimSpace(value)
	}
	if value, found, err := s.store.GetAppSetting(settingAIBaseURL); err != nil {
		return settings, err
	} else if found {
		settings.BaseURL = strings.TrimRight(strings.TrimSpace(value), "/")
	}
	if value, found, err := s.store.GetAppSetting(settingAIModel); err != nil {
		return settings, err
	} else if found {
		settings.Model = strings.TrimSpace(value)
	}
	if value, found, err := s.store.GetAppSetting(settingAIAPIKey); err != nil {
		return settings, err
	} else if found {
		settings.APIKey = strings.TrimSpace(value)
	}
	if value, found, err := s.store.GetAppSetting(settingAIEmbedBaseURL); err != nil {
		return settings, err
	} else if found {
		settings.EmbedBaseURL = strings.TrimRight(strings.TrimSpace(value), "/")
	}
	if value, found, err := s.store.GetAppSetting(settingAIEmbedModel); err != nil {
		return settings, err
	} else if found {
		settings.EmbedModel = strings.TrimSpace(value)
	}

	if settings.EmbedBaseURL == "" {
		settings.EmbedBaseURL = settings.BaseURL
	}
	if settings.EmbedAPIKey == "" {
		settings.EmbedAPIKey = settings.APIKey
	}
	if settings.Model == "" {
		settings.Model = "llama3.2"
	}
	if settings.EmbedModel == "" {
		settings.EmbedModel = "nomic-embed-text"
	}
	settings.Enabled = settings.Enabled && settings.BaseURL != ""

	return settings, nil
}

func toAISettingsModel(settings resolvedAISettings) *models.AISettings {
	return &models.AISettings{
		Enabled:      settings.Enabled,
		Provider:     settings.Provider,
		BaseURL:      settings.BaseURL,
		Model:        settings.Model,
		HasAPIKey:    settings.APIKey != "",
		EmbedBaseURL: settings.EmbedBaseURL,
		EmbedModel:   settings.EmbedModel,
	}
}
