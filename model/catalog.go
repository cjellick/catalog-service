package model

import (
	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

type Catalog struct {
	EnvironmentId string `json:"environmentId"`

	Name   string `json:"name"`
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Type   string `json:"type"`
}

type CatalogModel struct {
	Base
	Catalog
}

type CatalogResource struct {
	client.Resource
	Catalog
}

type CatalogCollection struct {
	client.Collection
	Data []CatalogResource `json:"data,omitempty"`
}

func GetCatalog(db *gorm.DB, id uint) *Catalog {
	var catalogModel CatalogModel
	db.First(&catalogModel, id)
	return &catalogModel.Catalog
}

func LookupCatalog(db *gorm.DB, environmentId, name string) *Catalog {
	var catalogModel CatalogModel
	if err := db.Where(&CatalogModel{
		Catalog: Catalog{
			Name: name,
		},
	}).Where("environment_id = ? OR environment_id = ?", environmentId, "global").First(&catalogModel).Error; err == gorm.ErrRecordNotFound {
		return nil
	}
	return &catalogModel.Catalog
}

func LookupCatalogs(db *gorm.DB, environmentId string) []Catalog {
	var catalogModels []CatalogModel
	db.Where("environment_id = ? OR environment_id = ?", environmentId, "global").Find(&catalogModels)
	var catalogs []Catalog
	for _, catalogModel := range catalogModels {
		catalogs = append(catalogs, catalogModel.Catalog)
	}
	return catalogs
}

// TODO: return error
func DeleteCatalog(db *gorm.DB, environmentId, name string) {
	db.Where(&CatalogModel{
		Catalog: Catalog{
			Name: name,
		},
	}).Where("environment_id = ? OR environment_id = ?", environmentId, "global").Delete(&CatalogModel{})
}
