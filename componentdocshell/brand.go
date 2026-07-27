package componentdocshell

import shellassets "github.com/araihu/goshtoso-app-shells/componentdocshell/assets"

// GoshtosoBrand returns an optional canonical Goshtoso brand configuration.
// Other consumers remain free to construct Brand directly.
func GoshtosoBrand(name, homeURL, assetPrefix string) Brand {
	return Brand{
		Name:       name,
		HomeURL:    homeURL,
		Logo:       goshtosoBrandMark(assetPrefix),
		FaviconURL: shellassets.GoshtosoFaviconURL(assetPrefix),
	}
}
