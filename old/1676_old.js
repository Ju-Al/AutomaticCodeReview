define(['tabbedView', 'globalize', 'require', 'emby-tabs', 'emby-button', 'emby-scroller'], function (TabbedView, globalize, require) {
    'use strict';

    function getTabs() {
        return [{
            name: globalize.translate('Home')
        }, {
            name: globalize.translate('Favorites')
        }];
    }

    function getDefaultTabIndex() {
        return 0;
    }

    function getRequirePromise(deps) {
        return new Promise(function (resolve, reject) {
            require(deps, resolve);
        });
    }

    function getTabController(index) {
        if (index == null) {
            throw new Error('index cannot be null');
        }

        var depends = [];

        switch (index) {
            case 0:
                depends.push('controllers/hometab');
                break;

            case 1:
                depends.push('controllers/favorites');
        }

        var instance = this;
        return getRequirePromise(depends).then(function (controllerFactory) {
            var controller = instance.tabControllers[index];

            if (!controller) {
                controller = new controllerFactory.default(instance.view.querySelector(".tabContent[data-index='" + index + "']"), instance.params);
                instance.tabControllers[index] = controller;
            }

            return controller;
        });
    }
}

export default HomeView;
