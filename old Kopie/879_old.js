const assert = require('assert');
    const URL = 'http://localhost:4567/lwc-dom-manual-quill';

    before(() => {
        browser.url(URL);
    });

    it('should not throw an error', () => {
        const button = browser.execute(function () {
            return document.querySelector('integration-lwc-dom-manual-quill').shadowRoot.querySelector('button');
        });
        const errorMessageEl = browser.execute(function () {
            return document.querySelector('integration-lwc-dom-manual-quill').shadowRoot.querySelector('.error');
        });

        button.click();

        browser.waitUntil(() => {
            return errorMessageEl.getText() === 'No error';
        }, 300, 'Should not have thrown an error');
    });
});
