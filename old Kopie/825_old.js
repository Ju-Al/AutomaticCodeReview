const path = require('path').join(__dirname, '..', 'package.json');

const UNNECESSARY_KEYS = [
	'babel',
	'greenkeeper',
	'bundlesize',
	'devDependencies',
	'eslintConfig'
];

UNNECESSARY_KEYS.forEach(key => delete pkg[key]);

pkg.scripts = {
	donate: pkg.scripts.donate,
	postinstall: pkg.scripts.postinstall,
	postpublish: 'mv -f .package.json package.json'
};

require('fs').writeFileSync(path, JSON.stringify(pkg, null, 2));
