import dialogHelper from 'dialogHelper';
define(['dialogHelper', 'dom', 'layoutManager', 'connectionManager', 'globalize', 'loading', 'browser', 'focusManager', 'scrollHelper', 'material-icons', 'formDialogStyle', 'emby-button', 'emby-itemscontainer', 'cardStyle'], function (dialogHelper, dom, layoutManager, connectionManager, globalize, loading, browser, focusManager, scrollHelper) {
    'use strict';

    browser = browser.default || browser;
    loading = loading.default || loading;
    focusManager = focusManager.default || focusManager;
    scrollHelper = scrollHelper.default || scrollHelper;

    var enableFocusTransform = !browser.slow && !browser.edge;

    function getEditorHtml() {
        var html = '';
        html += '<div class="formDialogContent scrollY">';
        html += '<div class="dialogContentInner dialog-content-centered">';
        html += '<div class="loadingContent hide">';
        html += '<h1>' + globalize.translate('DetectingDevices') + '...</h1>';
        html += '<p>' + globalize.translate('MessagePleaseWait') + '</p>';
        html += '</div>';
        html += '<h1 style="margin-bottom:.25em;" class="devicesHeader hide">' + globalize.translate('HeaderNewDevices') + '</h1>';
        html += '<div is="emby-itemscontainer" class="results vertical-wrap">';
        html += '</div>';
        html += '</div>';
        return html += '</div>';
    }

    function getDeviceHtml(device) {
        var padderClass;
        var html = '';
        var cssClass = 'card scalableCard';
        var cardBoxCssClass = 'cardBox visualCardBox';
        cssClass += ' backdropCard backdropCard-scalable';
        padderClass = 'cardPadder-backdrop';

        // TODO move card creation code to Card component

        if (layoutManager.tv) {
            cssClass += ' show-focus';

            if (enableFocusTransform) {
                cssClass += ' show-animation';
import layoutManager from 'layoutManager';
import connectionManager from 'connectionManager';
import globalize from 'globalize';
import loading from 'loading';
import browser from 'browser';
import focusManager from 'focusManager';
import scrollHelper from 'scrollHelper';
import 'material-icons';
import 'formDialogStyle';
import 'emby-button';
import 'emby-itemscontainer';
import 'cardStyle';

const enableFocusTransform = !browser.slow && !browser.edge;

function getEditorHtml() {
    let html = '';
    html += '<div class="formDialogContent scrollY">';
    html += '<div class="dialogContentInner dialog-content-centered">';
    html += '<div class="loadingContent hide">';
    html += '<h1>' + globalize.translate('DetectingDevices') + '...</h1>';
    html += '<p>' + globalize.translate('MessagePleaseWait') + '</p>';
    html += '</div>';
    html += '<h1 style="margin-bottom:.25em;" class="devicesHeader hide">' + globalize.translate('HeaderNewDevices') + '</h1>';
    html += '<div is="emby-itemscontainer" class="results vertical-wrap">';
    html += '</div>';
    html += '</div>';
    return html += '</div>';
}

function getDeviceHtml(device) {
    let html = '';
    let cssClass = 'card scalableCard';
    const cardBoxCssClass = 'cardBox visualCardBox';
    cssClass += ' backdropCard backdropCard-scalable';
    const padderClass = 'cardPadder-backdrop';

    // TODO move card creation code to Card component

    if (layoutManager.tv) {
        cssClass += ' show-focus';

        if (enableFocusTransform) {
            cssClass += ' show-animation';
        }
    }

    html += '<button type="button" class="' + cssClass + '" data-id="' + device.DeviceId + '" style="min-width:33.3333%;">';
    html += '<div class="' + cardBoxCssClass + '">';
    html += '<div class="cardScalable visualCardBox-cardScalable">';
    html += '<div class="' + padderClass + '"></div>';
    html += '<div class="cardContent searchImage">';
    html += '<div class="cardImageContainer coveredImage"><span class="cardImageIcon material-icons dvr"></span></div>';
    html += '</div>';
    html += '</div>';
    html += '<div class="cardFooter visualCardBox-cardFooter">';
    html += '<div class="cardText cardTextCentered">' + getTunerName(device.Type) + '</div>';
    html += '<div class="cardText cardTextCentered cardText-secondary">' + device.FriendlyName + '</div>';
    html += '<div class="cardText cardText-secondary cardTextCentered">';
    html += device.Url || '&nbsp;';
    html += '</div>';
    html += '</div>';
    html += '</div>';
    return html += '</button>';
}

function getTunerName(providerId) {
    switch (providerId = providerId.toLowerCase()) {
        case 'm3u':
            return 'M3U';

        case 'hdhomerun':
            return 'HDHomerun';

        case 'hauppauge':
            return 'Hauppauge';

        case 'satip':
            return 'DVB';

        default:
            return 'Unknown';
    }
}

function renderDevices(view, devices) {
    let i;
    let length;
    let html = '';

    for (i = 0, length = devices.length; i < length; i++) {
        html += getDeviceHtml(devices[i]);
    }

    if (devices.length) {
        view.querySelector('.devicesHeader').classList.remove('hide');
    } else {
        html = '<p><br/>' + globalize.translate('NoNewDevicesFound') + '</p>';
        view.querySelector('.devicesHeader').classList.add('hide');
    }

    const elem = view.querySelector('.results');
    elem.innerHTML = html;

    if (layoutManager.tv) {
        focusManager.autoFocus(elem);
    }
}

function discoverDevices(view, apiClient) {
    loading.show();
    view.querySelector('.loadingContent').classList.remove('hide');
    return ApiClient.getJSON(ApiClient.getUrl('LiveTv/Tuners/Discvover', {
        NewDevicesOnly: true
    })).then(function (devices) {
        currentDevices = devices;
        renderDevices(view, devices);
        view.querySelector('.loadingContent').classList.add('hide');
        loading.hide();
    });
}

function tunerPicker() {
    this.show = function (options) {
        const dialogOptions = {
            removeOnClose: true,
            scrollY: false
        };

        if (layoutManager.tv) {
            dialogOptions.size = 'fullscreen';
        } else {
            dialogOptions.size = 'small';
        }

        const dlg = dialogHelper.createDialog(dialogOptions);
        dlg.classList.add('formDialog');
        let html = '';
        html += '<div class="formDialogHeader">';
        html += '<button is="paper-icon-button-light" class="btnCancel autoSize" tabindex="-1"><span class="material-icons arrow_back"></span></button>';
        html += '<h3 class="formDialogHeaderTitle">';
        html += globalize.translate('HeaderLiveTvTunerSetup');
        html += '</h3>';
        html += '</div>';
        html += getEditorHtml();
        dlg.innerHTML = html;
        dlg.querySelector('.btnCancel').addEventListener('click', function () {
            dialogHelper.close(dlg);
        });
        let deviceResult;
        dlg.querySelector('.results').addEventListener('click', function (e) {
            const tunerCard = dom.parentWithClass(e.target, 'card');

            if (tunerCard) {
                const deviceId = tunerCard.getAttribute('data-id');
                deviceResult = currentDevices.filter(function (d) {
                    return d.DeviceId === deviceId;
                })[0];
                dialogHelper.close(dlg);
            }
        });

        if (layoutManager.tv) {
            scrollHelper.centerFocus.on(dlg.querySelector('.formDialogContent'), false);
        }

        const apiClient = connectionManager.getApiClient(options.serverId);
        discoverDevices(dlg, apiClient);

        if (layoutManager.tv) {
            scrollHelper.centerFocus.off(dlg.querySelector('.formDialogContent'), false);
        }

        return dialogHelper.open(dlg).then(function () {
            if (deviceResult) {
                return Promise.resolve(deviceResult);
            }

            return Promise.reject();
        });
    };
}

let currentDevices = [];

export default tunerPicker;
