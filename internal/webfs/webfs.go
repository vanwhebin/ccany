package webfs

import (
	"embed"
	"io/fs"
)

// Embed web directory and locales
//
//go:embed web/* locales/*
var webFS embed.FS

// GetWebFS returns the embedded web filesystem
func GetWebFS() fs.FS {
	// Return web subdirectory
	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	return webSubFS
}

// GetStaticFS returns the static filesystem
func GetStaticFS() fs.FS {
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		panic(err)
	}
	return staticFS
}

// GetLocalesFS returns the locales filesystem
func GetLocalesFS() fs.FS {
	localesFS, err := fs.Sub(webFS, "locales")
	if err != nil {
		panic(err)
	}
	return localesFS
}

// GetIndexHTML returns index.html content
func GetIndexHTML() ([]byte, error) {
	return webFS.ReadFile("web/index.html")
}

// GetSetupHTML returns setup.html content
func GetSetupHTML() ([]byte, error) {
	return webFS.ReadFile("web/setup.html")
}

// GetFavicon returns favicon.ico content
func GetFavicon() ([]byte, error) {
	return webFS.ReadFile("web/favicon.ico")
}

// GetLocaleFile returns locale file content
func GetLocaleFile(lang string) ([]byte, error) {
	return webFS.ReadFile("locales/" + lang + ".json")
}

// GetAvailableLocales returns list of available locales
func GetAvailableLocales() ([]string, error) {
	localesFS := GetLocalesFS()
	entries, err := fs.ReadDir(localesFS, ".")
	if err != nil {
		return nil, err
	}

	var locales []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "" {
			// Remove .json extension
			name := entry.Name()
			if len(name) > 5 && name[len(name)-5:] == ".json" {
				locales = append(locales, name[:len(name)-5])
			}
		}
	}

	return locales, nil
}
