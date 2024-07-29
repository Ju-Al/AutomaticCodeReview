     */
    protected $webPort = 4444;

    /**
     * @var array
     */
    protected $adminViews;

    /**
     * @var WebDriver $webDriver
     */
    protected $webDriver;

    public function setUp()
    {
        parent::setUp();
        if (empty(getenv('DOMAIN'))) {
            die('Must specify DOMAIN environment variable to run this test, like "DOMAIN=localhost/limesurvey" or "DOMAIN=limesurvey.localhost".');
        }

        $capabilities = DesiredCapabilities::phantomjs();
        $this->webDriver = RemoteWebDriver::create("http://localhost:{$this->webPort}/", $capabilities);
        $this->adminViews = require __DIR__."/data/views/adminViews.php";

    }

    /**
     * Tear down fixture.
     */
    public function tearDown()
    {
        // Close Firefox.
        $this->webDriver->quit();
    }

    /**
     * @param array $view
     * @return WebDriver
     */
    public function openView($view){
        $domain = getenv('DOMAIN');
        if (empty($domain)) {
            $domain = '';
        }
        $url = "http://{$domain}/index.php/admin/".$view['route'];
        return $this->webDriver->get($url);
    }

    public function adminLogin($userName,$passWord){
        $this->openView($this->adminViews['login']);
        $userNameField = $this->webDriver->findElement(WebDriverBy::id("user"));
        $userNameField->clear()->sendKeys($userName);
        $passWordField = $this->webDriver->findElement(WebDriverBy::id("password"));
        $passWordField->clear()->sendKeys($passWord);

        $submit = $this->webDriver->findElement(WebDriverBy::name('login_submit'));
        $submit->click();
        return $this->webDriver->wait(10, 1000)->until(
            WebDriverExpectedCondition::visibilityOfElementLocated(WebDriverBy::id('welcome-jumbotron'))
        );
    }

}
