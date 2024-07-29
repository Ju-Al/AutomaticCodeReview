	 */
	private function get_url_contents( $url ) {
		$external_url = $this->is_external_file( $url );
		$file_path    = $external_url ? $this->local_cache->get_filepath( $url ) : $this->get_file_path( $url );

		if ( empty( $file_path ) ) {
			Logger::error(
				'Couldn’t get the file path from the URL.',
				[
					'RUCSS warmup process',
					'url' => $url,
				]
			);

			return false;
		}

		$file_content = $external_url ? $this->local_cache->get_content( $url ) : $this->get_file_content( $file_path );

		if ( ! $file_content ) {
			Logger::error(
				'No file content.',
				[
					'RUCSS warmup process',
					'path' => $file_path,
				]
			);

			return false;
		}

		return $file_content;
	}

	/**
	 * Gets the CDN zones.
	 *
	 * @since  3.1
	 *
	 * @return array
	 */
	public function get_zones() {
		return [ 'all', 'css_and_js' ];
	}

}
