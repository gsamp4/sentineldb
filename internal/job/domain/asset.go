package domain

import (
	"fmt"
	"net"
	"regexp"
	"sentineldb/internal/job/models"
	"sentineldb/pkg/logger"

	"github.com/asaskevich/govalidator"
	"github.com/oklog/ulid/v2"
	"gorm.io/gorm"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type AssetRepositoryInterface interface {
    RegisterAsset(asset *models.Asset) error
    ListAssets() ([]models.Asset, error)
    GetAssetByID(id string) (*models.Asset, error)
    UpdateAsset(id string, label *string, active *bool) error
    SoftDeleteAsset(id string) error
}


type AssetRepository struct {
	DB *gorm.DB
	Logger *logger.Logger
}

func (a AssetRepository) RegisterAsset(asset *models.Asset) error {
    fmt.Println("DB is nil:", a.DB == nil)
    asset.ID = ulid.Make().String()
    err := a.DB.Create(asset).Error
    fmt.Println("REGISTER ERROR:", err)  // log aqui
    return err
}

func (a AssetRepository) ListAssets() ([]models.Asset, error) {
	var assets []models.Asset
	if err := a.DB.Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (a AssetRepository) GetAssetByID(id string) (*models.Asset, error) {
	var asset models.Asset
	if err := a.DB.First(&asset, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &asset, nil
}

func (a AssetRepository) UpdateAsset(id string, label *string, active *bool) error {
	updates := map[string]interface{}{}
	if label != nil {
		updates["label"] = *label
	}
	if active != nil {
		updates["active"] = *active
	}
	if len(updates) == 0 {
		return nil
	}
	res := a.DB.Model(&models.Asset{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (a AssetRepository) SoftDeleteAsset(id string) error {
	res := a.DB.Model(&models.Asset{}).Where("id = ?", id).Update("active", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("not found")
	}
	return nil
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