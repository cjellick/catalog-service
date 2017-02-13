package model

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/rancher/go-rancher/client"
)

type Template struct {
	EnvironmentId string `json:"environmentId"`
	CatalogId     uint   `sql:"type:integer REFERENCES catalog(id) ON DELETE CASCADE"`

	Name           string `json:"name"`
	IsSystem       string `json:"isSystem"`
	Description    string `json:"description"`
	DefaultVersion string `json:"defaultVersion" yaml:"version"`
	Path           string `json:"path"`
	Maintainer     string `json:"maintainer"`
	License        string `json:"license"`
	ProjectURL     string `json:"projectURL"`
	UpgradeFrom    string `json:"upgradeFrom"`
	FolderName     string `json:"folderName"`
	Catalog        string `json:"catalogId"`
	Base           string `json:"templateBase"`
	Icon           string `json:"icon"`
	IconFilename   string `json:"iconFilename"`
	Readme         string `json:"readme"`

	Categories []string          `sql:"-" json:"categories"`
	Labels     map[string]string `sql:"-" json:"labels"`

	Versions []Version `sql:"-"`
	Category string    `sql:"-"`
}

type TemplateModel struct {
	Base
	Template
}

type TemplateResource struct {
	client.Resource
	Template

	VersionLinks map[string]string `json:"versionLinks"`
}

type TemplateCollection struct {
	client.Collection
	Data []TemplateResource `json:"data,omitempty"`
}

func LookupTemplate(db *gorm.DB, environmentId, catalog, folderName, base string) *Template {
	var templateModel TemplateModel
	db.Raw(`
SELECT catalog_template.*
FROM catalog_template, catalog
WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
AND catalog_template.catalog_id = catalog.id
AND catalog.name = ?
AND catalog_template.base = ?
AND catalog_template.folder_name = ?
`, environmentId, "global", catalog, base, folderName).Scan(&templateModel)

	templateModel.Categories = lookupTemplateCategories(db, templateModel.ID)
	templateModel.Labels = lookupLabels(db, templateModel.ID)
	templateModel.Versions = lookupVersions(db, templateModel.ID)

	return &templateModel.Template
}

func LookupTemplates(db *gorm.DB, environmentId, catalog string, categories, categoriesNe []string) []Template {
	var templateModels []TemplateModel

	params := []interface{}{environmentId, "global"}
	if catalog != "" {
		params = append(params, catalog)
	}
	for _, category := range categories {
		params = append(params, category)
	}
	for _, categoryNe := range categoriesNe {
		params = append(params, categoryNe)
	}

	query := `
SELECT catalog_template.*
FROM catalog_template, catalog_template_category, catalog_category, catalog
WHERE (catalog_template.environment_id = ? OR catalog_template.environment_id = ?)
AND catalog_template.id = catalog_template_category.template_id
AND catalog_category.id = catalog_template_category.category_id
AND catalog_template.catalog_id = catalog.id`

	if catalog != "" {
		query += `
AND catalog.name = ?`
	}
	if len(categories) > 0 {
		query += fmt.Sprintf(`
AND catalog_category.name IN (%s)`, listQuery(len(categories)))
	}
	if len(categoriesNe) > 0 {
		query += fmt.Sprintf(`
AND catalog_category.name NOT IN (%s)`, listQuery(len(categoriesNe)))
	}

	db.Raw(query, params...).Find(&templateModels)

	var templates []Template
	for _, templateModel := range templateModels {
		templateModel.Categories = lookupTemplateCategories(db, templateModel.ID)
		templateModel.Labels = lookupLabels(db, templateModel.ID)
		templateModel.Versions = lookupVersions(db, templateModel.ID)
		templates = append(templates, templateModel.Template)
	}
	return templates
}

func listQuery(size int) string {
	var query string
	for i := 0; i < size; i++ {
		query += " ? ,"
	}
	return fmt.Sprintf("(%s)", strings.TrimSuffix(query, ","))
}
