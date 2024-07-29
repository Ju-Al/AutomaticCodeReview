        if (!$this->uriSigner->check($request->getSchemeAndHttpHost().$request->getBaseUrl().$request->getPathInfo().(null !== ($qs = $request->server->get('QUERY_STRING')) ? '?'.$qs : ''))) {
            throw new BadRequestHttpException();
        }

        return new RedirectResponse(base64_decode($request->query->get('redirect'), true));
    }
}
