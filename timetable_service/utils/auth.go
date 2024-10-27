package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

type HospitalService struct {
	client  *resty.Client
	baseURL string
}

type AccountService struct {
	client  *resty.Client
	baseURL string
}

func NewHospitalService() *HospitalService {
	baseURL := os.Getenv("HOSPITAL_SERVICE_URL")
	if baseURL == "" {
		log.Fatal("HOSPITAL_SERVICE_URL is not set")
	}
	client := resty.New()
	client.SetHostURL(baseURL)

	return &HospitalService{
		client:  client,
		baseURL: baseURL,
	}
}

func (h *HospitalService) ValidateHospitalAndRoom(hospitalID uint, roomName string, token string) (bool, error) {
	var hospital Hospital

	resp, err := h.client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&hospital).
		Get(fmt.Sprintf("/api/Hospitals/%d", hospitalID))

	if err != nil {
		return false, err
	}

	if resp.StatusCode() != 200 {
		return false, fmt.Errorf("failed to get hospital data: %s", resp.Status())
	}

	for _, room := range hospital.Rooms {
		if room.Name == roomName {
			return true, nil
		}
	}

	return false, fmt.Errorf("room '%s' not found in hospital '%s'", roomName, hospital.Name)
}

type Hospital struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Rooms   []Room `json:"rooms"`
}

type Room struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	HospitalID uint   `json:"hospitalId"`
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

func (a *AccountService) GetUserRoles(token string) ([]string, error) {
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

func (a *AccountService) GetUserID(token string) (uint, error) {
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
		return 0, fmt.Errorf("failed to get user ID, status: %s", resp.Status())
	}

	return result.ID, nil
}

type DoctorService struct {
	client  *resty.Client
	baseURL string
}

func NewDoctorService() *DoctorService {
	baseURL := os.Getenv("ACCOUNT_SERVICE_URL")
	if baseURL == "" {
		log.Fatal("ACCOUNT_SERVICE_URL is not set")
	}
	client := resty.New()
	client.SetHostURL(baseURL)

	return &DoctorService{
		client:  client,
		baseURL: baseURL,
	}
}

func (d *DoctorService) ValidateDoctor(doctorID uint, token string) (bool, error) {
	var result struct {
		ID uint `json:"id"`
	}

	resp, err := d.client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetResult(&result).
		Get(fmt.Sprintf("/api/Doctors/%d", doctorID))

	if err != nil {
		return false, err
	}

	if resp.StatusCode() != 200 {
		return false, fmt.Errorf("failed to validate doctor: %s", resp.Status())
	}

	return result.ID == doctorID, nil
}
