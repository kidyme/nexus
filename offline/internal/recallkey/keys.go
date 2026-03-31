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
	NameItemToItem = "item_to_item"
	NameUserToUser = "user_to_user"
	NameMF         = "mf"
)

const (
	RecallerLatest  = CategoryNonPersonal + "/" + NameLatest
	RecallerPopular = CategoryNonPersonal + "/" + NamePopular
)

// Key 返回统一召回器标识。
func Key(parts ...string) string {
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return strings.Join(result, "/")
}

// CFItemToItemRecaller 返回 item-to-item 召回器实例标识。
func CFItemToItemRecaller(recallerType string) string {
	return Key(CategoryCF, NameItemToItem, recallerType)
}
