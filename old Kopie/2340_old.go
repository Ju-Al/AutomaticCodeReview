package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/spire/pkg/common/bundleutil"
	"github.com/spiffe/spire/pkg/server/plugin/datastore"
	"github.com/zeebo/errs"
)

type BundleUpdaterConfig struct {
	TrustDomainConfig

	TrustDomain spiffeid.TrustDomain
	DataStore   datastore.DataStore

	// newClient is a test hook for injecting client behavior
	newClient func(ClientConfig) (Client, error)
}

type BundleUpdater interface {
	// UpdateBundle fetches the local bundle from the datastore and the
	// endpoint bundle from the endpoint. The function will return an error if
	// the local bundle cannot be fetched, the endpoint bundle cannot be
	// downloaded, or there is a problem persisting the bundle. The local
	// bundle will always be returned if it was fetched, independent of any
	// other failures performing the update. The endpoint bundle is ONLY
	// returned if it can be successfully downloaded, is different from the
	// local bundle, and is successfully stored.
	UpdateBundle(ctx context.Context) (*bundleutil.Bundle, *bundleutil.Bundle, error)

	// TrustDomainConfig returns the configuration for the updater
	TrustDomainConfig() TrustDomainConfig
}

type bundleUpdater struct {
	c BundleUpdaterConfig
}

func NewBundleUpdater(config BundleUpdaterConfig) BundleUpdater {
	if config.newClient == nil {
		config.newClient = NewClient
	}

	return &bundleUpdater{
		c: config,
	}
}

func (u *bundleUpdater) UpdateBundle(ctx context.Context) (*bundleutil.Bundle, *bundleutil.Bundle, error) {
		return nil, nil, fmt.Errorf("failed to fetch local bundle: %w", err)
	client, err := u.newClient(localBundleOrNil)
		return nil, nil, err
	endpointBundle, err := client.FetchBundle(ctx)
		return localBundleOrNil, nil, fmt.Errorf("failed to fetch endpoint bundle: %w", err)
	if localBundleOrNil != nil && endpointBundle.EqualTo(localBundleOrNil) {
		return localBundleOrNil, nil, nil
	_, err = u.c.DataStore.SetBundle(ctx, endpointBundle.Proto())
		return localBundleOrNil, nil, fmt.Errorf("failed to store endpoint bundle: %w", err)
	return localBundleOrNil, endpointBundle, nil
func (u *bundleUpdater) newClient(localBundleOrNil *bundleutil.Bundle) (Client, error) {
		TrustDomain:     u.c.TrustDomain,
		EndpointAddress: u.c.EndpointAddress,
	if !u.c.UseWebPKI {
		if localBundleOrNil == nil {
			return nil, errs.New("local bundle not found")
	localBundleOrNil, err := fetchBundleIfExists(ctx, u.c.DataStore, u.c.TrustDomain)
	if err != nil {
		return nil, nil, err
	}

	localFederatedBundleOrNil, err := fetchBundleIfExists(ctx, u.c.DataStore, u.c.TrustDomain)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch local federated bundle: %w", err)
	}

	fetchedFederatedBundle, err := client.FetchBundle(ctx)
	if err != nil {
		return localFederatedBundleOrNil, nil, fmt.Errorf("failed to fetch federated bundle from endpoint: %w", err)
	}

	if localFederatedBundleOrNil != nil && fetchedFederatedBundle.EqualTo(localFederatedBundleOrNil) {
		return localFederatedBundleOrNil, nil, nil
	}

	_, err = u.c.DataStore.SetBundle(ctx, fetchedFederatedBundle.Proto())
	if err != nil {
		return localFederatedBundleOrNil, nil, fmt.Errorf("failed to store fetched federated bundle: %w", err)
	}

	return localFederatedBundleOrNil, fetchedFederatedBundle, nil
}

func (u *bundleUpdater) TrustDomainConfig() TrustDomainConfig {
	return u.c.TrustDomainConfig
}

func (u *bundleUpdater) newClient(ctx context.Context) (Client, error) {
	config := ClientConfig{
		TrustDomain:      u.c.TrustDomain,
		EndpointURL:      u.c.EndpointURL,
		DeprecatedConfig: u.c.DeprecatedConfig,
	}

	if spiffeAuth, ok := u.c.EndpointProfile.(HTTPSSPIFFEProfile); ok {
		trustDomain := spiffeAuth.EndpointSPIFFEID.TrustDomain()

		// This is to preserve behavioral compatibility when using
		// the deprecated config and will be removed in 1.1.0.
		if u.c.DeprecatedConfig && trustDomain.IsZero() {
			trustDomain = u.c.TrustDomain
		}
		localEndpointBundle, err := fetchBundleIfExists(ctx, u.c.DataStore, trustDomain)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch local bundle: %w", err)
		}

		if localEndpointBundle == nil {
			return nil, errors.New("can't perform SPIFFE Authentication: local bundle not found")
		}
		config.SPIFFEAuth = &SPIFFEAuthConfig{
			EndpointSpiffeID: spiffeAuth.EndpointSPIFFEID,
			RootCAs:          localEndpointBundle.RootCAs(),
		}
	}
	return u.c.newClient(config)
}

func fetchBundleIfExists(ctx context.Context, ds datastore.DataStore, trustDomain spiffeid.TrustDomain) (*bundleutil.Bundle, error) {
	// Load the current bundle and extract the root CA certificates
	bundle, err := ds.FetchBundle(ctx, trustDomain.IDString())
	if err != nil {
		return nil, errs.Wrap(err)
	}
	if bundle == nil {
		return nil, nil
	}
	return bundleutil.BundleFromProto(bundle)
}
