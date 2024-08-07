<?php
declare(strict_types=1);

namespace WP_Rocket\Engine\Optimization\RUCSS\Warmup;

use WP_Rocket\Engine\Optimization\AssetsLocalCache;
use WP_Rocket\Engine\Optimization\RegexTrait;
use WP_Rocket\Engine\Optimization\UrlTrait;
use WP_Rocket\Logger\Logger;
use WP_Rocket_WP_Async_Request;

class ResourceFetcher extends WP_Rocket_WP_Async_Request {

	use RegexTrait, UrlTrait;

	const LINK_PATTERN = '<link\s+(?:[^>]+[\s"\'])?href\s*=\s*[\'"]\s*(?<url>[^\'"\s]+)\s*?[\'"](?:[^>]+)?\/?>';
	const SCRIPT_PATTERN = '<script\s+(?:[^>]+[\s\'"])?src\s*=\s*[\'"]\s*?(?<url>[^\'"\s]+)\s*?[\'"](?:[^>]+)?\/?>';

	/**
	 * Prefix
	 *
	 * @var string
	 */
	protected $prefix = 'rocket';

	/**
	 * Action
	 *
	 * @var string
	 */
	protected $action = 'saas_warmup';

	/**
	 * Resources array.
	 *
	 * @var array
	 */
	private $resources = [];

	/**
	 * Assets local cache instance
	 *
	 * @var AssetsLocalCache
	 */
	private $local_cache;

	/**
	 * Resource fetcher process instance.
	 *
	 * @var ResourceFetcherProcess
	 */
	private $process;

	/**
	 * Resource constructor.
	 *
	 * @param AssetsLocalCache       $local_cache Local cache instance.
	 * @param ResourceFetcherProcess $process     Resource fetcher process instance.
	 */
	public function __construct( AssetsLocalCache $local_cache, ResourceFetcherProcess $process ) {
		parent::__construct();

		$this->local_cache = $local_cache;
		$this->process     = $process;
	}

	/**
	 * Handle Collecting resources and save them into the DB
	 *
	 * @since 3.9
	 */
	public function handle() {
		// Grab resources from page HTML.
		$html = ! empty( $_POST['html'] ) ? wp_unslash( $_POST['html'] ) : '';// phpcs:ignore WordPress.Security.NonceVerification.Missing, WordPress.Security.ValidatedSanitizedInput.InputNotSanitized

		if ( empty( $html ) ) {
			return;
		}

		$this->find_resources( $html, 'css' );
		$this->find_resources( $html, 'js' );

		if ( empty( $this->resources ) ) {
			return;
		}

		// Send resources to the background process to be saved into DB.
		foreach ( $this->resources as $resource ) {
			$this->process->push_to_queue( $resource );
		}

		$this->process->save()->dispatch();
	}

	/**
	private function find_styles( $html ) {
		$links = $this->find( '<link\s+(?:[^>]+[\s"\'])?href\s*=\s*[\'"]\s*(?<url>[^\'"\s]+)\s*?[\'"](?:[^>]+)?\/?>', $html );
		if ( empty( $links ) ) {
		foreach ( $links as $link ) {
			if ( ! (bool) preg_match( '/rel=[\'"]stylesheet[\'"]/is', $link[0] ) ) {
			$contents = $this->get_url_contents( $link['url'] );
				'url'     => rocket_add_url_protocol( $link['url'] ),
				'type'    => 'css',
	 * Find scripts with src on the page HTML.
	 * @param string $html Page HTML.
	private function find_scripts( $html ) {
		$scripts = $this->find( '<script\s+(?:[^>]+[\s\'"])?src\s*=\s*[\'"]\s*?(?<url>[^\'"\s]+)\s*?[\'"](?:[^>]+)?\/?>', $html );

		if ( empty( $scripts ) ) {
			return;
		}

		foreach ( $scripts as $script ) {
			$contents = $this->get_url_contents( $script['url'] );

			if ( empty( $contents ) ) {
				continue;
			}

			$this->resources[] = [
				'url'     => rocket_add_url_protocol( $script['url'] ),
				'content' => $contents,
				'type'    => 'js',
			];
		}
	 * Find link styles on the page HTML.
	 *
	 * @since 3.9
	 *
	 * @param string $html Page HTML.
	 * @param string $type Specify to look for 'css' or 'js' resources.
	 *
	 * @return void
	 */
	private function find_resources( string $html, string $type ) {
		$pattern = ( 'css' === $type) ? SELF::LINK_PATTERN : SELF::SCRIPT_PATTERN;

		$resources = $this->find( $pattern, $html );

		if ( empty( $resources ) ) {
			return;
		}

		foreach ( $resources as $resource ) {

			if ( 'css' === $type && ! $this->is_link_stylesheet( $resource[0] ) ) {
				continue;
			}

			$contents = $this->get_url_contents( $resource['url'] );

			if ( empty( $contents ) ) {
				continue;
			}

			$this->resources[] = [
				'url'     => rocket_add_url_protocol( $resource['url'] ),
				'content' => $contents,
				'type'    => $type,
			];
		}
	}

	/**
	 * Check that a link element is a stylesheet.
	 *
	 * @since 3.9
	 *
	 * @param string $link The link element to check.
	 *
	 * @return bool True for stylesheet; false for anything else.
	 */
	private function is_link_stylesheet( string $link ) {
		return (bool) preg_match( '/rel=[\'"]stylesheet[\'"]/is', $link );
	}

	/**
	 * Get url file contents.
	 *
	 * @param string $url File url.
	 *
	 * @return string
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

			return '';
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

			return '';
		}

		return $file_content;
	}

	/**
	 * Gets the CDN zones.
	 *
	 * @return array
	 */
	public function get_zones() {
		return [ 'all', 'css_and_js' ];
	}

}
