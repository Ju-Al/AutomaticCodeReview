    }

    /**
     * @param string                                     $fileIdentifier    can be a tl_files uuid/id/path or an absolute path
     * @param int|string|array|PictureConfiguration|null $sizeConfiguration a picture size configuration or reference
     * @param string|null                                $alt               overwrite alt attribute, set to null for defaults
     */
    public function createPicture(string $fileIdentifier, $sizeConfiguration, string $alt = null): array
    {
        $this->framework->initialize();

        $filePath = $this->getFilepath($fileIdentifier, $filesModel);

        $picture = $this->pictureFactory->create($filePath, $sizeConfiguration);

        return [
            self::PICTURE_IMAGE => $picture->getImg($this->rootDir, $this->staticUrl),
            self::PICTURE_SOURCES => $picture->getSources($this->rootDir, $this->staticUrl),
            self::PICTURE_ALT => $alt ?? $this->getAltAttribute($filesModel),
        ];
    }

    public function createImage(string $fileIdentifier, $sizeConfiguration): array
    {
        // todo
        throw new \RuntimeException('not implemented');
    }

    /**
     * Try to locate a file by querying the DBAFS ($identifier = uuid/id/path),
     * fallback to interpret $identifier as absolute/relative file path.
     */
    private function getFilepath(string $identifier, FilesModel &$filesModel = null): string
    {
        /** @var Validator $validatorAdapter */
        $validatorAdapter = $this->framework->getAdapter(Validator::class);

        /** @var FilesModel $filesModelAdapter */
        $filesModelAdapter = $this->framework->getAdapter(FilesModel::class);

        if ($validatorAdapter->isUuid($identifier)) {
            $filesModel = $filesModelAdapter->findByUuid($identifier);
            $dbafsItem = true;
        } elseif (is_numeric($identifier)) {
            $filesModel = $filesModelAdapter->findById((int) $identifier);
            $dbafsItem = true;
        } else {
            $filesModel = $filesModelAdapter->findByPath($identifier);
            $dbafsItem = null !== $filesModel;
        }

        if ($dbafsItem) {
            if (null === $filesModel) {
                throw new \InvalidArgumentException("DBAFS item '$identifier' could not be found.");
            }

            if ('file' !== $filesModel->type) {
                throw new \InvalidArgumentException("DBAFS item '$identifier' was found but is not a file.");
            }

            return Path::makeAbsolute($filesModel->path, $this->rootDir);
        }

        if (Path::isAbsolute($identifier)) {
            return Path::canonicalize($identifier);
        }

        return Path::makeAbsolute($identifier, $this->rootDir);
    }

    private function getAltAttribute(?FilesModel $filesModel): ?string
    {
        if (null === $filesModel) {
            return null;
        }

        /** @var Frontend $frontend */
        $frontend = $this->framework->getAdapter(Frontend::class);

        $metaData = $frontend->getMetaData($filesModel->meta, $this->locale);

        return $metaData['alt'] ?? null;
    }
}
