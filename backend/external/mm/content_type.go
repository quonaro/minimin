package mm

// ContentType categorizes the kind of content a source can provide.
type ContentType string

const (
	ContentMod          ContentType = "mod"
	ContentResourcepack ContentType = "resourcepack"
	ContentShaderpack   ContentType = "shaderpack"
)

// AllContentTypes returns every supported content type.
func AllContentTypes() []ContentType {
	return []ContentType{ContentMod, ContentResourcepack, ContentShaderpack}
}

// IsValidContentType reports whether v is a known content type.
func IsValidContentType(v string) bool {
	switch ContentType(v) {
	case ContentMod, ContentResourcepack, ContentShaderpack:
		return true
	}
	return false
}
