package gcs

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/graymeta/stow"
)

// Kind represents the name of the location/storage type.
const Kind = "gcs"

const (
	// The service account json blob.
	ConfigJSON      = "json"
	ConfigProjectId = "project_id"
	ConfigEndpoint  = "endpoint"
	ConfigScopes    = "scopes"
)

func init() {
	validatefn := func(config stow.Config) error {
		_, ok := config.Config(ConfigJSON)
		if !ok {
			return errors.New("missing JSON configuration")
		}

		_, ok = config.Config(ConfigProjectId)
		if !ok {
			return errors.New("missing Project ID")
		}
		return nil
	}
	makefn := func(config stow.Config) (stow.Location, error) {
		_, ok := config.Config(ConfigJSON)
		if !ok {
			return nil, errors.New("missing JSON configuration")
		}

		_, ok = config.Config(ConfigProjectId)
		if !ok {
			return nil, errors.New("missing Project ID")
		}

		// Create a new client
		ctx, client, err := newGoogleStorageClient(config)
		if err != nil {
			return nil, err
		}

		// Create a location with given config and client
		loc := &Location{
			config: config,
			client: client,
			ctx:    ctx,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn, validatefn)
}

// Attempts to create a session based on the information given.
func newGoogleStorageClient(config stow.Config) (context.Context, *storage.Client, error) {
	json, _ := config.Config(ConfigJSON)

	scopes := []string{storage.ScopeFullControl}
	if s, ok := config.Config(ConfigScopes); ok && s != "" {
		scopes = strings.Split(s, ",")
	}

	endpoint := ""
	if s, ok := config.Config(ConfigEndpoint); ok && s != "" {
		endpoint = s
	}
	print("\nendpoint=", endpoint)

	ctx := context.Background()
	var creds *google.Credentials
	var err error
	if json != "" {
		creds, err = google.CredentialsFromJSON(ctx, []byte(json), scopes...)
		if err != nil {
			print("\nan error returned with credentials")
			print(err.Error())
			return nil, nil, err
		}
	} else {
		print("\nattempting to use default credentials for google cloud storage")
		creds, err = google.FindDefaultCredentials(ctx, scopes...)
		if err != nil {
			print("\nfailed to fetch default credentials")
			return nil, nil, err
		}
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		print("\nerror while creating storage client")
		return nil, nil, err
	}

	print("\ncontext and client has been created")
	return ctx, client, nil
}
