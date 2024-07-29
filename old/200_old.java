
  /**
   * Keep an existing manifest as part of the table.
   * <p>
   * The manifest must be part of the table's current snapshot.
   *
   * @param file a manifest file
   * @return this for method chaining
   */
  RewriteManifests keepManifest(ManifestFile file);

}
