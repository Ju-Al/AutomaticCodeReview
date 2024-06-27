import TabbedView from 'tabbedView';
define(['tabbedView', 'globalize', 'require', 'emby-tabs', 'emby-button', 'emby-scroller'], function (TabbedView, globalize, require) {
    'use strict';
import globalize from 'globalize';
import 'emby-tabs';
import 'emby-button';
import 'emby-scroller';

class HomeView extends TabbedView {
    constructor(view, params) {
        super(view, params);
    }

    function getTabs() {
        return [{
            name: globalize.translate('Home')
        }, {
            name: globalize.translate('Favorites')
        }];
    setTitle() {
        Emby.Page.setTitle(null);
    }

    function getDefaultTabIndex() {
    onPause() {
        super.onPause(this);
        document.querySelector('.skinHeader').classList.remove('noHomeButtonHeader');
    }

    onResume(options) {
        super.onResume(this, options);
        document.querySelector('.skinHeader').classList.add('noHomeButtonHeader');
    }

    getDefaultTabIndex() {
        return 0;
    }

    function getRequirePromise(deps) {
        return new Promise(function (resolve, reject) {
            require(deps, resolve);
        });
    getTabs() {
        return [{
            name: globalize.translate('Home')
        }, {
            name: globalize.translate('Favorites')
        }];
    }

    function getTabController(index) {
    getTabController(index) {
        if (index == null) {
            throw new Error('index cannot be null');
        }

        var depends = [];
        let depends = '';

        switch (index) {
            case 0:
                depends.push('controllers/hometab');
                depends = 'controllers/hometab';
                break;

            case 1:
                depends.push('controllers/favorites');
                depends = 'controllers/favorites';
        }

        var instance = this;
        return getRequirePromise(depends).then(function (controllerFactory) {
            var controller = instance.tabControllers[index];
        const instance = this;
        return import(depends).then(({ default: controllerFactory }) => {
            let controller = instance.tabControllers[index];

            if (!controller) {
                controller = new controllerFactory.default(instance.view.querySelector(".tabContent[data-index='" + index + "']"), instance.params);
                instance.tabControllers[index] = controller;
            }

            return controller;
        });
    }
}

export default HomeView;
