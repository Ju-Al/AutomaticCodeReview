<?php
        $label = Utility::evaluateEelExpression($this->getExpression(), $this->eelEvaluator, array('node' => $node), $this->defaultContextConfiguration);
namespace Neos\ContentRepository\Domain\Model;

/*
 * This file is part of the Neos.ContentRepository package.
 *
 * (c) Contributors of the Neos Project - www.neos.io
 *
 * This package is Open Source Software. For the full copyright and license
 * information, please view the LICENSE file which was distributed with this
 * source code.
 */

use Neos\Eel\EelEvaluatorInterface;
use Neos\Eel\Exception;
use Neos\Eel\Utility;
use Neos\Flow\Annotations as Flow;
use Neos\Flow\ObjectManagement\DependencyInjection\DependencyProxy;

/**
 * The expression based node label generator that is used as default if a label expression is configured.
 *
 */
class ExpressionBasedNodeLabelGenerator implements NodeLabelGeneratorInterface
{
    /**
     * @Flow\Inject
     * @var EelEvaluatorInterface
     */
    protected $eelEvaluator;

    /**
     * @Flow\InjectConfiguration("labelGenerator.eel.defaultContext")
     * @var array
     */
    protected $defaultContextConfiguration;

    /**
     * @var string
     */
    protected $expression = '${(node.nodeType.label ? node.nodeType.label : node.nodeType.name) + \' (\' + node.name + \')\'}';

    /**
     * @return string
     */
    public function getExpression()
    {
        return $this->expression;
    }

    /**
     * @param string $expression
     */
    public function setExpression($expression)
    {
        $this->expression = $expression;
    }

    /**
     * @return void
     */
    public function initializeObject()
    {
        if ($this->eelEvaluator instanceof DependencyProxy) {
            $this->eelEvaluator->_activateDependency();
        }
    }

    /**
     * Render a node label
     *
     * @param NodeInterface $node
     * @return string
     * @throws Exception
     */
    public function getLabel(NodeInterface $node)
    {
        $label = null;
        if ($node->getNodeType()->isOfType('TYPO3.Neos:ContentCollection') && $node->isAutoCreated()) {
            $parentNode = $node->getParent();
            $property = 'childNodes.' . $node->getName() . '.label';
            if ($parentNode->getNodeType()->hasConfiguration($property)) {
                $labelConfiguration = $parentNode->getNodeType()->getConfiguration($property);
                try {
                    $label = Utility::evaluateEelExpression($labelConfiguration, $this->eelEvaluator, [
                        'node' => $node,
                        'parentNode' => $parentNode
                    ], $this->defaultContextConfiguration);
                } catch (Exception $exception) {
                    if ($exception->getCode() !== 1410441849) {
                        throw $exception;
                    }
                    $label = (string)$labelConfiguration;
                }
            }
        }
        if ($label === null) {
            $label = Utility::evaluateEelExpression($this->getExpression(), $this->eelEvaluator, ['node' => $node], $this->defaultContextConfiguration);
        }
        return $label;
    }
}
