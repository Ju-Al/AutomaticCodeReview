//
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

var azure = require('azure');
var StorageUtil = require('../../util/storage.util');
var util = require('util');
var utils = require('../../util/utils');
var validation = require('../../util/validation');
var commander = require('commander');
var fs = require('fs');
var path = require('path');
var blobUtils = require('../../util/blobUtils');
var splitDestinationUri = blobUtils.splitDestinationUri;
var Wildcard = utils.Wildcard;
var getStorageSettings = StorageUtil.getStorageServiceSettings;
var performStorageOperation = StorageUtil.performStorageOperation;
var startProgress = StorageUtil.startProgress;
var endProgress = StorageUtil.endProgress;
var setOperationTimeout = StorageUtil.setOperationTimeout;
var BlobConstants = azure.Constants.BlobConstants;
var SpeedSummary = azure.BlobService.SpeedSummary;

var $ = utils.getLocaleString;

/**
* Add storge account command line options
*/
commander.Command.prototype.addStorageAccountOption = function () {
  this.option('-a, --account-name <accountName>', $('the storage account name'));
  this.option('-k, --account-key <accountKey>', $('the storage account key'));
  this.option('-c, --connection-string <connectionString>', $('the storage connection string'));
  this.option('-vv', $('run storage command in debug mode'));
  return this;
};

