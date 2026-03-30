// Package recallkey 定义 offline 召回器分类与标识。
package recallkey

import "strings"

const (
	CategoryNonPersonal = "non_personal"
	CategoryCF          = "cf"
	CategoryExternal    = "external"
)

const (
	NameLatest     = "latest"
	NamePopular    = "popular"
	NameUserToUser = "user_to_user"
	NameItemToItem = "item_to_item"
	NameMF         = "mf"
)

const (
	RecallerLatest     = CategoryNonPersonal + "/" + NameLatest
	RecallerPopular    = CategoryNonPersonal + "/" + NamePopular
	RecallerUserToUser = CategoryCF + "/" + NameUserToUser
	RecallerItemToItem = CategoryCF + "/" + NameItemToItem
	RecallerMF         = CategoryCF + "/" + NameMF
)

// Key 返回统一召回器标识。
func Key(category, name string) string {
	category = strings.TrimSpace(category)
	name = strings.TrimSpace(name)
	if category == "" {
		return name
	}
	if name == "" {
		return category
	}
	return category + "/" + name
}
