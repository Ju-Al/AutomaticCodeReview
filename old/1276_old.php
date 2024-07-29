    {
        parent::__construct($baseUrl);
    }

    /**
     * Fetch Links
     *
     * Fetches a set of links corresponding to an OpenURL
     *
     * @param string $openURL openURL (url-encoded)
     *
     * @return string         raw data returned by resolver
     */
    public function fetchLinks($openURL)
    {
        return '';
    }

    /**
     * Parse Links
     *
     * Parses an XML file returned by a link resolver
     * and converts it to a standardised format for display
     *
     * @param string $url Data returned by fetchLinks
     *
     * @return array         Array of values
     */
    public function parseLinks($url)
    {
        return [];
    }
}
