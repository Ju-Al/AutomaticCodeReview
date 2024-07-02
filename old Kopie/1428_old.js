////////////////////////////////////////////////////////////////////////////
//
// Copyright 2016 Realm Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////

'use strict';

const fs = require('fs-extra');
const path = require('path');
const os = require('os');
const child_process = require('child_process');
const fetch = require('node-fetch');
const ini = require('ini').parse;
const decompress = require('decompress');

const MANIFEST_FILENAME = '.manifest.list';

function exec() {
    const args = Array.from(arguments);
    return new Promise((resolve, reject) => {
        function callback(error, stdout, stderr) {
            if (error) {
                reject(error);
            } else {
                resolve(stdout.trim());
            }
        }
        args.push(callback);
        child_process.exec.apply(null, args);
    });
}

function printProgress(input, totalBytes, archive) {
    const ProgressBar = require('progress');
    const StreamCounter = require('stream-counter');

    const message = `Downloading ${archive} [:bar] (:ratek)`;
    const bar = new ProgressBar(message, {
       complete: '=',
       incomplete: ' ',
       width: process.stdout.columns - message.length,
       total: totalBytes / 1024
     });

     input.pipe(new StreamCounter()).on('progress', function() {
         bar.tick((this.bytes / 1024) - bar.curr);
     });
}

function download(serverFolder, archive, destination) {
    const url = `https://static.realm.io/downloads/${serverFolder}/${archive}`;
    return fetch(url).then((response) => {
        if (response.status !== 200) {
            throw new Error(`Error downloading ${url} - received status ${response.status} ${response.statusText}`);
        }

        const lastModified = new Date(response.headers.get('Last-Modified'));
        return fs.exists(destination)
                 .then(exists => {
                     if (!exists) {
                         return saveFile();
                     } else {
                         return fs.stat(destination)
                                  .then(stat => {
                                      if (stat.mtime.getTime() !== lastModified.getTime()) {
                                          return saveFile();
                                      }
                         })
                     }
        });

        function saveFile() {
            if (process.stdout.isTTY) {
                printProgress(response.body, parseInt(response.headers.get('Content-Length')), archive);
            } else {
                console.log(`Downloading ${archive}`);
            }
            return new Promise((resolve) => {
                const file = fs.createWriteStream(destination);
                response.body.pipe(file).once('finish', () => file.close(resolve));
            }).then(() => fs.utimes(destination, lastModified, lastModified));
        }
    });
}

function extract(downloadedArchive, targetFolder, archiveRootFolder) {
    console.log(`Extracting ${path.basename(downloadedArchive)} => ${targetFolder}`);
    const decompressOptions = /tar\.xz$/.test(downloadedArchive) ? { plugins: [ require('decompress-tarxz')() ] } : undefined;
    if (!archiveRootFolder) {
        return decompress(downloadedArchive, targetFolder, decompressOptions);
    } else {
        const tempExtractLocation = path.resolve(os.tmpdir(), path.basename(downloadedArchive, path.extname(downloadedArchive)));
        return decompress(downloadedArchive, tempExtractLocation, decompressOptions)
               .then(() => fs.readdir(path.resolve(tempExtractLocation, archiveRootFolder)))
               .then(items => Promise.all(items.map(item => {
                   const source = path.resolve(tempExtractLocation, archiveRootFolder, item);
                   const target = path.resolve(targetFolder, item);
                   return fs.copy(source, target, { filter: n => path.extname(n) !== '.so' });
               })))
               .then(() => fs.remove(tempExtractLocation))
    }
}

function acquire(desired, target) {
    const corePath = desired.CORE_ARCHIVE && path.resolve(os.tmpdir(), desired.CORE_ARCHIVE);
    const syncPath = desired.SYNC_ARCHIVE && path.resolve(os.tmpdir(), desired.SYNC_ARCHIVE);

    return fs.emptyDir(target)
        .then(() => corePath && download(desired.CORE_SERVER_FOLDER, desired.CORE_ARCHIVE, corePath))
        .then(() => corePath && extract(corePath, target, desired.CORE_ARCHIVE_ROOT))
        .then(() => syncPath && download(desired.SYNC_SERVER_FOLDER, desired.SYNC_ARCHIVE, syncPath))
        .then(() => syncPath && extract(syncPath, target, desired.SYNC_ARCHIVE_ROOT))
        .then(() => writeManifest(target, desired))
}

function getSyncCommitSha(version) {
  return exec(`git ls-remote git@github.com:realm/realm-sync.git --tags "v${version}^{}"`)
         .then(stdout => {
           if (!Boolean(stdout)) {
             return exec(`git ls-remote git@github.com:realm/realm-sync.git --tags "v${version}"`)
           } else {
             return stdout;
           }
         }).then(stdout => /([^\t]+)/.exec(stdout)[0]);
}

