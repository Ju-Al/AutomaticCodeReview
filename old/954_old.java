        assertNotEquals(options1, createDefaultOptions());
    }

    @Test
    public void testNotEqualsBasicAuthUser()
    {
        SauceConnectOptions options1 = createDefaultOptions();
        options1.setBasicAuthUser(null);
        assertNotEquals(options1, createDefaultOptions());
    }

    @Test
    public void testNotEqualsHost()
    {
        SauceConnectOptions options1 = createDefaultOptions();
        options1.setHost(null);
        assertNotEquals(options1, createDefaultOptions());
    }

    @Test
    public void testNotEqualsPort()
    {
        SauceConnectOptions options1 = createDefaultOptions();
        options1.setPort(8888);
        assertNotEquals(options1, createDefaultOptions());
    }

    @Test
    public void testNotEqualsNoSslBumpDomains()
    {
        SauceConnectOptions options1 = createDefaultOptions();
        options1.setNoSslBumpDomains(null);
        assertNotEquals(options1, createDefaultOptions());
    }

    @Test
    public void testNotEqualsSkipProxyHostsPattern()
    {
        SauceConnectOptions options1 = createDefaultOptions();
        options1.setSkipProxyHostsPattern(null);
        assertNotEquals(options1, createDefaultOptions());
    }

    @Test
    public void testNotEqualsRestUrl()
    {
        SauceConnectOptions options1 = createDefaultOptions();
        options1.setRestUrl(null);
        assertNotEquals(options1, createDefaultOptions());
    }

    private static Path mockPidPath() throws IOException
    {
        Path pidPath = mock(Path.class);
        PowerMockito.mockStatic(Files.class);
        when(Files.createTempFile(PID_FILE_NAME, PID_EXTENSION)).thenReturn(pidPath);
        return pidPath;
    }

    private SauceConnectOptions createDefaultOptions()
    {
        SauceConnectOptions options = new SauceConnectOptions();
        options.setProxy(PROXY);
        options.setHost(HOST);
        options.setBasicAuthUser(USER);
        options.setNoSslBumpDomains("--no-ssl-bump-domains all");
        options.setSkipProxyHostsPattern("vividus\\.dev");
        options.setRestUrl(SAUCE_LABS_REST_URL);
        options.setPort(PORT);
        return options;
    }
}
