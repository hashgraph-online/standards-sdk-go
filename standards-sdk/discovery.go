package standardssdk

const (
	// PackageName is the stable package identifier.
	PackageName = "standardssdk"
	// ImportPath is the canonical import path for this package.
	ImportPath = "github.com/hashgraph-online/standards-sdk-go/standards-sdk"
)

// Identity returns the package import path.
func Identity() string {
	return ImportPath
}