/**
* Init storage blob command
*/
exports.init = function (cli) {

  //Init StorageUtil
  StorageUtil.init(cli);

  /**
  * Define storage blob command usage
  */
  var storage = cli.category('storage')
    .description($('Commands to manage your Storage objects'));

  var logger = cli.output;

  var container = storage.category('container')
    .description($('Commands to manage your Storage container'));

  container.command('list [prefix]')
    .description($('List storage containers with wildcard'))
    .option('-p, --prefix <prefix>', $('the storage container name prefix'))
    .addStorageAccountOption()
    .execute(listAzureContainersWithAcl);

  container.command('show [container]')
    .description($('Show details of the specified storage container'))
    .option('--container <container>', $('the storage container name'))
    .addStorageAccountOption()
    .execute(showAzureContainer);

  container.command('create [container]')
    .description($('Create a storage container'))
    .option('--container <container>', $('the storage container name'))
    .option('-p, --permission <permission>', $('the storage container ACL permission(Off/Blob/Container)'))
    .addStorageAccountOption()
    .execute(createAzureContainer);

  container.command('delete [container]')
    .description($('Delete the specified storage container'))
    .option('--container <container>', $('the storage container name'))
    .option('-q, --quiet', $('remove the specified Storage container without confirmation'))
    .addStorageAccountOption()
    .execute(deleteAzureContainer);

  container.command('set [container]')
    .description($('Set storage container ACL'))
    .option('--container <container>', $('the storage container name'))
    .option('-p, --permission <permission>', $('the storage container ACL permission(Off/Blob/Container)'))
    .addStorageAccountOption()
    .execute(setAzureContainer);

  var blob = storage.category('blob')
    .description($('Commands to manage your Storage blob'));

  blob.command('list [container] [prefix]')
    .usage('[options] [container] [prefix]')
    .description($('List storage blob in the specified storage container use wildcard and blob name prefix'))
    .option('--container <container>', $('the storage container name'))
    .option('-p, --prefix <prefix>', $('the blob name prefix'))
    .addStorageAccountOption()
    .execute(listAzureBlob);

  blob.command('show [container] [blob]')
    .usage('[options] [container] [blob]')
    .description($('Show details of the specified storage blob'))
    .option('--container <container>', $('the storage container name'))
    .option('-b, --blob <blobName>', $('the storage blob name'))
    .addStorageAccountOption()
    .execute(showAzureBlob);

  blob.command('delete [container] [blob]')
    .usage('[options] [container] [blob]')
    .description($('Delete the specified storage blob'))
    .option('--container <container>', $('the storage container name'))
    .option('-b, --blob <blobName>', $('the storage blob name'))
    //TODO
    //nodesdk don't support deleteBlob with snapshot http header
    //.option('-d, --deleteSnapshot', 'Delete the blob with snapshots')
    .option('-q, --quiet', $('remove the specified Storage blob without confirmation'))
    .addStorageAccountOption()
    .execute(deleteAzureBlob);

  blob.command('upload [file] [container] [blob]')
    .usage('[options] [file] [container] [blob]')
    .description($('Upload the specified file to storage blob'))
    .option('-f, --file <file>', $('the local file path'))
    .option('--container <container>', $('the storage container name'))
    .option('-b, --blob <blobName>', $('the storage blob name'))
    .option('-t, --blobtype <blobtype>', $('the storage blob type(Page, Block)'))
    .option('-p, --properties <properties>', $('the storage blob properties for uploaded file. Properties are key=value pairs and separated with semicolon(;). Available properties are contentType, contentEncoding, contentLanguage, cacheControl'))
    .option('-m, --metadata <metadata>', $('the storage blob metadata for uploaded file. Metadata are key=value pairs and separated with semicolon(;)'))
    .option('--concurrenttaskcount <concurrenttaskcount>', $('the maximum number of concurrent upload requests'))
    .option('-q, --quiet', $('overwrite the specified Storage blob without confirmation'))
    .addStorageAccountOption()
    .execute(uploadAzureBlob);

  blob.command('download [container] [blob] [destination]')
    .usage('[options] [container] [blob] [destination]')
    .description($('Download the specified storage blob'))
    .option('--container <container>', $('the storage container name'))
    .option('-b, --blob <blobName>', $('the storage blob name'))
    .option('-d, --destination [destination]', $('download destination file or directory path'))
    .option('-m, --checkmd5', $('check md5sum for the downloaded file'))
    .option('--concurrenttaskcount <concurrenttaskcount>', $('the maximum number of concurrent download requests'))
    .option('-q, --quiet', $('overwrite the destination file without confirmation'))
    .addStorageAccountOption()
    .execute(downloadAzureBlob);

  var copy = blob.category('copy')
    .description($('Commands to manage your blob copy operations'));

  copy.command('start [sourceUri] [destContainer]')
    .usage('[options] [sourceUri] [destContainer]')
    .description($('Start to copy the specified storage blob which completes asynchronously'))
    .option('--sas-token <sasToken>', $('the source storage SAS token'))
    .option('--source-container <sourceContainer>', $('the source storage container name'))
    .option('--source-blob <sourceBlob>', $('the source storage blob name'))
    .option('--source-uri <sourceUri>', $('the source storage blob absolute uri'))
    .option('--dest-account-name <destAccountName>', $('the destination storage account name'))
    .option('--dest-account-key <destAccountKey>', $('the destination storage account key'))
    .option('--dest-connection-string <destConnectionString>', $('the destination storage connection string'))
    .option('--dest-container <destContainer>', $('the destination storage container name'))
    .option('--dest-blob <destBlob>',$('the destination storage blob name'))
    .option('-q, --quiet',$('overwrite the destination blob without confirmation'))
    .addStorageAccountOption()
    .execute(startBlobCopy);

  copy.command('show [container] [blob]')
    .usage('[options] [container] [blob]')
    .description($('Show the copy status'))
    .option('--container <container>', $('the destination container in the blob copy start operation'))
    .option('--blob <blob>', $('the destination blob in the blob copy start operation'))
    .addStorageAccountOption()
    .execute(showBlobCopy);

  copy.command('stop [container] [blob] [copyid]')
    .usage('[options] [container] [blob] [copyid]')
    .description($('Stop the copy operation'))
    .option('--container <container>', $('the destination container in the blob copy start operation'))
    .option('--blob <blob>', $('the destination blob in the blob copy start operation'))
    .option('--copyid <copyid>', $('the copy ID which is returned from blob copy start operation'))
    .addStorageAccountOption()
    .execute(stopBlobCopy);

  /**
  * Implement storage blob cli
  */

  /**
  * Operation concurrency.
  *   -1 means operations are fully parallelized.
  *   However the concurrent REST calls are limited by performStorageOperation
  */
  var opConcurrency = -1;

  /**
  * Get Storage blob operation object
  * @param {string} [operationName] operation name
  * @return {StorageOperation} storage blob operation
  */
  function getStorageBlobOperation(serviceClient, operationName) {
    var operation = new StorageUtil.StorageOperation();
    operation.type = StorageUtil.OperationType.Blob;
    operation.operation = operationName;
    operation.service = serviceClient;
    return operation;
  }

  /**
  * Extract the storage account options from the specified options
  */
  function getStorageAccountOptions(options) {
    return {
      accountName: options.accountName,
      accountKey: options.accountKey,
      connectionString: options.connectionString
    };
  }

  /**
  * Get Storage blob operation options
  */
  function getStorageBlobOperationDefaultOption() {
    var option = {};
    setOperationTimeout(option);
    return option;
  }

  /**
  * Get blob service account from user specified credential or env variables
  */
  function getBlobServiceClient(options) {
    var serviceClient = StorageUtil.getServiceClient(StorageUtil.getBlobService, options);
      applyBlobServicePatch(serviceClient);
      return serviceClient;
    }


  /**
  * Add a space at the end of string and use single line prompt
  */
  function addTailSpace(str) {
    if (str && str.length && str[str.length - 1] != ' ') {
      str += ' ';
    }
    return str;
  }

  /**
  * Ask for input if the specified value is not given
  */
  function promptIfNotGiven(label, value, _) {
    label = addTailSpace(label);
    value = cli.interaction.promptIfNotGiven(label, value, _);
    return value;
  }

  /**
  * Confirm the specified action
  */
  function confirm(msg, _) {
    msg = addTailSpace(msg);
    var ok = cli.interaction.confirm(msg, _);
    if (!ok) {
      logger.warn($('The operation is cancelled'));
    }
    return ok;
  }

  /**
  * List storage container with acl
  * @param {string} prefix container prefix
  * @param {object} options commadline options
  * @param {callback} _ callback function
  */
  function listAzureContainersWithAcl(prefix, options, _) {
    var blobService = getBlobServiceClient(options);
    var listOperation = getStorageBlobOperation(blobService, 'listAllContainers');
    var tips = $('Getting storage containers');
    var containerOpts = getStorageBlobOperationDefaultOption();
    var useWildcard = false;
    containerOpts.include = 'metadata';

    if (Wildcard.containWildcards(prefix)) {
      containerOpts.prefix = Wildcard.getNonWildcardPrefix(prefix);
      useWildcard = true;
    } else {
      containerOpts.prefix = prefix;
    }

    var containers = [];
    startProgress(tips);

    try {
      /*jshint camelcase:false*/
      performStorageOperation(listOperation, _, containerOpts).forEach_(_, opConcurrency, function (_, container) {
        /*jshint camelcase:true*/
        if (useWildcard && !Wildcard.isMatch(container.name, prefix)) {
          return;
        }
        containers.push(container);
        var aclOperation = getStorageBlobOperation(blobService, 'getContainerAcl');
        var aclOptions = {};
        setOperationTimeout(aclOptions);
        try {
          var permission = performStorageOperation(aclOperation, _, container.name, aclOptions);
          var level = StorageUtil.containerAccessLevelToString(permission.publicAccessLevel);
          container.publicAccessLevel = level;
        } catch (e) {
          logger.warn(e.message || e);
        }
      });
    } finally {
      endProgress();
    }

    cli.interaction.formatOutput(containers, function (outputData) {
      if (outputData.length === 0) {
        logger.info($('No containers found'));
      } else {
        logger.table(outputData, function (row, item) {
          row.cell($('Name'), item.name);
          row.cell($('Public-Access'), item.publicAccessLevel);
          row.cell($('Last-Modified'), item.properties['last-modified']);
        });
      }
    });
  }

  /**
  * Show the details for the specified storage container
  * @param {string} container container name
  */
  function showAzureContainer(container, options, _) {
    var blobService = getBlobServiceClient(options);
    container = promptIfNotGiven($('Container name: '), container, _);
    var propertiesOperation = getStorageBlobOperation(blobService, 'getContainerProperties');
    var tips = $('Getting Storage container information');
    var showOptions = getStorageBlobOperationDefaultOption();
    var aclOperation = getStorageBlobOperation(blobService, 'getContainerAcl');
    var properties = {};

    startProgress(tips);

    try {
      //Get Container Properties operation returns all user-defined metadata and system properties for the specified container.
      properties = performStorageOperation(propertiesOperation, _, container, showOptions);
      var permission = performStorageOperation(aclOperation, _, container, showOptions);
      var level = StorageUtil.containerAccessLevelToString(permission.publicAccessLevel);
      properties.publicAccessLevel = level;
    } catch (e) {
      if (StorageUtil.isNotFoundException(e)) {
        throw new Error(util.format($('Container %s doesn\'t exist'), container));
      } else {
        throw e;
      }
    } finally {
      endProgress();
    }

    logger.json(properties);
  }


  /**
  * Create a storage container
  */
  function createAzureContainer(container, options, _) {
    var blobService = getBlobServiceClient(options);
    container = promptIfNotGiven($('Container name: '), container, _);
    var operation = getStorageBlobOperation(blobService, 'createContainerIfNotExists');
    var tips = util.format($('Creating storage container %s'), container);
    var storageOptions = getStorageBlobOperationDefaultOption();
    var permission = options.permission;
    if (permission) {
      validation.isValidEnumValue(permission, Object.keys(BlobConstants.BlobContainerPublicAccessType));
    }

    startProgress(tips);
    try {
      var created = performStorageOperation(operation, _, container, storageOptions);
      if (created === false) {
        throw new Error(util.format($('Container \'%s\' already exists'), container));
      } else if (permission) {
        var aclOperation = getStorageBlobOperation(blobService, 'setContainerAcl');
        var accessLevel = StorageUtil.stringToContainerAccessLevel(permission);
        performStorageOperation(aclOperation, _, container, null, accessLevel, storageOptions);
      }
    } finally {
      endProgress();
    }

    logger.verbose(util.format($('Container %s created successfully'), container));
    showAzureContainer(container, getStorageAccountOptions(options), _);
  }

  /**
  * Delete the specified storage container
  */
  function deleteAzureContainer(container, options, _) {
    var blobService = getBlobServiceClient(options);
    container = promptIfNotGiven($('Container name: '), container, _);
    var tips = util.format($('Deleting Container %s'), container);
    var operation = getStorageBlobOperation(blobService, 'deleteContainer');
    var storageOptions = getStorageBlobOperationDefaultOption();
    var force = !!options.quiet;

    if (force !== true) {
      force = confirm(util.format($('Do you want to remove the storage container %s?'), container), _);
      if (force !== true) {
        return;
      }
    }

    startProgress(tips);

    try {
      performStorageOperation(operation, _, container, storageOptions);
    } catch (e) {
      if (StorageUtil.isNotFoundException(e)) {
        throw new Error(util.format($('Can not find container \'%s\''), container));
      } else {
        throw e;
      }
    } finally {
      endProgress();
    }

    logger.info(util.format($('Container %s deleted successfully'), container));
  }

  /**
  * Set container acl(properties/metadata)
  */
  function setAzureContainer(container, options, _) {
    var blobService = getBlobServiceClient(options);
    container = promptIfNotGiven('Container name: ', container, _);
    var tips = 'Set container';
    startProgress(tips);

    try {
      if (options.permission) {
        setAzureContainerAcl(blobService, container, options.permission, _);
      }
    } finally {
      endProgress();
    }
    showAzureContainer(container, getStorageAccountOptions(options), _);
  }

  /**
  * Set container acl
  */
  function setAzureContainerAcl(blobService, container, permission, _) {
    var operation = getStorageBlobOperation(blobService, 'setContainerAcl');
    var storageOptions = getStorageBlobOperationDefaultOption();
    validation.isValidEnumValue(permission, Object.keys(BlobConstants.BlobContainerPublicAccessType));
    var accessLevel = StorageUtil.stringToContainerAccessLevel(permission);
    performStorageOperation(operation, _, container, null, accessLevel, storageOptions);
  }

  /**
  * List storage blob in the specified container
  */
  function listAzureBlob(container, blobName, options, _) {
    var blobService = getBlobServiceClient(options);
    var specifiedContainerName = promptIfNotGiven($('Container name:'), container, _);
    var tips = util.format($('Getting blobs in container %s'), specifiedContainerName);
    var operation = getStorageBlobOperation(blobService, 'listAllBlobs');
    var storageOptions = getStorageBlobOperationDefaultOption();
    var useWildcard = false;
    var inputBlobName = blobName;
    if (Wildcard.containWildcards(inputBlobName)) {
      storageOptions.prefix = Wildcard.getNonWildcardPrefix(inputBlobName);
      useWildcard = true;
    } else {
      storageOptions.prefix = inputBlobName;
    }
    storageOptions.include = 'snapshots,metadata,copy';
    var blobs = [];

    startProgress(tips);

    try {
      blobs = performStorageOperation(operation, _, specifiedContainerName, storageOptions);
    } finally {
      endProgress();
    }

    var outputBlobs = [];

    if (useWildcard) {
      for (var i = 0, len = blobs.length; i < len; i++) {
        var blob = blobs[i];
        if (Wildcard.isMatch(blob.name, inputBlobName)) {
          outputBlobs.push(blob);
        }
      }
    } else {
      outputBlobs = blobs;
    }

    cli.interaction.formatOutput(outputBlobs, function (outputData) {
      if (outputData.length === 0) {
        logger.info($('No blobs found'));
      } else {
        logger.table(outputData, function (row, item) {
          row.cell($('Name'), item.name);
          row.cell($('BlobType'), item.properties.blobtype);
          row.cell($('Length'), item.properties['content-length']);
          row.cell($('Content-Type'), item.properties['content-type']);
          row.cell($('Last-Modified'), item.properties['last-modified']);
          row.cell($('SnapshotTime'), item.snapshot || '');
        });
      }
    });
  }

  /**
  * Show the details of the specified Storage blob
  */
  function showAzureBlob(containerName, blobName, options, _) {
    var blob = getAzureBlobProperties(containerName, blobName, options, _);
    logBlobProperties(blob);
  }

  /**
  * Log blob properties
  */
  function logBlobProperties(properties) {
    if (!properties) return;
    cli.interaction.formatOutput(properties, function (data) {
      var outputProperties = ['container', 'blob', 'blobType', 'contentLength', 'contentType', 'contentMD5'];
      var output = outputProperties.map(function (propertyName) { return { property: propertyName, value: data[propertyName] }; });
      logger.table(output, function (row, item) {
        row.cell($('Property'), item.property);
        row.cell($('Value'), item.value);
      });
    });
  }

  /**
  * Get azure blob properties
  */
  function getAzureBlobProperties(container, blobName, options, _) {
    var blobService = getBlobServiceClient(options);
    var specifiedContainerName = promptIfNotGiven($('Container name:'), container, _);
    var specifiedBlobName = promptIfNotGiven($('Blob name:'), blobName, _);
    var storageOptions = getStorageBlobOperationDefaultOption();
    var blob = {};
    var propertiesOperation = getStorageBlobOperation(blobService, 'getBlobProperties');
    var tips = $('Getting Storage blob information');

    startProgress(tips);

    try {
      blob = performStorageOperation(propertiesOperation, _, specifiedContainerName, specifiedBlobName, storageOptions);
    } catch (e) {
      if (StorageUtil.isNotFoundException(e)) {
        throw new Error(util.format($('Blob %s in Container %s doesn\'t exist'), specifiedBlobName, specifiedContainerName));
      } else {
        throw e;
      }
    } finally {
      endProgress();
    }
    return blob;
  }

  /**
  * Show the details of the specified Storage blob
  */
  function deleteAzureBlob(container, blobName, options, _) {
    var blobService = getBlobServiceClient(options);
    var specifiedContainerName = promptIfNotGiven($('Container name:'), container, _);
    var specifiedBlobName = promptIfNotGiven($('Blob name:'), blobName, _);
    var storageOptions = getStorageBlobOperationDefaultOption();
    var tips = util.format($('Deleting Blob %s in container %s'), blobName, container);
    var operation = getStorageBlobOperation(blobService, 'deleteBlob');
    startProgress(tips);

    try {
      performStorageOperation(operation, _, specifiedContainerName, specifiedBlobName, storageOptions);
    } catch (e) {
      if (StorageUtil.isNotFoundException(e)) {
        throw new Error(util.format($('Can not find blob \'%s\' in container \'%s\''), specifiedBlobName, specifiedContainerName));
      } else {
        throw e;
      }
    } finally {
      endProgress();
    }

    logger.info(util.format($('Blob %s deleted successfully'), blobName));
  }

  /**
  * upload local file to blob
  */
  function uploadAzureBlob(file, container, blobName, options, _) {
    var blobService = getBlobServiceClient(options);
    var blobTypeName = options.blobtype || 'BLOCK';
    validation.isValidEnumValue(blobTypeName, Object.keys(BlobConstants.BlobTypes));
    var specifiedContainerName = promptIfNotGiven($('Container name:'), container, _);
    var specifiedFileName = promptIfNotGiven($('File name:'), file, _);
    var specifiedBlobName = blobName;
    var specifiedBlobType = BlobConstants.BlobTypes[blobTypeName.toUpperCase()];
    var storageOptions = getStorageBlobOperationDefaultOption();
    var properties = StorageUtil.parseKvParameter(options.properties);
    var force = options.quiet;
    storageOptions.metadata = StorageUtil.parseKvParameter(options.metadata);
    storageOptions.storeBlobContentMD5 = true;
    StorageUtil.formatBlobProperties(properties, storageOptions);
    var summary = new SpeedSummary(specifiedBlobName);
    storageOptions.speedSummary = summary;

    if (!specifiedBlobName) {
      specifiedBlobName = path.basename(specifiedFileName);
    }
    specifiedBlobName = StorageUtil.convertFileNameToBlobName(specifiedBlobName);

    if (!utils.fileExists(specifiedFileName, _)) {
      throw new Error(util.format($('Local file %s doesn\'t exist'), specifiedFileName));
    }
    var fsStatus = fs.stat(specifiedFileName, _);
    if (!fsStatus.isFile()) {
      throw new Error(util.format($('%s is not a file'), specifiedFileName));
    }
    var tips = '';
    if (force !== true) {
      var blobProperties = null;
      try {
        tips = util.format($('Checking blob %s in container %s'), specifiedBlobName, specifiedContainerName);
        startProgress(tips);
        var propertiesOperation = getStorageBlobOperation(blobService, 'getBlobProperties');
        blobProperties = performStorageOperation(propertiesOperation, _,
          specifiedContainerName, specifiedBlobName, storageOptions);
      } catch (e) {
        if (!StorageUtil.isNotFoundException(e)) {
          throw e;
        }
      } finally {
        endProgress();
      }

      if (blobProperties !== null) {
        if (blobProperties.blobType !== specifiedBlobType) {
          throw new Error(util.format($('BlobType mismatch. The current blob type is %s'),
            blobProperties.blobType));
        } else {
          if (!confirm(util.format($('Do you want to remove the blob %s in container %s?'),
  specifiedBlobName, specifiedContainerName), _)) {
            return;
          }
        }
      }
    }

    tips = util.format($('Uploading %s to blob %s in container %s'), specifiedFileName, specifiedBlobName, specifiedContainerName);
    var operation = getStorageBlobOperation(blobService, 'createBlockBlobFromLocalFile');
    storageOptions.parallelOperationThreadCount = options.concurrenttaskcount || StorageUtil.threadsInOperation;
    var printer = getSpeedPrinter(summary);
    var intervalId = -1;
    if (!logger.format().json) {
      intervalId = setInterval(printer, 1000);
    }
    startProgress(tips);
    endProgress();
    try {
      if (blobTypeName.toLowerCase() === 'page') {
        //Upload page blob
        operation = getStorageBlobOperation(blobService, 'createPageBlobFromLocalFile');
        performStorageOperation(operation, _, specifiedContainerName, specifiedBlobName, specifiedFileName, storageOptions);
      } else {
        //Upload block blob
        operation = getStorageBlobOperation(blobService, 'createBlockBlobFromLocalFile');
        performStorageOperation(operation, _, specifiedContainerName, specifiedBlobName, specifiedFileName, storageOptions);
      }
    } catch (e) {
      printer(true);
      throw e;
    } finally {
      printer(true);
      clearInterval(intervalId);
    }
    showAzureBlob(specifiedContainerName, specifiedBlobName, getStorageAccountOptions(options), _);
  }

  /**
  * Download storage blob
  */
  function downloadAzureBlob(container, blobName, destination, options, _) {
    var blobService = getBlobServiceClient(options);
    var specifiedContainerName = promptIfNotGiven($('Container name:'), container, _);
    //Default download destination is the current directory.
    var specifiedFileName = destination || '.';
    var specifiedBlobName = promptIfNotGiven($('Blob name:'), blobName, _);
    var dirName = '';
    var fileName = '';
    var isDirectory = false;
    var force = options.quiet;
    if (utils.pathExistsSync(specifiedFileName)) {
      var fsStatus = fs.stat(specifiedFileName, _);
      isDirectory = fsStatus.isDirectory();
    } else {
      if (specifiedFileName === '.' ||
          (specifiedFileName.length && specifiedFileName[specifiedFileName.length - 1] === path.sep)) {
        isDirectory = true;
      }
    }

    if (isDirectory) {
      dirName = specifiedFileName;
      fileName = '';
    } else {
      fileName = path.basename(specifiedFileName);
      dirName = path.dirname(specifiedFileName);
    }

    if (!utils.fileExists(dirName, _)) {
      throw new Error(util.format($('Local directory %s doesn\'t exist'), dirName));
    }

    if (!fileName) {
      var structure = StorageUtil.getStructureFromBlobName(specifiedBlobName);
      fileName = structure.fileName;
      fileName = utils.escapeFilePath(fileName);
      structure.dirName = StorageUtil.recursiveMkdir(dirName, structure.dirName);
      fileName = path.join(structure.dirName, fileName);
      dirName = '.'; //FileName already contain the dirname
    }

    var fullName = path.join(dirName, fileName);
    if (force !== true && utils.fileExists(fullName, _)) {
      if (!confirm(util.format($('Do you want to overwrite %s?'), fullName), _)) {
        return;
      }
    }
    var tips = util.format($('Download blob %s in container %s to %s'), specifiedBlobName, specifiedContainerName, fullName);
    var storageOptions = getStorageBlobOperationDefaultOption();
    var operation = getStorageBlobOperation(blobService, 'getBlobToLocalFile');
    storageOptions.parallelOperationThreadCount = options.concurrenttaskcount || StorageUtil.threadsInOperation;
    var summary = new SpeedSummary(specifiedBlobName);
    storageOptions.speedSummary = summary;
    storageOptions.checkMD5sum = options.checkmd5;

    startProgress(tips);
    endProgress();
    var printer = getSpeedPrinter(summary);
    var intervalId = -1;
    if (!logger.format().json) {
      intervalId = setInterval(printer, 1000);
    }
    var downloadBlob = {};
    try {
      downloadBlob = performStorageOperation(operation, _, specifiedContainerName, specifiedBlobName, fullName, storageOptions);
    } catch (e) {
      printer(true);
      if (StorageUtil.isNotFoundException(e)) {
        throw new Error(util.format($('Can not find blob \'%s\' in container \'%s\''), specifiedBlobName, specifiedContainerName));
      } else {
        throw e;
      }
    } finally {
      printer(true);
      clearInterval(intervalId);
    }

    if (options.checkmd5) {
      var calcTips = $('Calculating content md5');
      var blobProperties = {};
      startProgress(calcTips);
      try {
        var propertiesOperation = getStorageBlobOperation(blobService, 'getBlobProperties');
        blobProperties = performStorageOperation(propertiesOperation, _,
          specifiedContainerName, specifiedBlobName, storageOptions);
      } finally {
        endProgress();
      }

      if (!blobProperties.contentMD5) {
        logger.warn(util.format($('Blob contentMd5 is missing, and the local file md5 is %s'), downloadBlob.contentMD5));
      } else {
        if (blobProperties.contentMD5 === downloadBlob.contentMD5) {
          logger.info(util.format($('Md5checksum validation passed, and md5checksum is %s'), downloadBlob.contentMD5));
        } else {
          throw new Error(util.format($('Md5checksum validation failed. Blob md5 is %s, but local file md5 is %s'), blobProperties.contentMD5, downloadBlob.contentMD5));
        }
      }
    }
    var downloadedBlob = getAzureBlobProperties(specifiedContainerName, specifiedBlobName, getStorageAccountOptions(options), _);
    if (downloadedBlob) {
      downloadedBlob['fileName'] = fullName;
    }

    cli.interaction.formatOutput(downloadedBlob, function (data) {
      logger.info(util.format($('File saved as %s'), data.fileName));
    });
  }

  /**
  * Get a printer for speed summary
  */
  function getSpeedPrinter(summary) {
    var clearBuffer = new Buffer(79, 'utf8');
    clearBuffer.fill(' ');
    clearBuffer = clearBuffer.toString();
    var done = false;
    return function (newline) {
      if (logger.format().json || done) return;
      var tips = util.format($('Percentage: %s%% (%s/%s) Average Speed: %s Elapsed Time: %s '), summary.getCompletePercent(),
        summary.getCompleteSize(), summary.getTotalSize(), summary.getAverageSpeed(), summary.getElapsedSeconds());
      fs.writeSync(1, '\r' + clearBuffer + '\r');
      process.stdout.write(tips);
      if (newline) {
        process.stdout.write('
');
        done = true;
      }
    };
  }

  /**
  * Patch for azure node sdk
  */
  function applyBlobServicePatch(blobService) {

    /*
    * List all containers
    * NOTICE: All the caller should use the options parameter since it's just a internal implementation
    */
    blobService.listAllContainers = function (options, callback) {
      StorageUtil.listWithContinuation(blobService.listContainersSegmentedWithPrefix, blobService, options.prefix, null, options, callback);
    };

    /*
    * List all blobs
    * NOTICE: All the caller should use the options parameter since it's just a internal implementation
    */
    blobService.listAllBlobs = function (container, options, callback) {
      StorageUtil.listWithContinuation(blobService.listBlobsSegmentedWithPrefix, blobService, container, options.prefix, null, options, callback);
    };
  }

  /**
  * Get destination account options
  */
  function getDestAccountOptions(options) {
    var isNameDefined = options.destAccountName !== undefined;
    var isKeyDefined = options.destAccountKey !== undefined;
    var isConnectionStringDefined = options.destConnectionString !== undefined;
    var isAccountDefined = isNameDefined || isKeyDefined;
    var destOption = options;

    if (isConnectionStringDefined && isAccountDefined) {
      throw new Error($('Please only define one of them: 1. --dest-connection-string. 2 --dest-account-name and --dest-account-key'));
    } else {
      if (isConnectionStringDefined) {
        destOption.connectionString = options.destConnectionString;
      } else if (isAccountDefined) {
        if (isNameDefined && isKeyDefined) {
          destOption.connectionString = util.format('DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s', options.destAccountName, options.destAccountKey);
        } else {
          throw new Error($('Please set both --dest-account-name and --dest-account-key'));
        }
      }
    }

    return destOption;
  }

  /**
  * Start blob copy
  */
  function startBlobCopy(sourceUri, destContainer, options, _) {
    // Make up the source blob Uri
    var source = sourceUri;
    var sourceContainer = options.sourceContainer;
    var sourceBlob = options.sourceBlob;
    if (!sourceUri) {
      if (!options.sourceContainer && !options.sourceBlob) {
        source = promptIfNotGiven($('Source URI:'), sourceUri, _);
      } else {
        if (!options.sourceContainer) {
          sourceContainer = promptIfNotGiven($('Source container:'), sourceContainer, _);
        } else if (!options.sourceBlob) {
          sourceBlob = promptIfNotGiven($('Source blob:'), sourceBlob, _);
        }

        var settings = getStorageSettings(options);
        var host = settings._blobEndpoint.primaryHost;
        if (host === undefined) {
          throw new Error($('The blob endpoint is invalid'));
        }
        if (host.lastIndexOf('/') !== host.length - 1) {
          source = host + '/';
        } else {
          source = host;
        }
        if (sourceContainer !== '$root') {
          source += (sourceContainer + '/');
        }
        source += encodeURIComponent(sourceBlob);
      }
    }

    // Get parameters for the ACL checking
    var embeddedSas = false;
    var blobResource = splitDestinationUri(source);
    if (!sourceContainer) {
      sourceContainer = blobResource.container;
    }
    if (!sourceBlob) {
      if (!blobResource.blobName) {
        throw new Error($('The source blob is invalid'));
      }

      // Parse the blob name from URI when sourceUri is specified with SAS token.
      var index = blobResource.blobName.lastIndexOf('?');
      if (index != -1) {
        embeddedSas = true;
        blobResource.blobName = blobResource.blobName.substring(0, index);
      }
      sourceBlob = decodeURIComponent(blobResource.blobName);
    }

    // Make sure the read access to the source blob
    var sasToken;
    if (options.sasToken) {
      var isNameDefined = options.accountName !== undefined;
      var isKeyDefined = options.accountKey !== undefined;
      var isConnectionStringDefined = options.connectionString !== undefined;
      if (isConnectionStringDefined || isNameDefined || isKeyDefined) {
        throw new Error($('Please only define one of them: 1. --connection-string. 2 --account-name and --account-key 3. --sas-token'));
      } else {
        sasToken = options.sasToken;
      }
    } else if (!embeddedSas) {
      var sourceblobService = getBlobServiceClient(options);
      var aclOperation = getStorageBlobOperation(sourceblobService, 'getContainerAcl');
      var showOptions = getStorageBlobOperationDefaultOption();
      var permission = performStorageOperation(aclOperation, _, sourceContainer, showOptions);
      if (BlobConstants.BlobContainerPublicAccessType.OFF == permission.publicAccessLevel) {
        // Grant temporary SAS token to the source blob
        var sharedAccessPolicy = {
          AccessPolicy: {
            Permissions: BlobConstants.SharedAccessPermissions.READ,
            Expiry: azure.date.daysFromNow(7)
          },
        };
        sasToken = sourceblobService.generateSharedAccessSignature(sourceContainer,sourceBlob,sharedAccessPolicy);
        if (!sasToken) {
          throw new Error($('The source blob is not accessible'));
        }
      }
    }
    source += (sasToken === undefined ? '' : ('?' + sasToken));

    // Check destination information
    var destBlobName = options.destBlob ? options.destBlob : sourceBlob;
    var destContainerName = promptIfNotGiven($('Destination container name:'), destContainer, _);
    if (!destContainerName) {
      throw new Error($('The destination container name is required'));
    }

    // Start copy operation
    var destOption = getDestAccountOptions(options);
    var destBlobService = getBlobServiceClient(destOption);
    var operation = getStorageBlobOperation(destBlobService, 'startCopyBlob');
    var tips = util.format($('Start copying blob %s'), source);
    var storageOptions = getStorageBlobOperationDefaultOption();

    var copyOp;
    var retry;
    startProgress(tips);
    while (retry === true || retry === undefined) {
      try {
        if (retry === undefined && !options.quiet) {
          storageOptions.accessConditions = { 'if-none-match': '*' };
        } else {
          storageOptions.accessConditions = {};
        }
        retry = false;
        copyOp = performStorageOperation(operation,_,source,destContainerName,destBlobName,storageOptions);
      } catch (e) {
        if (StorageUtil.isBlobExistsException(e)) {
          retry = confirm(util.format($('Do you want to overwrite the existing blob %s in container %s? (Y/N)'),destBlobName,destContainerName),_);
        }
        if (!retry) {
          endProgress();
          throw e;
        }
      }
    }
    endProgress();

    cli.interaction.formatOutput(copyOp, function (outputData) {
      var output = [outputData];
      logger.table(output,function(row,item) {
        row.cell($('Copy ID'), item.copyId);
        row.cell($('Status'), item.copyStatus);
      });
    });
  }

  /**
  * Show blob copy status
  */
  function showBlobCopy(container, blob, options, _) {
    var destOption = getDestAccountOptions(options);
    var blobProps = getAzureBlobProperties(container,blob,destOption,_);

    if (blobProps.copyId === undefined) {
      throw new Error($('Can not find copy task on specified blob'));
    } else {
      cli.interaction.formatOutput(blobProps,function(outputData) {
        var output = [outputData];
        logger.table(output,function(row,item) {
          row.cell($('Copy ID'),item.copyId);
          row.cell($('Progress'),item.copyProgress);
          row.cell($('Status'),item.copyStatus);
          if (item.copyStatusDescription) {
            row.cell($('Description'),item.copyStatusDescription);
          }
        });
      });
    }
  }

  /**
  * Stop blob copy
  */
  function stopBlobCopy(container, blob, copyid, options, _) {
    var destOption = getDestAccountOptions(options);
    var blobService = getBlobServiceClient(destOption);
    var operation = getStorageBlobOperation(blobService, 'abortCopyBlob');
    var tips = util.format($('Stop copying blob %s in container %s of copy id %s'), blob, container, copyid);
    var storageOptions = getStorageBlobOperationDefaultOption();

    var stopOp;
    startProgress(tips);
    try {
      stopOp = performStorageOperation(operation, _, container, blob, copyid, storageOptions);
    } finally {
      endProgress();
    }

    logger.info(util.format($('Copying blob %s to container %s has been stopped successfully'), blob, container));
  }
};