package ep

import (
	"fmt"

	"github.com/xumak-grid/bedrock/awscli"
)

const (
	// DefaultExtensionVersion represents the default new version to the init package
	DefaultExtensionVersion = "0.0.0-SNAPSHOT"
	gridEPBucket            = "grid-ep-packages"
)

// InitPackage represents an initial source package to start an EP project
type InitPackage struct {
	// EP init package version (7.1)
	Version string
	// the version that InitPackage is using in poms (701.0.0-SNAPSHOT)
	PlatformVersion string
	// bucket where this package is stored (grid-ep-packages)
	Bucket string
	// key to locate the object in the s3 (construction/EP-Commerce-7.1.0.zip)
	Key string
}

// InitPackages returns the ep initial packages available in S3
// since this packages do not change over the time when a new package
// is uploaded to S3 is necessary to add it here.
func InitPackages() []InitPackage {

	return []InitPackage{
		InitPackage{
			Version:         "7.1",
			PlatformVersion: "701.0.0-SNAPSHOT",
			Bucket:          gridEPBucket,
			Key:             "construction/EP-Commerce-7.1.0.zip",
		},
	}
}

// PreSignedURL creates a new pre-signed url for an InitPackage with 1 hour expiration
func (p *InitPackage) PreSignedURL() (string, error) {

	sess, err := awscli.Session()
	if err != nil {
		return "", err
	}
	s3obj := awscli.NewS3Object(p.Bucket, p.Key)

	return s3obj.PreSignedURL(sess, 1)
}

// FindInitPackage finds an init package version from the InitPackages list
func FindInitPackage(version string) (InitPackage, error) {

	for _, pk := range InitPackages() {
		if pk.Version == version {
			return pk, nil
		}
	}
	return InitPackage{}, fmt.Errorf("not found init package '%v' EP version", version)
}