function appendCoreRequirements(desired, dependencies, options) {
    desired.CORE_SERVER_FOLDER = `core/v${dependencies.REALM_CORE_VERSION}`;
    let flavor = options.debug ? 'Debug' : 'Release';

    switch (options.platform) {
        case 'mac':
            desired.CORE_SERVER_FOLDER += `/macos/${flavor}`;
            desired.CORE_ARCHIVE = `realm-core-${flavor}-v${dependencies.REALM_CORE_VERSION}-Darwin-devel.tar.gz`;
            return Promise.resolve();
        case 'ios':
            flavor = flavor === 'Debug' ? 'MinSizeDebug' : flavor;
            desired.CORE_SERVER_FOLDER += `/ios/${flavor}`;
            desired.CORE_ARCHIVE = `realm-core-${flavor}-v${dependencies.REALM_CORE_VERSION}-iphoneos.tar.gz`;
            return Promise.resolve();
        case 'win':
            if (!options.arch) throw new Error(`Specifying '--arch' is required for platform 'win'`);
            const arch = options.arch === 'ia32' ? 'Win32' : options.arch;
            desired.CORE_SERVER_FOLDER += `/windows/${arch}/nouwp/${flavor}`;
            desired.CORE_ARCHIVE = `realm-core-${flavor}-v${dependencies.REALM_CORE_VERSION}-Windows-${arch}-devel.tar.gz`;
            return Promise.resolve();
        case 'linux':
            desired.CORE_SERVER_FOLDER = 'core';
            desired.CORE_ARCHIVE = `realm-core-${dependencies.REALM_CORE_VERSION}.tgz`;
            desired.CORE_ARCHIVE_ROOT = `realm-core-${dependencies.REALM_CORE_VERSION}`;
            return Promise.resolve();
        default:
            return Promise.reject(new Error(`Unsupported sync platform '${options.platform}'`));
    }
}

function appendSyncRequirements(desired, dependencies, options) {
    desired.SYNC_SERVER_FOLDER = 'sync';
    let flavor = options.debug ? 'Debug' : 'Release';

    switch (options.platform) {
        case 'mac':
            desired.SYNC_ARCHIVE = `realm-sync-node-cocoa-${dependencies.REALM_SYNC_VERSION}.tar.gz`;
            desired.SYNC_ARCHIVE_ROOT = `realm-sync-node-cocoa-${dependencies.REALM_SYNC_VERSION}`;
            return Promise.resolve();
        case 'ios':
            desired.SYNC_ARCHIVE = `realm-sync-cocoa-${dependencies.REALM_SYNC_VERSION}.tar.xz`;
            desired.SYNC_ARCHIVE_ROOT = `core`;
            return Promise.resolve();
        case 'win':
            const arch = options.arch === 'ia32' ? 'Win32' : options.arch;
            desired.SYNC_ARCHIVE = `realm-sync-${flavor}-v${dependencies.REALM_SYNC_VERSION}-Windows-${arch}-devel.tar.gz`;
            return appendCoreRequirements(desired, dependencies, options)
                .then(() => getSyncCommitSha(dependencies.REALM_SYNC_VERSION))
                .then(sha => desired.SYNC_SERVER_FOLDER += `/sha-version/${sha}`);
        default:
            return Promise.reject(new Error(`Unsupported sync platform '${options.platform}'`));
    }
}

function calculateRequirements(options, dependencies) {
    const desired = {};
    return (options.sync ? appendSyncRequirements : appendCoreRequirements)(desired, dependencies, options)
        .then(() => desired);
}

function writeManifest(target, manifest) {
    return fs.writeFile(path.resolve(target, MANIFEST_FILENAME), ini.encode(manifest));
}

function readManifest(target) {
    try {
        return ini.parse(fs.readFileSync(path.resolve(target, MANIFEST_FILENAME), 'utf8'));
    } catch (e) {
        return false;
    }
}

function shouldSkipAcquire(target, desiredManifest, force) {
    const existingManifest = readManifest(target);

    if (!existingManifest) {
        console.log('No manifest at the target, proceeding.');
    } else if (!Object.keys(desiredManifest).every(key => existingManifest[key] === desiredManifest[key])) {
        console.log('Target has is non-empty but has a different manifest, overwriting.');
    } else if (force) {
        console.log('Target has a matching manifest but download forced.');
    } else {
        console.log('Matching manifest already exists at target - nothing to do (use --force to override)');
        return true;
    }
    return false;
}

const optionDefinitions = [
    { name: 'platform', type: String, defaultOption: true },
    { name: 'arch', type: String },
    { name: 'sync', type: Boolean },
    { name: 'debug', type: Boolean },
    { name: 'force', type: Boolean },
];
const options = require('command-line-args')(optionDefinitions);
if (options.platform === '..\\win') {
    options.platform = 'win'; // handle gyp idiocy
}

const vendorDir = path.resolve(__dirname, '../vendor');

let realmDir = path.resolve(vendorDir, `realm-${options.platform}`);
if (options.arch) {
    realmDir += `-${options.arch}`;
}
if (options.debug) {
    realmDir += '-dbg'
}

const dependencies = ini.parse(fs.readFileSync(path.resolve(__dirname, '../dependencies.list'), 'utf8'));

calculateRequirements(options, dependencies)
    .then(manifest => {
        console.log('Desired manifest', manifest);
        if (!shouldSkipAcquire(realmDir, manifest, options.force)) {
            return acquire(manifest, realmDir)
        }
    })
    .then(() => console.log('Success'))
    .catch(error => {
        console.error(error);
        process.exit(1);
    });
