                    }
                break;
                default:
                    throw new \InvalidArgumentException(sprintf('Unknown asset source option "%s" specified for Neos asset source "%s". Please check your settings.', $optionName, $assetSourceIdentifier), 1513327774584);
            }
        }
    }

    /**
     * @param string $assetSourceIdentifier
     * @param array $assetSourceOptions
     * @return AssetSource
     */
    public static function createFromConfiguration(string $assetSourceIdentifier, array $assetSourceOptions): AssetSource
    {
        return new static($assetSourceIdentifier, $assetSourceOptions);
    }

    /**
     * @return string
     */
    public function getIdentifier(): string
    {
        return $this->assetSourceIdentifier;
    }

    /**
     * @return string
     */
    public function getLabel(): string
    {
        return 'Neos';
    }

    /**
     * @return AssetProxyRepository
     */
    public function getAssetProxyRepository(): AssetProxyRepository
    {
        if ($this->assetProxyRepository === null) {
            $this->assetProxyRepository = new NeosAssetProxyRepository($this);
        }

        return $this->assetProxyRepository;
    }

    /**
     * @return bool
     */
    public function isReadOnly(): bool
    {
        return false;
    }

    /**
     * Internal method used by NeosAssetProxy
     *
     * @param AssetInterface $asset
     * @return Uri|null
     * @throws ThumbnailServiceException
     * @throws AssetServiceException
     */
    public function getThumbnailUriForAsset(AssetInterface $asset): ?UriInterface
    {
        $actionRequest = ($this->asyncThumbnails ? $this->createActionRequest() : null);
        $thumbnailConfiguration = $this->thumbnailService->getThumbnailConfigurationForPreset('Neos.Media.Browser:Thumbnail', ($actionRequest !== null));
        $thumbnailData = $this->assetService->getThumbnailUriAndSizeForAsset($asset, $thumbnailConfiguration, $actionRequest);
        return isset($thumbnailData['src']) ? new Uri($thumbnailData['src']) : null;
    }

    /**
     * Internal method used by NeosAssetProxy
     *
     * @param AssetInterface $asset
     * @return Uri|null
     * @throws ThumbnailServiceException
     * @throws AssetServiceException
     */
    public function getPreviewUriForAsset(AssetInterface $asset): ?UriInterface
    {
        $actionRequest = ($this->asyncThumbnails ? $this->createActionRequest() : null);
        $thumbnailConfiguration = $this->thumbnailService->getThumbnailConfigurationForPreset('Neos.Media.Browser:Preview', ($actionRequest !== null));
        $thumbnailData = $this->assetService->getThumbnailUriAndSizeForAsset($asset, $thumbnailConfiguration, $actionRequest);
        return isset($thumbnailData['src']) ? new Uri($thumbnailData['src']) : null;
    }

    /**
     * @return ActionRequest|null
     */
    private function createActionRequest(): ?ActionRequest
    {
        $requestHandler = $this->bootstrap->getActiveRequestHandler();
        if ($requestHandler instanceof HttpRequestHandlerInterface) {
            return new ActionRequest($requestHandler->getHttpRequest());
        }
        return null;
    }
}
