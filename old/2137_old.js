        name,
        compiler =>
            compiler.hooks.compilation.tap(name, compilation =>
                compilation.hooks.normalModuleLoader.tap(
                    name,
                    (loaderContext, mod) => {
                        if (isRCR(mod) && !loaderIsInstalled(mod)) {
                            const renderers = [];
                            const api = {
                                add(name, importString) {
                                    renderers.push([name, importString]);
                                }
                            };
                            richContentRenderers.call(api);
                            mod.loaders.push({
                                ident: __filename,
                                loader,
                                options: {
                                    renderers
                                }
                            });
                        }
                    }
                )
            )
    );
};
