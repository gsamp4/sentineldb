package domain

import (
	"fmt"
	"net"
	"regexp"
	"sentineldb/internal/job/models"

	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AssetRepositoryInterface interface {
    RegisterAsset(asset *models.Asset) error
}

type AssetRepository struct {
    DB *gorm.DB
}

func (a AssetRepository) RegisterAsset(asset *models.Asset) error {
    return a.DB.Create(asset).Error
}

func ValidateAsset(assetType string, value string) error {
	// Validate asset type
	if assetType != "ip" && assetType != "domain" && assetType != "email" {
		return fmt.Errorf("invalid asset type: %s", assetType)
	}

	if assetType == "ip" && net.ParseIP(value) == nil {
		return fmt.Errorf("invalid IP address: %s", value)
	}

	if assetType == "email" && !isValidEmail(value) {
		return fmt.Errorf("invalid email address: %s", value)
	}

	if assetType == "domain" {
		isValid := govalidator.IsDNSName(value)
		if !isValid {
			return fmt.Errorf("invalid domain name: %s", value)
		}
	}
	return  nil
}

func isValidEmail(email string) bool {
    return emailRegex.MatchString(email)
}