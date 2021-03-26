package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	gcrauthn "github.com/google/go-containerregistry/pkg/authn"
	gcrname "github.com/google/go-containerregistry/pkg/name"
	gcrgoogle "github.com/google/go-containerregistry/pkg/v1/google"
)

type manifest struct {
	Digest string
	Info   gcrgoogle.ManifestInfo
}

var (
	stdout = os.Stdout
	stderr = os.Stderr

	tokenPtr = flag.String("token", os.Getenv("GCRCLEANER_TOKEN"), "Authentication token")
	repoPtr  = flag.String("repo", "", "Repository name")
)

func main() {
	flag.Parse()
	if err := realMain(); err != nil {
		fmt.Fprintf(stderr, "%s\n", err)
		os.Exit(1)
	}
}

func realMain() error {

	// Try to find the best "authentication

	if *repoPtr == "" {
		*repoPtr = "asia.gcr.io/tikivn/xlearning"
		// return fmt.Errorf("missing -repo")
	}

	var auther gcrauthn.Authenticator
	if *tokenPtr != "" {

		auther = &gcrauthn.Bearer{Token: *tokenPtr}
	} else {
		var err error
		auther, err = gcrgoogle.NewEnvAuthenticator()

		if err != nil {
			return fmt.Errorf("failed to setup auther: %w", err)
		}
	}

	// create a new repository
	gcrrepo, err := gcrname.NewRepository(*repoPtr)
	if err != nil {
		return fmt.Errorf("failed to get repo %v: %w", *repoPtr, err)
	}

	// list tags
	tags, err := gcrgoogle.List(gcrrepo, gcrgoogle.WithAuth(auther))
	if err != nil {
		return fmt.Errorf("failed to list tags for repo %v: %w", *repoPtr, err)
	}

	var manifests = make([]manifest, 0, len(tags.Manifests))
	for k, m := range tags.Manifests {
		manifests = append(manifests, manifest{k, m})
	}

	sort.Slice(manifests, func(i, j int) bool {
		return manifests[j].Info.Created.Before(manifests[i].Info.Created)
	})

	fmt.Println(manifests)
	return nil
}
