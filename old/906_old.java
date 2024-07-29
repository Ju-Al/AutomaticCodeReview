    {
        return isSauceConnectMetaTagContainedIn(
                bddRunContext.getRunningStory().getRunningScenario().getScenario().getMeta());
    }

    private boolean isSauceConnectMetaTagContainedIn(Meta meta)
    {
        return ControllingMetaTag.isContainedIn(meta, SAUCE_CONNECT_META_TAG);
    }

    public void setSauceConnectEnabled(boolean sauceConnectEnabled)
    {
        this.sauceConnectEnabled = sauceConnectEnabled;
    }
}
