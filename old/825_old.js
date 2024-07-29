
pkg.scripts = {
	donate: pkg.scripts.donate,
	postinstall: pkg.scripts.postinstall,
	postpublish: 'mv -f .package.json package.json'
};

require('fs').writeFileSync(path, JSON.stringify(pkg, null, 2));
