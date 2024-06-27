//
 * List blob references in container.
batchFileUtils.listContainerContents = function (prefix, container, blobService, _) {
    storageOptions.prefix = prefix;
    var contents = storageUtil.performStorageOperation(operation, _, container, storageOptions);
    return contents;
// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//

var __ = require('underscore');
var fs = require('fs');
var glob = require('glob');
var url = require('url');
var crypto = require('crypto');
var storage = require('azure-storage');
var pathUtil = require('path');
var util = require('util');
var profile = require('../../util/profile');
var batchUtil = require('./batch.util');
var cliUtils = require('../../util/utils');
var storageUtil = require('../../util/storage.util');

var startProgress = batchUtil.startProgress;
var endProgress = batchUtil.endProgress;
var $ = cliUtils.getLocaleString;
var cli = null;
var logger = null;
var resolvedStorageClient;
var resourceFileCache = {};

var GROUP_PREFIX = 'fgrp-';
var MAX_GROUP_LENGTH = 63 - GROUP_PREFIX.length;
var SAS_EXPIRY = 7 * 24 * 3600 * 1000; // 7 days
var SAS_PERMISSIONS = 'r'; // Read permissions 
var STRIP_PATH = /^[\/\\]+|[\/\\]+$/g; // Remove any leading or trailing '/' or '\\'
var ROUND_DATE = 2 * 60 * 1000; // Round to nearest 2 minutes


var batchFileUtils = {};
/**
* Init cli module
*/
batchFileUtils.init = function (azureCli) {
  cli = azureCli;
  storageUtil.init(cli);
  logger = cli.output;
  resolvedStorageClient = null;
  resourceFileCache = {};
};

/**
 * Return all blob refeferences in a container cache that meet a
 * prefix requirement.
 * @param {string} prefix         optional blob name prefix as filter
 * @param {string} container      container name
 * @param {callback} _            callback function
 */
batchFileUtils.filterResourceCache = function (container, prefix, _) {
  var filtered = [];
  resourceFileCache[container].forEach_(_, 1, function(_, blob) {
    if (__.isUndefined(prefix)) {
      filtered.push(blob);
    } else if (blob.filePath.startsWith(prefix)) {
      filtered.push(blob);
    }
  });
  return filtered;
};

/**
 * List blob references in container.
 * @param {string} source         source parameters, this object should be same as new ResourceFile object (elements include fileGroup/container, prefix)
 * @param {string} container      container name
 * @param {object} blobService    storage blob service client
 * @param {callback} _            callback function
 */
batchFileUtils.listContainerContents = function (source, container, blobService, _) {
  if (!resourceFileCache[container]) {
    resourceFileCache[container] = [];
    var operation = storageUtil.getStorageOperation(blobService, storageUtil.OperationType.Blob, 'listAllBlobs');
    var storageOptions = storageUtil.getStorageOperationDefaultOption();
    var blobs = storageUtil.performStorageOperation(operation, _, container, storageOptions);
    blobs.forEach_(_, 1, function(_, blob) {
      var blobSas =  source.fileGroup ? batchFileUtils.generateSasToken(blob, container, blobService) : batchFileUtils.constructSasUrl(blob, url.parse(source.containerUrl));
      var fileName = pathUtil.basename(blob.name);
      var fileNameOnly = fileName.substr(0, fileName.lastIndexOf('.'));
      resourceFileCache[container].push({  url : blobSas, filePath : blob.name, fileName : fileName, fileNameWithoutExtension : fileNameOnly });
    });
  }
  return batchFileUtils.filterResourceCache(container, source.prefix, _);
};

/**
 * Generate a blob URL with SAS token.
 * @param {object} blob           blob reference
 * @param {string} container      container name
 * @param {object} blobService    storage blob service client
 */
batchFileUtils.generateSasToken = function (blob, container, blobService) {
  var start = new Date();
  var expiry = new Date(start.getTime() + SAS_EXPIRY);
  var sharedAccessPolicy = storageUtil.getSharedAccessPolicy('blob', SAS_PERMISSIONS, null, null, start, expiry, null, null);
  var sas = blobService.generateSharedAccessSignature(container, blob.name, sharedAccessPolicy);
  return blobService.getUrl(container, blob.name, sas);
};

/**
 * Make up blob URL with container URL
 * @param {object} blob           blob reference
 * @param {object} uri            container uri
 */
batchFileUtils.constructSasUrl = function (blob, uri) {
  var newuri = __.clone(uri);
  newuri.pathname += '/' + blob.name;
  return url.format(newuri);
};

/**
 * List blob references in container.
 * @param {object} source     source parameters, this object should be same as new ResourceFile object (elements include fileGroup/container, prefix)
 */
batchFileUtils.getContainerList = function(source, _) {
  var storageClient;
  var container;
  var isAutoStorage = false;

  if (source.fileGroup) {
    //Input data stored in auto-storage
    storageClient = batchFileUtils.resolveStorageAccount({}, _);
    container = batchFileUtils.getContainerName(source.fileGroup);
    isAutoStorage = true;
  } else if (source.containerUrl) {
    var uri = url.parse(source.containerUrl);
    if (!__.isString(uri.query)) {
      throw new Error('Invalid container url.');
    }
    var connection = {
      accountName : uri.hostname.split('.')[0],
      sas : uri.query
    };
    storageClient = storageUtil.getServiceClient(storageUtil.getBlobService, connection);
    storageClient.listAllBlobs = function(container, options, callback) {
      storageUtil.listWithContinuation(storageClient.listBlobsSegmentedWithPrefix, 
      storageClient, storageUtil.ListContinuationTokenArgIndex.Blob, 
      container, options.prefix, null, options, callback);
    };
    container = uri.pathname.split('/')[1];
  } else {
    throw new Error('Unknown source.');
  }

  return batchFileUtils.listContainerContents(source, container, storageClient, _);
};

/**
 * Convert a list of blobs to a list of ResourceFiles
 * @param {object[]} blobs                  list of blob references to convert
 * @param {object} resourceProperties       resource parameters
 * @param {string} container                container name
 * @param {object} blobService              storage blob service client
 * @param {callback} _                      callback function
 */
batchFileUtils.convertBlobsToResourceFiles = function (blobs, resourceProperties, container, blobService, _) {
  var resourceFiles = [];
  if (__.isUndefined(blobs) || blobs.length === 0) {
    throw new Error(util.format('No input data found with reference %s', resourceProperties.source.prefix));
  }
  if (blobs.length === 1 && blobs[0].filePath === resourceProperties.source.prefix) {
    //Single file reference: filePath should be treated as file path
    var filePath = resourceProperties.filePath || blobs[0].filePath;
    resourceFiles.push({
      'blobSource': blobs[0].url,
      'filePath': filePath
    });
  } else {
    //Multiple file reference: filePath should be treated as a directory
    var baseFilePath = '';
    if (resourceProperties.filePath) {
      baseFilePath = resourceProperties.filePath.replace(STRIP_PATH, '') + '/'; 
    }
    blobs.forEach_(_, 1, function(_, blob) {
      var filePath = baseFilePath + blob.filePath;
      resourceFiles.push({
      'blobSource': blob.url,
      'filePath': filePath
      });
    });
  }
  //Add filemode to every resourceFile
  if (resourceProperties.fileMode) {
    resourceFiles.forEach(function(file) {
      file.fileMode = resourceProperties.fileMode;
    });
  }
  return resourceFiles;
};

/**
 * Convert new resourceFile reference to server-supported reference
 * @param {object} resourceFile     resource parameters
 * @param {object} options          command line options
 * @param {callback} _              callback function
 */
batchFileUtils.resolveResourceFile = function (resourceFile, options, _) {
  if (resourceFile.blobSource) {
    //Support original resourceFile reference
    if (!resourceFile.filePath) {
      throw new Error('Malformed ResourceFile: \'blobSource\' must also have \'filePath\' attribute');
    }
    return [resourceFile];
  }
  if (!resourceFile.source) {
    throw new Error('Malformed ResourceFile: Must have either \'source\' or \'blobSource\'');
  }
  var storageClient = batchFileUtils.resolveStorageAccount(options, _);
  var container;
  var blobs;
  if (resourceFile.source.fileGroup) {
    //Input data stored in auto-storage
    container = batchFileUtils.getContainerName(resourceFile.source.fileGroup);
    blobs = batchFileUtils.listContainerContents(resourceFile.source, container, storageClient, _);
    return batchFileUtils.convertBlobsToResourceFiles(blobs, resourceFile, container, storageClient, _);

  } else if (resourceFile.source.containerUrl) {
    //Input data storage in arbitrary container
    var uri = url.parse(resourceFile.source.containerUrl);
    container = uri.pathname.split('/')[1];
    blobs = batchFileUtils.listContainerContents(resourceFile.source, container, storageClient, _);
    return batchFileUtils.convertBlobsToResourceFiles(blobs, resourceFile, container, storageClient, _);

  } else if (resourceFile.source.url) {
    //TODO: Input data from an arbitrary HTTP GET source
    throw new Error('Not implemented');

  } else {
    throw new Error('Malformed ResourceFile');
  }
};

/**
 * Resolve Auto-Storage account from supplied Batch Account
 * @param {object} options      command line options
 * @param {callback} _          callback function
 */
batchFileUtils.resolveStorageAccount = function (options, _) {
  if (resolvedStorageClient) {
    return resolvedStorageClient;
  }
  var resourceGroup = options.resourceGroup;
  var subscriptionOrName = options.subscription;
  var account = options.accountName;
  var connection = {};
  var batchAccount;
  var batchService = batchUtil.createBatchManagementClient(subscriptionOrName, batchUtil.PREVIEW_USER_AGENT_SUFFIX);

  if (!account) {
    //Check for configured Batch Account env vars
    var batchEndpoint = process.env[batchUtil.ENV_SDK_ACCOUNT_ENDPOINT];
    var batchName = process.env[batchUtil.ENV_SDK_ACCOUNT_NAME];
    if (!batchEndpoint || !batchName) {
      throw new Error($('No Storage account or Batch account specified'));
    }
    if (resourceGroup) {
      //If a resource group was supplied, we can use that to query the Batch Account
      batchAccount = batchService.batchAccountOperations.get(resourceGroup, batchName, _);
    } else {
      //Otherwise we need to parse the URL for a region in order to identify
      //the Batch account in the subscription
      //Example URL: https://batchaccount.westus.batch.azure.com
      //TODO: clean this logic up a bit...
      var region = url.parse(batchEndpoint).hostname.split('.')[1];
      var batchAccounts = batchService.batchAccountOperations.list(_).filter_(_, function(_, x){return (x.name === batchName && x.location === region);});
      if (batchAccounts.length != 1) {
        throw new Error(util.format($('No Batch account called %s in region %s found in subscription %s'), batchName, region, subscriptionOrName));
      }
      batchAccount = batchAccounts[0];
    }
  } else if (resourceGroup) {
    //Get account details as specified on cmd line
    batchAccount = batchService.batchAccountOperations.get(resourceGroup, account, _);
  } else {
    throw new Error($('No Storage account or Batch account specified'));
  }
  if (!batchAccount.autoStorage) {
    throw new Error(util.format($('Specificed Batch account %s has no linked storage account'), account));
  }
  try {
    //Parse auto-storage resource ID in order to query for credentials
    //Example resource ID: /subscriptions/xxx-x-x-x-xxx/resourceGroups/my_resources/providers/Microsoft.Storage/storageAccounts/my-storage
    //TODO: Refine parsing
    subscriptionOrName = profile.current.getSubscription(batchAccount.autoStorage.storageAccountId.split('/')[2]);
    resourceGroup = batchAccount.autoStorage.storageAccountId.split('/')[4];
    var provider = batchAccount.autoStorage.storageAccountId.split('/')[6];
    account = batchAccount.autoStorage.storageAccountId.split('/')[8];
    var storageAccessKey;
    var storageService;

    //Get storage credentials and create client
    if(provider == 'Microsoft.Storage') {
      storageService = cliUtils.createStorageResourceProviderClient(subscriptionOrName);
      var keys = storageService.storageAccounts.listKeys(resourceGroup, account, _);
      storageAccessKey = keys['keys'][0]['value'];
    } else if(provider == 'Microsoft.ClassicStorage') {
      storageService = cliUtils.createStorageClient(subscriptionOrName);
      var result = storageService.storageAccounts.getKeys(account, _);
      storageAccessKey = result.primaryKey;
    } else {
      throw new Error(util.format($('Unknown storage account provider %s'), provider));
    }

    connection = { accountName: account, accountKey: storageAccessKey };
    var storageClient = storageUtil.getServiceClient(storageUtil.getBlobService, connection);
    storageClient.listAllBlobs = function(container, options, callback) {
      storageUtil.listWithContinuation(storageClient.listBlobsSegmentedWithPrefix, storageClient, storageUtil.ListContinuationTokenArgIndex.Blob, container, options.prefix, null, options, callback);
    };
    resolvedStorageClient = storageClient;
    return storageClient;

  } catch (e) {
    if (e.code === 'ResourceGroupNotFound') {
      throw new Error(util.format($('Specificed resource group %s not found'), resourceGroup));
    }
    if (e.code === 'ResourceNotFound') {
      throw new Error(util.format($('Specificed account %s not found in resource group %s'), account, resourceGroup));
    }
    throw(e);
  }
};

/**
 * Generate list of files to upload and the relative directory.
 * @param {string} localPath      local path/directory/pattern
 * @param {callback} _            callback function
 */
batchFileUtils.resolveFilePaths = function (localPath, _) {
  var resolved = {'localPath': localPath.replace(STRIP_PATH, ''), 'files': []};
  if(localPath.indexOf('*') > -1) {
    //Supplied path is a pattern - relative directory will be the
    //path up to the first wildcard
    var refDir = localPath.split('*')[0];
    resolved.localPath = refDir.match(STRIP_PATH) ? refDir.replace(STRIP_PATH, '') : pathUtil.dirname(refDir);
    resolved.files = glob(localPath, _);
    return resolved;
  }

  var fsStatus = fs.stat(localPath, _);
  if (fsStatus.isDirectory()) {
    //Supplied path is a directory
    var dirFiles =  fs.readdir(localPath, _);
    resolved.files = dirFiles.map(function (x) {return pathUtil.join(localPath, x); });

  } else {
    //Supplied path is a file
    resolved.localPath = pathUtil.dirname(localPath);
    resolved.files.push(localPath);
  }
  return resolved;
};

/**
 * Generate valid container name from file group name.
 * @param {string} fileGroup    specified file group name
 */
batchFileUtils.generateContainerName = function (fileGroup) {
  fileGroup = fileGroup.toLowerCase();
  var cleanGroup;
  //Check for any chars that aren't 'a-z', '0-9' or '-'
  var validChars = /^[a-z0-9][-a-z0-9]*$/;
  //Replace any underscores or double-hyphens with single hyphen
  var underscoresAndHyphens = /[_-]+/g;
  //Remove a trailing hyphen if present
  var trailingHyphen = /-$/;

  cleanGroup = fileGroup.replace(underscoresAndHyphens, '-');
  cleanGroup = cleanGroup.replace(trailingHyphen, '');

  if (!cleanGroup.match(validChars)) {
    throw new Error(util.format('File group name \'%s\' contains illegal characters. File group names only support alphanumeric characters, underscores and hyphens.', fileGroup));
  }
  
  if (cleanGroup === fileGroup && fileGroup.length <= MAX_GROUP_LENGTH) {
    //If specified group name is clean, no need to add hash
    return fileGroup;
  } else {
    //If we had to transform the group name, add hash of original name
    var hash = crypto.createHash('sha1').update(fileGroup).digest('hex');
    var newGroup = cleanGroup + '-' + hash;
    if (newGroup.length > MAX_GROUP_LENGTH) {
     return cleanGroup.slice(0, 15) + '-' + hash;
    }
    return newGroup;
  }
};

/**
 * Get valid container name from file group name with prefix.
 * @param {string} fileGroup    specified file group name
 */
batchFileUtils.getContainerName = function (fileGroup) {
  return GROUP_PREFIX + batchFileUtils.generateContainerName(fileGroup);
};

/**
 * Upload the specified file to the specified container
 * @param {string} sourcefile     full path to local file to be uploaded
 * @param {string} destination    fileGroup name for container
 * @param {string} filename       relative path of local file to be used as blob name
 * @param {object} options        command line options
 * @param {object} blobService    storage blob service client
 * @param {callback} _            callback function
 */
batchFileUtils.uploadBlob = function (sourcefile, destination, filename, options, blobService, _) {
  var fsStatus = fs.stat(sourcefile, _);
  if (fsStatus.isDirectory()) {
    return;
  }
  if (options.flatten) {
    //Flatten local directory structure
    filename = pathUtil.basename(filename);
  }
  
  //Create upload container with sanitized file group name
  var containerName = batchFileUtils.getContainerName(destination);
  operation = storageUtil.getStorageOperation(blobService, storageUtil.OperationType.Blob, 'createContainerIfNotExists');
  storageUtil.performStorageOperation(operation, _, containerName);

  var blobName = filename;
  if (options.path) {
    //Add any specified virtual directories
    var blobPrefix = options.path.replace(STRIP_PATH, '');
    blobName = blobPrefix + '/' + filename.replace(STRIP_PATH, '');
  }
  var storageOptions = storageUtil.getStorageOperationDefaultOption();
  storageOptions.parallelOperationThreadCount = storageUtil.threadsInOperation;

  //We want to validate the file as we upload, and only complete the operation
  //if all the data transfers successfully
  storageOptions.storeBlobContentMD5 = true;
  storageOptions.useTransactionalMD5 = true;

  //We store the lastmodified timestamp in order to prevent overwriting with
  //out-dated or duplicate data. TODO: Investigate cleaner options for handling this.
  var rounded = new Date(Math.round(fsStatus.mtime.getTime() / ROUND_DATE) * ROUND_DATE);
  storageOptions.metadata = {'lastmodified': rounded.toISOString() };

  var summary = new storage.BlobService.SpeedSummary(blobName);
  storageOptions.speedSummary = summary;
  blobName = storageUtil.convertFileNameToBlobName(blobName);

  if (!cliUtils.fileExists(sourcefile, _)) {
    throw new Error(util.format($('Failed to find file \'%s\''), sourcefile));
  }
  if (!fsStatus.isFile()) {
    throw new Error(util.format($('%s is not a file'), sourcefile));
  }
  var sizeLimit = storageUtil.MaxBlockBlobSize;
  if (fsStatus.size > sizeLimit) {
    throw new Error(util.format($('The local file size %d exceeds the Azure blob size limit %d bytes'), fsStatus.size, sizeLimit));
  }

  var operation;
  try {
    //Check if blob already exists in storage, and compare the files lastmodified
    //date. We only upload and overwrite if the local file has not been uploaded
    //or is more recent.
    operation = storageUtil.getStorageOperation(blobService, storageUtil.OperationType.Blob, 'getBlobProperties');
    var props = storageUtil.performStorageOperation(operation, _, containerName, blobName);
    if (props.metadata.lastmodified) {
      var lastModified = new Date(props.metadata.lastmodified);
      if (lastModified >= rounded){
        logger.info(util.format($('File \'%s\' already exists and up-to-date - skipping'), blobName));
        return;
      }
    }

  } catch (e) {
    if (!e.code || e.code != 'NotFound') {
      throw e;
    }
  }

  var tips = util.format($('Uploading %s to blob %s in container %s'), sourcefile, blobName, containerName);
  operation = storageUtil.getStorageOperation(blobService, storageUtil.OperationType.Blob, 'createBlockBlobFromLocalFile');
  storageOptions.parallelOperationThreadCount = storageOptions.parallelOperationThreadCount;
  var printer = storageUtil.getSpeedPrinter(summary);
  var intervalId = -1;
  if (!logger.format().json) {
    intervalId = setInterval(printer, 1000);
  }
  startProgress(tips);
  try {
    //Upload block blob
    //TODO: Get this uploading in parallel.
    //TODO: Investigate compression + chunking performance enhancement proposal.
    storageUtil.performStorageOperation(operation, _, containerName, blobName, sourcefile, storageOptions);
  } catch (e) {
    printer(true);
    throw e;
  } finally {
    printer(true);
    clearInterval(intervalId);
    endProgress();
  }
};

module.exports = batchFileUtils;