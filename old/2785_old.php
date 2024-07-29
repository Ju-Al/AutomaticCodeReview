
		return $this->cache_path . $filename;
	}

	/**
	 * Writes the content to a file
	 *
	 * @since 3.1
	 *
	 * @param string $content Content to write.
	 * @param string $file    Path to the file to write in.
	 * @return bool
	 */
	protected function write_file( $content, $file ) {
		if ( $this->filesystem->is_readable( $file ) ) {
			return true;
		}

		if ( ! rocket_mkdir_p( dirname( $file ) ) ) {
			return false;
		}

		return rocket_put_content( $file, $content );
	}
}
