        $this->disableSessionWrites();  // avoid session write timing bug
        // Send both GET and POST variables to search class:
        $request = $params->fromQuery() + $params->fromPost();

        $rManager = $this->recommendPluginManager;
        $setupCallback = function ($runner, $params, $searchId) use ($rManager) {
            $listener = new RecommendListener($rManager, $searchId);
            $config = [];
            $rawConfig = $params->getOptions()
                ->getRecommendationSettings($params->getSearchHandler());
            foreach ($rawConfig['side'] as $value) {
                $settings = explode(':', $value);
                if ($settings[0] === 'SideFacetsDeferred') {
                    $settings[0] = 'SideFacets';
                    $config['side'][] = implode(':', $settings);
                }
            }
            $listener->setConfig($config);
            $listener->attach($runner->getEventManager()->getSharedManager());

            $params->setLimit(0);
            if (is_callable([$params, 'getHierarchicalFacetLimit'])) {
                $params->setHierarchicalFacetLimit(-1);
            }
            $options = $params->getOptions();
            $options->disableHighlighting();
            $options->spellcheckEnabled(false);
        };

        $runner = $this->searchRunner;
        $results = $runner->run($request, DEFAULT_SEARCH_BACKEND, $setupCallback);

        if ($results instanceof \VuFind\Search\EmptySet\Results) {
            $this->logError('Faceting request failed');
            return $this->formatResponse('', self::STATUS_HTTP_ERROR);
        }

        $recommend = $results->getRecommendations('side');
        $recommend = reset($recommend);

        $context = [
            'recommend' => $recommend,
            'params' => $results->getParams(),
            'searchClassId' => DEFAULT_SEARCH_BACKEND
        ];
        if (isset($request['enabledFacets'])) {
            // Render requested facets separately
            $response = [];
            $hierarchicalFacets = [];
            $options = $results->getOptions();
            if (is_callable([$options, 'getHierarchicalFacets'])) {
                $hierarchicalFacets = $options->getHierarchicalFacets();
                $hierarchicalFacetSortOptions
                    = $recommend->getHierarchicalFacetSortOptions();
            }
            $checkboxFacets = $results->getParams()->getCheckboxFacets();
            $sideFacetSet = $recommend->getFacetSet();
            $results = $recommend->getResults();
            foreach ($request['enabledFacets'] as $facet) {
                if (strpos($facet, ':')) {
                    foreach ($checkboxFacets as $checkboxFacet) {
                        if ($facet !== $checkboxFacet['filter']) {
                            continue;
                        }
                        list($field, $value) = explode(':', $facet, 2);
                        $checkboxResults = $results->getFacetList(
                            [$field => $value]
                        );
                        if (!isset($checkboxResults[$field]['list'])) {
                            $response[$facet]['checkboxCount'] = 0;
                            continue 2;
                        }
                        $count = 0;
                        $truncate = substr($value, -1) === '*';
                        if ($truncate) {
                            $value = substr($value, 0, -1);
                        }
                        foreach ($checkboxResults[$field]['list'] as $item) {
                            if ($item['value'] == $value
                                || ($truncate
                                && preg_match('/^' . $value . '/', $item['value']))
                                || ($item['value'] == 'true' && $value == '1')
                                || ($item['value'] == 'false' && $value == '0')
                            ) {
                                $count += $item['count'];
                            }
                        }
                        $response[$facet]['checkboxCount'] = $count;
                        continue 2;
                    }
                }
                if (in_array($facet, $hierarchicalFacets)) {
                    // Return the facet data for hierarchical facets
                    $facetList = $sideFacetSet[$facet]['list'];

                    if (!empty($hierarchicalFacetSortOptions[$facet])) {
                        $this->facetHelper->sortFacetList(
                            $facetList,
                            'top' === $hierarchicalFacetSortOptions[$facet]
                        );
                    }

                    $facetList = $this->facetHelper->buildFacetArray(
                        $facet, $facetList, $results->getUrlQuery(), false
                    );

                    if (!empty($this->facetConfig->FacetFilters->$facet)
                        || !empty($this->facetConfig->ExcludeFilters->$facet)
                    ) {
                        $filters = !empty($this->facetConfig->FacetFilters->$facet)
                            ? $this->facetConfig->FacetFilters->$facet->toArray()
                            : [];
                        $excludeFilters
                            = !empty($this->facetConfig->ExcludeFilters->$facet)
                            ? $this->facetConfig->ExcludeFilters->$facet->toArray()
                            : [];

                        $facetList = $this->facetHelper->filterFacets(
                            $facetList,
                            $filters,
                            $excludeFilters
                        );
                    }

                    $response[$facet]['list'] = $facetList;
                } else {
                    $context['facet'] = $facet;
                    $context['cluster'] = $sideFacetSet[$facet] ?? [];
                    $response[$facet]['html'] = $this->renderer->render(
                        'Recommend/SideFacets/facet.phtml',
                        $context
                    );
                }
            }
            return $this->formatResponse(['facets' => $response]);
        }

        // Render full sidefacets
        $html = $this->renderer->render(
            'Recommend/SideFacets.phtml',
            $context
        );
        return $this->formatResponse(compact('html'));
    }
}
