package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

type AccountService struct {
	client  *resty.Client
	baseURL string
}

func NewAccountService() *AccountService {
	baseURL := os.Getenv("ACCOUNT_SERVICE_URL")
	if baseURL == "" {
		log.Fatal("ACCOUNT_SERVICE_URL is not set")
	}
	client := resty.New()
	client.SetHostURL(baseURL)

	return &AccountService{
		client:  client,
		baseURL: baseURL,
	}
}

func (a *AccountService) ValidateToken(token string) (bool, error) {
	var result struct {
		IsValid bool `json:"isValid"`
	}

	resp, err := a.client.R().
		SetQueryParam("accessToken", token).
		SetResult(&result).
		Get("/api/Authentication/Validate")

	if err != nil {
		return false, err
	}

	if resp.StatusCode() != 200 {
		return false, fmt.Errorf("failed to validate token: %s", resp.Status())
	}

	return result.IsValid, nil
}

func (a *AccountService) GetRolesByID(token string) ([]string, error) {
	var result struct {
		Roles []struct {
			Name string `json:"name"`
		} `json:"roles"`
	}

	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&result).
		Get("/api/Accounts/Me")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get user roles, status: %s", resp.Status())
	}

	var roles []string
	for _, role := range result.Roles {
		roles = append(roles, role.Name)
	}

	return roles, nil
}

func (a *AccountService) GetRolesByAccountID(accountID uint, token string) ([]string, error) {
	var result struct {
		Roles []struct {
			Name string `json:"name"`
		} `json:"roles"`
	}

	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&result).
		Get(fmt.Sprintf("/api/Accounts/%d/roles", accountID))

	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе ролей аккаунта %d: %v", accountID, err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("не удалось получить роли для аккаунта ID %d: статус: %s, ответ: %s", accountID, resp.Status(), resp.String())
	}

	var roles []string
	for _, role := range result.Roles {
		roles = append(roles, role.Name)
	}

	return roles, nil
}

func (a *AccountService) GetAccountID(token string) (uint, error) {
	var result struct {
		ID uint `json:"id"`
	}

	resp, err := a.client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&result).
		Get("/api/Accounts/Me")

	if err != nil {
		return 0, err
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("failed to get account ID, status: %s", resp.Status())
	}

	return result.ID, nil
}
