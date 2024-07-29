 */
class NodeTypeConstraintFactory
{
    /**
     * @Flow\Inject
     * @var Domain\Service\NodeTypeManager
     */
    protected $nodeTypeManager;

    /**
     * @param string $serializedFilters
     * @return NodeTypeConstraints
     */
    public function parseFilterString(string $serializedFilters): NodeTypeConstraints
    {
        $explicitlyAllowedNodeTypeNames = [];
        $explicitlyDisallowedNodeTypeNames = [];

        $onlyNegatedFilters = true;
        $nodeTypeFilterParts = Arrays::trimExplode(',', $serializedFilters);
        foreach ($nodeTypeFilterParts as $nodeTypeFilterPart) {
            if (\mb_strpos($nodeTypeFilterPart, '!') === 0) {
                $negate = true;
                $nodeTypeFilterPart = \mb_substr($nodeTypeFilterPart, 1);
            } else {
                $onlyNegatedFilters = false;
                $negate = false;
            }
            $nodeTypeFilterPartSubTypes = array_merge([$nodeTypeFilterPart], array_map(function (Domain\Model\NodeType $nodeType) {
                return $nodeType->getName();
            }, $this->nodeTypeManager->getSubNodeTypes($nodeTypeFilterPart, true)));

            foreach ($nodeTypeFilterPartSubTypes as $nodeTypeFilterPartSubType) {
                if ($negate) {
                    $explicitlyDisallowedNodeTypeNames[] = $nodeTypeFilterPartSubType;
                } else {
                    $explicitlyAllowedNodeTypeNames[] = $nodeTypeFilterPartSubType;
                }
            }
        }

        // in case there are no filters, we fall back to allowing every node type.
        // Furthermore, if there are only negated filters, we also fall back to allowing every node type (when the blacklist does not match)
        $wildcardAllowed = empty($serializedFilters) || (!empty($serializedFilters) && $onlyNegatedFilters);

        return new NodeTypeConstraints($wildcardAllowed, $explicitlyAllowedNodeTypeNames, $explicitlyDisallowedNodeTypeNames);
    }
}