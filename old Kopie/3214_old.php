<?php
declare(strict_types=1);

/*
 * This file is part of Contao.
 *
 * (c) Leo Feyer
 *
 * @license LGPL-3.0-or-later
 */

namespace Contao\CoreBundle\Csrf;

use Symfony\Component\HttpFoundation\RequestStack;
use Symfony\Component\Security\Csrf\CsrfToken;
use Symfony\Component\Security\Csrf\CsrfTokenManager;
use Symfony\Component\Security\Csrf\TokenGenerator\TokenGeneratorInterface;
use Symfony\Component\Security\Csrf\TokenStorage\TokenStorageInterface;

class ContaoCsrfTokenManager extends CsrfTokenManager
{
    /**
     * @var RequestStack
     */
    private $requestStack;

    /**
     * @var string
     */
    private $csrfCookiePrefix;

    public function __construct(RequestStack $requestStack, string $csrfCookiePrefix, TokenGeneratorInterface $generator = null, TokenStorageInterface $storage = null, $namespace = null)
    {
        $this->requestStack = $requestStack;
        $this->csrfCookiePrefix = $csrfCookiePrefix;

        parent::__construct($generator, $storage, $namespace);
    }

    public function isTokenValid(CsrfToken $token): bool
    {
        // Skip the CSRF token validation if the request has no cookies, no
        // authenticated user and the session has not been started
        if (
            ($request = $this->requestStack->getMasterRequest())
            && 'POST' === $request->getRealMethod()
            && !$request->getUserInfo()
            && (
                0 === $request->cookies->count()
                || [$this->csrfCookiePrefix.$token->getId()] === $request->cookies->keys()
            )
            && !($request->hasSession() && $request->getSession()->isStarted())
        ) {
            return true;
        }

        return parent::isTokenValid($token);
    }
}
