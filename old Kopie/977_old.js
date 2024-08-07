import { Application } from 'spectron';
import { defineSupportCode } from 'cucumber';
import electronPath from 'electron';
import environment from '../../source/common/environment';

const context = {};
const DEFAULT_TIMEOUT = 20000;
let scenariosCount = 0;

const printMainProcessLogs = () => (
  context.app.client.getMainProcessLogs()
    .then((logs) => {
      console.log('========= DAEDALUS LOGS =========');
      logs.forEach((log) => console.log(log));
      console.log('=================================');
      return true;
    })
);

defineSupportCode(({ BeforeAll, Before, After, AfterAll, setDefaultTimeout }) => {
  // The cucumber timeout should be high (and never reached in best case)
  // because the errors thrown by webdriver.io timeouts are more descriptive
  // and helpful than "this step timed out after 5 seconds" messages
  setDefaultTimeout(DEFAULT_TIMEOUT + 1000);

  // Boot up the electron app before all features
  BeforeAll({ timeout: 5 * 60 * 1000 }, async () => {
    const app = new Application({
      path: electronPath,
      args: ['./dist/main/index.js'],
      env: {
        NODE_ENV: environment.TEST,
      },
      waitTimeout: DEFAULT_TIMEOUT
    });
    await app.start();
    await app.client.waitUntilWindowLoaded();
    context.app = app;
  });

  // Make the electron app accessible in each scenario context
  Before({ timeout: DEFAULT_TIMEOUT * 2 }, async function () {
    this.client = context.app.client;
    this.browserWindow = context.app.browserWindow;

    // Set timeouts of various operations:

    // Determines when to interrupt a script that is being evaluated.
    this.client.timeouts('script', DEFAULT_TIMEOUT);
    // Provides the timeout limit used to interrupt navigation of the browsing context.
    this.client.timeouts('pageLoad', DEFAULT_TIMEOUT);
    // Do not set 'implicit' timeout here because of this issue:
    // https://github.com/webdriverio/webdriverio/issues/974

    // Reset backend
    await this.client.executeAsync((done) => {
      const resetBackend = () => {
        if (daedalus.stores.networkStatus.isConnected) {
          daedalus.api.ada.testReset()
            .then(() => daedalus.api.localStorage.reset())
            .then(done)
            .catch((error) => done(error));
        } else {
          setTimeout(resetBackend, 50);
        }
      };
      resetBackend();
    });

    const url = `file://${__dirname}/../../dist/renderer/index.html`;

    // Load fresh root url with test environment for each test case
    await this.client.url(url);

    // Ensure that frontend is synced and ready before test case
    await this.client.executeAsync((done) => {
      const waitUntilSyncedAndReady = () => {
        if (daedalus.stores.networkStatus.isSynced) {
          done();
        } else {
          setTimeout(waitUntilSyncedAndReady, 50);
        }
      };
      waitUntilSyncedAndReady();
    });
  });

  // after the node update acceptance test shuts it down via 'kill-process'
  // eslint-disable-next-line prefer-arrow-callback
  After({ tags: '@preventAppQuit' }, async function () {
    await context.app.restart();
  });

  // eslint-disable-next-line prefer-arrow-callback
  After(async function ({ result }) {
    scenariosCount++;
    if (result.status === 'failed') {
      await printMainProcessLogs();
    }
  });

  // eslint-disable-next-line prefer-arrow-callback
  AfterAll(async function () {
    if (!context.app.running) return;

    if (scenariosCount === 0) {
      await printMainProcessLogs();
    }

    return context.app.stop();
  });
});
