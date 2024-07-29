			return [];
		}

		return [
			'rocket_lazyload_background_images' => 'disable_lazyload_background_images',
		];
	}

	/**
	 * Disable lazyload for background images when using Bridge theme
	 *
	 * @since 3.3.4
	 * @author Remy Perona
	 *
	 * @return bool
	 */
	public function disable_lazyload_background_images() {
		return false;
	}
}
