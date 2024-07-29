		);
	}

	protected function setup_hook( $tag, $callback, $priority = 10, $num_args = 1 ) {
		$hook = new WP_Hook();
		$hook->add_filter( $tag, $callback, $priority, $num_args );

		return $hook;
	}
}
