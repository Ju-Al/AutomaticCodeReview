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

'use strict';
/*jshint esnext: true */

var __ = require('underscore');
var util = require('util');
var fs = require('fs');
var shellEscape = require('shell-escape');
var utils = require('../../util/utils');
var fileUtils = require('./batch.fileUtils');
var url = require('url');
var path = require('path');

var $ = utils.getLocaleString;

var batchTemplateUtils = {};
batchTemplateUtils.fileEgressEnvName = 'AZ_BATCH_FILE_UPLOAD_CONFIG';
batchTemplateUtils.rootFileUploadUrl = 'https://raw.githubusercontent.com/Azure/azure-xplat-cli/batch-beta';

/**
 * Constructs a command line for the start task/job prep task which will run the setup script.
 * @param {object} existingTask   The original start task/job preparation task
 * @param {array} commandInfo     The additional command info to add
 * @returns {object}              An updated start task/job preparation task
 */
batchTemplateUtils.constructSetupTask = function (existingTask, commandInfo) {
  var result;
  if(!__.isUndefined(existingTask)) {
    result = __.clone(existingTask);
  } else {
    result = {};
  }
  
  var commands = [];
  var resources = [];
  var isWindows;

  __.each(commandInfo, function (cmd) {
    if (!__.isUndefined(cmd)) { 
      commands.push(cmd.cmdLine);
      if (!__.isUndefined(cmd.resourceFiles)) {
        resources = __.union(resources, cmd.resourceFiles);
      }
      if (__.isUndefined(isWindows)) {
        isWindows = cmd.isWindows;
      } else {
        if (isWindows !== cmd.isWindows) {
          throw new Error('The command is not compatiable with Windows and Linux.');
        }
      }
    }
  });

  if (!__.isUndefined(result.commandLine)) {
    commands.push(result.commandLine);
  }

  if (__.isEmpty(commands)) {
    return undefined;
  }

  if (!__.isUndefined(result.resourceFiles)) {
    resources = __.union(resources, result.resourceFiles);
  }

  if (isWindows) {
    var fullWinCommand = commands.join(' & ');
    result.commandLine = util.format('cmd.exe /c "%s"', fullWinCommand);
  } else {
    // Escape the users command line
    var fullLinuxCommand = shellEscape([commands.join(';')]);
    result.commandLine = util.format('/bin/bash -c %s', fullLinuxCommand); 
  }

  if (!__.isEmpty(resources)) {
    result.resourceFiles = resources;
  }

  // Must run elevated and wait for success for the setup step
  result.runElevated = true;
  result.waitForSuccess = true;

  return result;
};

/**
 * Process a task's outputFiles section and update the task accordingly.
 * @param {object} task           A task specification
 * @returns {object}              A new task object with modifications
 */
batchTemplateUtils.parseTaskOutputFiles = function (task) {
  if (__.isUndefined(task.outputFiles)) {
    return task;
  }

  var newTask = __.omit(task, 'outputFiles');
  
  // Validate the output file configuration
  __.each(task.outputFiles, function (outputFile) {
    if (__.isUndefined(outputFile.filePattern)) {
      throw new Error('outputFile must include filePattern');
    }
    if (__.isUndefined(outputFile.destination)) {
      throw new Error('outputFile must include destination');
    }
    if (__.isUndefined(outputFile.uploadDetails)) {
      throw new Error('outputFile must include uploadDetails');
    }
    if (__.isUndefined(outputFile.destination.container)) {
      throw new Error('outputFile.destination must include container');
    }
    if (__.isUndefined(outputFile.destination.container.containerSas)) {
      throw new Error('outputFile.destination.container must include containerSas');
    }
    if (__.isUndefined(outputFile.uploadDetails.taskStatus)) {
      throw new Error('outputFile.uploadDetails must include taskStatus');
    }
  });

  // Edit the command line to include output file upload at the end
  var uploadCommandLine = '$AZ_BATCH_JOB_PREP_WORKING_DIR/uploadfiles.sh';

  // Escape the user command line
  newTask.commandLine = shellEscape([util.format('%s;%s > $AZ_BATCH_TASK_DIR/uploadlog.txt 2>&1', newTask.commandLine, uploadCommandLine)]);

  newTask.commandLine = util.format('/bin/bash -c %s', newTask.commandLine);

  var config = { outputFiles: task.outputFiles };
  var configStr = JSON.stringify(config);
  if (__.isUndefined(newTask.environmentSettings)) {
    newTask.environmentSettings = [];
  }
  newTask.environmentSettings.push({ name: batchTemplateUtils.fileEgressEnvName, value: configStr });

  return newTask;
};

/**
 * Process a job and its collection of tasks for any tasks which use outputFiles. If a task does use outputFiles, we 
 * add to the jobs jobPrepTask for the install step.
 * NOTE: This edits the task collection in-line!

 * @param {object} job            A job specification
 * @param {object} tasks          A collection of tasks
 * @returns {object}              A new job object with modifications
 */
batchTemplateUtils.processJobForOutputFiles = function(job, tasks) {
  var mustEditJob = false;
  for(var i = 0; i < tasks.length; i++) {
    var originalTask = tasks[i];
    tasks[i] = batchTemplateUtils.parseTaskOutputFiles(tasks[i]);
    if(originalTask !== tasks[i]) {
      mustEditJob = true;
    }
  }

  if(mustEditJob) {
    var setupCommandLine = 'setup.sh';
    var command = util.format('%s > $AZ_BATCH_JOB_PREP_DIR/uploadsetuplog.txt 2>&1', setupCommandLine);
    var resourceFiles = [];

    let rootUrl = batchTemplateUtils.rootFileUploadUrl;
    if(!__.isUndefined(process.env.FILE_EGRESS_OVERRIDE_URL)) {
      rootUrl = process.env.FILE_EGRESS_OVERRIDE_URL;
    }

    // TODO: If we have any issues with this being hosted in GITHUB we'll have to move it elsewhere (local and then upload to their storage?)
    resourceFiles.push({
      blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/batchfileuploader.py'), 
      filePath: 'batchfileuploader.py' 
    });
    resourceFiles.push({ 
      blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/configuration.py'), 
      filePath: 'configuration.py' 
    });
    resourceFiles.push({ 
      blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/requirements.txt'), 
      filePath: 'requirements.txt' 
    });
    resourceFiles.push({ 
      blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/setup.sh'), 
      filePath: 'setup.sh' 
    });
    resourceFiles.push({
       blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/uploadfiles.sh'), 
       filePath: 'uploadfiles.sh' 
      });
    resourceFiles.push({ 
      blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/uploader.py'), 
      filePath: 'uploader.py' 
    });
    resourceFiles.push({ 
      blobSource: url.resolve(rootUrl, 'lib/commands/batch/fileegress/util.py'), 
      filePath: 'util.py' 
    });

    return { cmdLine: command, isWindows: false, resourceFiles: resourceFiles};
  }

  return undefined;
};

batchTemplateUtils.processResourceFiles = function (request, _) {
  //TODO: This wont work if request is an array of objects
  Object.keys(request).forEach_(_, 1, function (_, key) {
    if ((key === 'resourceFiles' || key === 'commonResourceFiles') && request[key] && typeof(request[key]) === 'object') {
      var newResources = [];
      request[key].forEach_(_, 1, function (_, resource) {
        fileUtils.resolveResourceFile(resource, {}, _).forEach(function(newResource){
          newResources.push(newResource);
        });
      });
      request[key] = newResources;
    } else if (typeof(request[key]) === 'object') {
      request[key] = batchTemplateUtils.processResourceFiles(request[key], _);
    }
  });
  return request;
};

/**
 * Parse task collection task factory object, and return task list
 * @param {object} taskFactory      The task factory as JSON object 
 */
batchTemplateUtils.parseTaskCollectionTaskFactory = function (taskFactory) {
  var tasks = taskFactory.tasks;

  if (!__.isArray(tasks) || __.isEmpty(tasks)) {
    throw new Error($('No task is defined in taskCollection task factory.'));
  }

  var taskObjs = [];
  __.each(tasks, function(task) {
    if (__.isUndefined(task.id) || __.isUndefined(task.commandLine)) {
      throw new Error('id and commandLine is required for task.');
    }

    var newTask = {
      id: task.id,
      commandLine: task.commandLine
    };

    var propertiesToCopy = __.pick(task, 'displayName', 'resourceFiles', 'envrionmentSettings', 'constraints', 'runElevated', 
      'multiInstanceSettings', 'dependsOn', 'exitConditions', 'clientExtensions', 'outputFiles', 'packageReferences');
    __.extend(newTask, propertiesToCopy);

    taskObjs.push(newTask);
  });
  return taskObjs;
};

/**
 * Parse parametric sweep set, and return all possible values in array
 * @param {object} parameterSets            The sweep in JSON object
 */
batchTemplateUtils.parseParameterSets = function (parameterSets) {
  var parameterList = [];

  if (!__.isArray(parameterSets) || parameterSets.length < 1) {
    throw new Error('No parameter set is defined.');
  }

  __.each(parameterSets, function(parameterSet) {
    if (__.isUndefined(parameterSet.start)) {
      throw new Error($('No start in parameter set'));
    }

    var start = parseInt(parameterSet.start);
    if (__.isUndefined(parameterSet.end)) {
      throw new Error($('No end in parameter set'));
    }
    
    var end = parseInt(parameterSet.end);
    var step = 1;
    if (!__.isUndefined(parameterSet.step)) {
      step = parseInt(parameterSet.step);
      if (step === 0) {
        throw new Error($('Step in parameter set can not be 0'));
      }
    }

    if (start > end && step > 0) {
      throw new Error($('Step has to be negative number when end is greater than start.'));
    }
    if (start < end && step < 0) {
      throw new Error($('Step has to be positive number when end is less than start.'));
    }

    var list = [];
    var i;

    if (start <= end) {
      for (i = start; i <= end; i += step) {
        list.push(i);
      }
    } else {
      for (i = start; i >= end; i += step) {
        list.push(i);
      }
    }

    parameterList.push(list);
  });

  return parameterList;
};

// A full permutation helper function on parameters input
// parameters : array of values would like to do full permutation
// level, list : helper parameter, should start with 0 and []
// callback: the callback function to be called when each permutation is generated
function permutations(parameters, level, list, callback) {
  if (level >= parameters.length) {
    callback(list);
  } else {
    for (var i = 0; i < parameters[level].length; i++) {
      var newlist = list.slice();
      newlist.push(parameters[level][i]);
      permutations(parameters, level + 1, newlist, callback);
    }
  }
}

//
// Parse the repeat task JSON object
//
function parseRepeatTask(task) {
  if (task.id) {
    throw new Error('should not have id for repeat task.');
  }

  if (__.isUndefined(task.commandLine)) {
    throw new Error('commandLine is required for task.');
  }

  var newTask = {
    commandLine: task.commandLine
  };

  var propertiesToCopy = __.pick(task, 'displayName', 'resourceFiles', 'environmentSettings', 'constraints', 'runElevated', 
    'exitConditions', 'clientExtensions', 'outputFiles', 'packageReferences');
  __.extend(newTask, propertiesToCopy);
  return newTask;
}

//
// Reverse string
//
function reverseString(s) {
  for (var i = s.length - 1, o = ''; i >= 0; ) {
    o += s[i--];
  }
  return o;
}

//
// This is to support non-ARM-style direct parameter string substitution as a 
// simplification of the concat function. We may wish to remove this 
// if we want to adhere more strictly to ARM.
//
function isSubstitution(template, start, end) {
  if (template[start-1] === '"' && template[end+1] === '"') {
    return false;
  }
  return true;
}

/**
 * Replace string with sweep value
 * @param {strung} data - The string to be replaced
 * @param {array} parameters - The sweep value, each value map to {0}, {1}, .. {n} 
 */
batchTemplateUtils.replacementParameter = function (data, parameters) {
  // By design, user should escape all the literal '{' or '}' to '{{' or '}}'
  // All other '{' or '}' characters are used for replacement

  // Handle '{' and '}' escape scenario : replace '{{' to LEFT_BRACKET_REPLACE_CHAR, and '}}' to RIGHT_BRACKET_REPLACE_CHAR
  // The reverse function is used to handle {{{0}}}
  if (__.isUndefined(data)) {
    return data;
  }
  //The first code unit of a surrogate pair is always in the range from 0xD800 to 0xDBFF, this means we cannot use \uD800 and \uD801 as the 
  //replacement characters because they are part of a pair. If we do use it, then sometimes they show up as 65535 instead of D800.
  var LEFT_BRACKET_REPLACE_CHAR = '\uE800';
  var RIGHT_BRACKET_REPLACE_CHAR = '\uE801';
  data = reverseString(reverseString(data.replace(/\{\{/g, LEFT_BRACKET_REPLACE_CHAR)).replace(/\}\}/g, RIGHT_BRACKET_REPLACE_CHAR));
  // Handle {n} or {n:m} scenario
  var reg = /\{(\d+)(:(\d+))?\}/g;
  data = data.replace(reg, function(r, r1, r2, r3) {
    var n = parseInt(r1);
    if (n >= parameters.length) {
      throw new Error(util.format('The parameter pattern \'%s\' is out of bound.', r));
    }
    var numberStr = parameters[n].toString();
    if (r.indexOf(':') > -1) {
      // This is {n:m} scenario
      if (parameters[n] < 0) {
        throw new Error(util.format('The parameter \'%s\' is negative and cannot be used in pattern \'%s\'.', parameters[n], r));
      }
      var m = parseInt(r3);
      if (r3 < 1 || r3 > 9) {
        throw new Error(util.format('The parameter pattern \'%s\' is out of bound. The padding number can be only between 1 to 9.', r));
      }

      var padding = numberStr.length >= m ? numberStr : new Array(m - numberStr.length + 1).join('0') + numberStr;
      return padding;
    }
    else {
      // This is just {n} scenario
      return numberStr;
    }
  });

  if ((data.indexOf('{') > -1) || (data.indexOf('}') > -1)) {
    throw new Error('Invalid use of bracket characters, did you forget to escape (using {{}})?');
  }
  // Replace LEFT_BRACKET_REPLACE_CHAR back to '{', and RIGHT_BRACKET_REPLACE_CHAR back to '}'
  var lbReg = new RegExp(LEFT_BRACKET_REPLACE_CHAR, 'g');
  var rbReg = new RegExp(RIGHT_BRACKET_REPLACE_CHAR, 'g');

  data = data.replace(rbReg, '}').replace(lbReg, '{');

  return data;
};

/**
 * Parse parametric sweep task factory object, and return task list
 * @param {object} taskFactory      The task factory as JSON object 
 */
batchTemplateUtils.parseParametricSweep = function (taskFactory) {
  if (!__.isArray(taskFactory.parameterSets) || __.isEmpty(taskFactory.parameterSets)) {
    throw new Error($('No parameter set is defined in parametric sweep task factory.'));
  }

  if (__.isUndefined(taskFactory.repeatTask)) {
    throw new Error($('No repeat task is defined in parametric sweep task factory.'));
  }

  var parameters = batchTemplateUtils.parseParameterSets(taskFactory.parameterSets);
  var repeatTask = parseRepeatTask(taskFactory.repeatTask);

  var taskObjs = [];
  permutations(parameters, 0, [], function(replacements) {
    // Need deep clone here
    var newTask = JSON.parse(JSON.stringify(repeatTask));
    __.extend(newTask, { id: taskObjs.length.toString() });

    newTask.commandLine = batchTemplateUtils.replacementParameter(newTask.commandLine, replacements);

    if (newTask.displayName) {
      newTask.displayName = batchTemplateUtils.replacementParameter(newTask.displayName, replacements);
    }

    var k;
    if (newTask.resourceFiles) {
      newTask.resourceFiles.forEach(function (resource) {
        if (resource.source) {
          resource.source.fileGroup = batchTemplateUtils.replacementParameter(resource.source.fileGroup, replacements);
          resource.source.prefix = batchTemplateUtils.replacementParameter(resource.source.prefix, replacements);
          resource.source.containerUrl = batchTemplateUtils.replacementParameter(resource.source.containerUrl, replacements);
          resource.source.url = batchTemplateUtils.replacementParameter(resource.source.url, replacements);
        } else {
          resource.blobSource = batchTemplateUtils.replacementParameter(resource.blobSource, replacements);
        }
        resource.filePath = batchTemplateUtils.replacementParameter(resource.filePath, replacements);
      });
    }
    
    if (newTask.environmentSettings) {
      for (k = 0; k < newTask.environmentSettings.length; k++) {
        newTask.environmentSettings[k].name = batchTemplateUtils.replacementParameter(newTask.environmentSettings[k].name, replacements);
        newTask.environmentSettings[k].value = batchTemplateUtils.replacementParameter(newTask.environmentSettings[k].value, replacements);
      }
    }

    if (newTask.clientExtensions && newTask.clientExtensions.dockerOptions) {
      newTask.clientExtensions.dockerOptions.image = 
        batchTemplateUtils.replacementParameter(newTask.clientExtensions.dockerOptions.image, replacements);

      if (newTask.clientExtensions.dockerOptions.dataVolumes) {
        for (k = 0; k < newTask.clientExtensions.dockerOptions.dataVolumes.length; k++) {
          newTask.clientExtensions.dockerOptions.dataVolumes[k].hostPath = 
            batchTemplateUtils.replacementParameter(newTask.clientExtensions.dockerOptions.dataVolumes[k].hostPath, replacements);
          newTask.clientExtensions.dockerOptions.dataVolumes[k].containerPath = 
            batchTemplateUtils.replacementParameter(newTask.clientExtensions.dockerOptions.dataVolumes[k].containerPath, replacements);
        }
      }
      if (newTask.clientExtensions.dockerOptions.sharedDataVolumes) {
        for (k = 0; k < newTask.clientExtensions.dockerOptions.sharedDataVolumes.length; k++) {
          newTask.clientExtensions.dockerOptions.sharedDataVolumes[k].name = 
            batchTemplateUtils.replacementParameter(newTask.clientExtensions.dockerOptions.sharedDataVolumes[k].name, replacements);
          newTask.clientExtensions.dockerOptions.sharedDataVolumes[k].containerPath = 
            batchTemplateUtils.replacementParameter(newTask.clientExtensions.dockerOptions.sharedDataVolumes[k].containerPath, replacements);
        }
      }
    }
    if(newTask.outputFiles) {
      newTask.outputFiles.forEach(function(outputFile) {
        outputFile.filePattern = batchTemplateUtils.replacementParameter(outputFile.filePattern, replacements);
        if(outputFile.destination && outputFile.destination.container) {
          outputFile.destination.container.path = batchTemplateUtils.replacementParameter(outputFile.destination.container.path, replacements);
          outputFile.destination.container.containerSas = batchTemplateUtils.replacementParameter(outputFile.destination.container.containerSas, replacements);
        }
      });
    }
    taskObjs.push(newTask);
  });

  if (taskFactory.mergeTask) {
    var mergeTask = parseRepeatTask(taskFactory.mergeTask);
    __.extend(mergeTask, { id: 'merge' });
    __.extend(mergeTask, { dependsOn: { taskIdRanges: { start : 0, end : taskObjs.length -1 }} });

    taskObjs.push(mergeTask);
  }
  return taskObjs;
};

/**
 * Render the content of an ARM property
 * @param {string} name - name of the property to render
 * @param {object} template - the template we're rendering
 * @param {object} parameters - parameter values used by the template
 */
function parseArmParameter(name, template, parameters)
{
  if (!template.parameters){
    throw new Error(util.format('Template defines no parameters but tries to use \'%s\'.', name));
  }

  var parameterDefinition =  template.parameters[name];
  if (!parameterDefinition) {
    throw new Error(util.format('Template does not define parameter \'%s\'.', name));
  } 

  var userValue = parameterDefinition.defaultValue;
  if (parameters && parameters[name]) {
    // Support both ARM and dictionary syntax
    if (parameters[name].value) {
      // ARM: '<<propertyName>>' : { 'value' : '<<propertyValue>>' }
      userValue = parameters[name].value;
    } else {
      // Dictionary: '<<propertyName>>' : <<propertyValue>>'
      userValue = parameters[name];
    }
  }

  if (!userValue) {
    throw new Error(util.format('No value supplied for parameter \'%s\' and no default value.', name));
  }

  if (typeof(userValue) === 'object') {
    // If substitute value is a complex object - it may require 
    // additional parameter substitutions
    return batchTemplateUtils.parseTemplate(JSON.stringify(userValue), template, parameters);
  }

  if (parameterDefinition.type) {
    // Enforce the declared type of the parameter
    // Earlier validation should catch any errors so these shouldn't throw'

    if (parameterDefinition.type === 'int') {
      return parseInt(userValue);
    }

    if (parameterDefinition.type === 'bool') {
      if (typeof(userValue) == 'boolean') {
        return userValue;
      }

      return userValue === 'true';
    }

    if (parameterDefinition.type === 'string') {
      return String(userValue);
    }

    throw new Error(util.format('Parameter type \'%s\' is not supported', parameterDefinition.type));
  }

  return userValue;
}

/**
 * Render the value of an ARM variable
 * @param {string} name - name of the variable to render
 * @param {object} template - the template we're rendering
 * @param {object} parameters - parameter values used by the template
 * @returns {various} contents of variable, strongly typed
 */
function parseArmVariable(name, template, parameters) {
  const variable = batchTemplateUtils.parseARMStyleTemplate(template.variables[name], template, parameters);
  if (typeof(variable) === 'object') {
    // If substitute value is a complex object - it may require 
    // additional parameter substitutions
    return  batchTemplateUtils.parseTemplate(JSON.stringify(variable), template, parameters);
  }

  return variable;
}

/**
 * Given that a string starts at the index specified, scan for the end of that string
 * @param {char} delimiter - delimiter for which to search
 * @param {string} content - string to scan
 * @param {int} startIndex - index of the character after opening string delimiter
 * @returns {int} index of the closing string delimiter 
 */
function findEndOfString(delimiter, content, startIndex) {
  var index = startIndex;
  while(index < content.length) {
    if (content[index] === '\\') {
      // Found escaped character, skip it
      index++;
    } else if (content[index] === delimiter) {
      return index;
    }
    index++;
  }
  return index;  
}

/**
 * Scan a string to find a specified delimiter, respecting nesting of brackets and strings
 * @param {char} delimiter - delimiter for which to search
 * @param {string} content - string to scan
 * @param {int} startIndex - first candidate character to check for the delimiter
 * @returns {int} index of the delimiter when found 
 */
function findDelimiter(delimiter, content, startIndex) {
  var index = startIndex;
  while (index < content.length) {
    switch (content[index]) {
      case delimiter:
        return index;

      case '(':
        index = findDelimiter(')', content, index+1);
        break;

      case '[':
        index = findDelimiter(']', content, index+1);
        break;

      case '"':
        index = findEndOfString('"', content, index+1);
        break;

      case '\'':
        index = findEndOfString('\'', content, index+1);
        break;
    }
    index++;
  }

  return index;
}

/**
 * Evaluate the concat() function
 * @param {string} content - text passed to the function for evaluation
 * @param {object} template - the template we're rendering
 * @param {object} parameters - parameter values used by the template
 * @returns {string} result of concatenation
 */
function parseArmConcat(content, template, parameters) {
  var index = 0;
  var result = '';
  while (index < content.length) {
    var end = findDelimiter(',', content, index);
    var argument = content.substring(index, end).trim();
    var evaluated = batchTemplateUtils.parseARMStyleTemplate(argument, template, parameters);
    result = result + evaluated;
    index = end + 1;
  }

  return result;
}

/**
 * Determine if a section of the template is an ARM reference, and calculate
 * the replacement accordingly. The result will be correctly typed to suit the 
 * parameter definition (e.g. will return a number if the parameter requires a number)
 * @param {string} value - a section of the template contained within []
 * @param {object} jsonTemplate - contents of the template file as an object
 * @param {object} properties - contents of the parameters file as an object
 */
batchTemplateUtils.parseARMStyleTemplate = function (value, template, properties) {
  if (!isNaN(value)) {
    return value;
  }
  if (value[0] === '[' && value[value.length - 1] === ']') {
    // Remove the enclosing brackets to check the contents
    value = value.substring(1, value.length - 1);
    return batchTemplateUtils.parseARMStyleTemplate(value, template, properties);
  }
  
  if (value[0] === '(' && value[value.length - 1] === ')') {
    // If the section is surrounded by ( ), then we need to further process the contents
    // as either a parameter name, or a concat operation
    return batchTemplateUtils.parseARMStyleTemplate(value.substring(1, value.length - 1), template, properties);
  }
  
  if (value[0] === '\'' && value[value.length - 1] === '\'') {
    // If a string, remove quotes in order to perform parameter look-up
    return value.substring(1, value.length - 1);
  }

  var subValue;
  if (new RegExp('^parameters').test(value)) {
    const parameterName  = value.substring(12, value.length - 2);
    return parseArmParameter(parameterName, template, properties);

  } else if (new RegExp('^variables').test(value)) {
    const variableName = value.substring(11, value.length - 2);
    return parseArmVariable(variableName, template, properties);

  } else if (new RegExp('^concat').test(value)) {
    const concatArguments = value.substring(7, value.length - 1);
    return parseArmConcat(concatArguments, template, properties);

  } else if (new RegExp('^reference').test(value)) {
    throw new Error(util.format('ARM-style \'reference\' syntax not supported.', subValue));

  } else {
    return value;
  }
};

/**
 * Parse task factory object, and return task list
 * @param {object} job      The job as JSON object 
 */
batchTemplateUtils.parseTaskFactory = function (job) {
  var taskFactory = job.taskFactory;
  if (__.isUndefined(taskFactory.type)) {
    throw new Error('No type property in taskFactory.');
  }

  delete job.taskFactory;

  if (taskFactory.type == 'parametricSweep') {
    return batchTemplateUtils.parseParametricSweep(taskFactory);
  } else if (taskFactory.type == 'taskCollection') {
    return batchTemplateUtils.parseTaskCollectionTaskFactory(taskFactory);
  } else {
    throw new Error(util.format('%s is not a valid task factory', taskFactory.type));
  }
};

/**
 * Given a string value (including quotes), evaluate any embedded template expressions delimited 
 * by '[' and ']'.'
 * @param {string} template - contents of the template string
 * @param {object} jsonTemplate - contents of the template as an object (contains parameter 
 * definitions, variables, and so on)
 * @param {object} parameters - contents of the parameters file as an object */
function parseTemplateString(template, jsonTemplate, parameters) {
  let updatedString = '';
  let currentIndex = 0;
  while (currentIndex < template.length) {

    const expressionStart = template.indexOf('[', currentIndex);
    if (expressionStart === -1) {
      // No template expression to evaluate
      break;
    }

    if (expressionStart < template.length-1 && template[expressionStart+1] === '[') {
      // Found escaped expression
      updatedString = updatedString + template.substring(currentIndex, expressionStart) + '[';
      currentIndex = expressionStart+2;
      continue;
    } 

    const  expressionFinish = findDelimiter(']', template, expressionStart+1);
    if (expressionFinish >= template.length) {
      // No closing delimiter for the expression (not our problem)
      break;
    }

    const expression = template.substring(expressionStart+1, expressionFinish); // everything between [ and ]
    const parsed = batchTemplateUtils.parseARMStyleTemplate(expression, jsonTemplate, parameters);
    if (isSubstitution(template, expressionStart, expressionFinish)) {
      // Replacing within the middle of a string
      updatedString = updatedString + template.substring(currentIndex, expressionStart) + parsed;
      currentIndex = expressionFinish+1;

    } else if (typeof(parsed) === 'number' || typeof(parsed) === 'boolean') {
      // Replacing an entire element value, and we want to remove any surrounding quotes
      updatedString = updatedString + template.substring(currentIndex, expressionStart -1 ) + parsed;
      currentIndex = expressionFinish+2;

    } else if (typeof(parsed) === 'object') {
      let json = JSON.stringify(parsed);
      updatedString = updatedString + template.substring(currentIndex, expressionStart -1 ) + json;
      currentIndex = expressionFinish+2;

    } else {
      updatedString = updatedString + template.substring(currentIndex, expressionStart) + parsed;
      currentIndex = expressionFinish+1;
    }
  }
  
  if (currentIndex < template.length) {
    updatedString = updatedString + template.substring(currentIndex);
  }

  return updatedString;
}

/**
 * Expand all parameters, and variables in the template.
 * @param {string} template - contents of the template file as a string
 * @param {object} jsonTemplate - contents of the template file as an object
 * @param {object} parameters - contents of the parameters file as an object
 */
batchTemplateUtils.parseTemplate = function (template, jsonTemplate, parameters) {

  /*
   * We want to expand all template expressions (delimited by '[' and ']') in the supplied template 
   * string. However, that syntax collides with JSON syntax for arrays and we don't want to collide 
   * with any of those. To avoid such a collision, we iterate through all of the string values 
   * (delimited by double quotes (")) and then expand template expressions only within those. 
   */
  let updatedJson = '';
  let currentIndex = 0;
  while (currentIndex < template.length) {

    const stringStart = template.indexOf('"', currentIndex);
    if (stringStart === -1) {
      // Didn't find another string to expand
      break;
    }

    const stringFinish = findEndOfString('"', template, stringStart+1);
    if (stringFinish >= template.length) {
      // Didn't find terminating quote for string (not our problem)
      break;
    }

    // Get the content of the string including the quotes
    const stringContent = template.substring(stringStart, stringFinish+1);
    if (stringContent.includes('[')) {
      const parsed = parseTemplateString(stringContent, jsonTemplate, parameters);
      if (stringStart > currentIndex) {
        updatedJson = updatedJson + template.substring(currentIndex, stringStart);
      }
      updatedJson = updatedJson + parsed;
    } else {
      updatedJson = updatedJson + template.substring(currentIndex, stringFinish+1);
    }

    currentIndex = stringFinish+1;
  }

  if (currentIndex < template.length) {
      updatedJson = updatedJson + template.substring(currentIndex);
  }

  return JSON.parse(updatedJson);
};

/**
 * These properties are reserved for application template use
 * and may not be used on jobs using an application template
 */
batchTemplateUtils.propertiesReservedForTemplateUse = [
  'jobManagerTask', 
  'jobPreparationTask', 
  'jobReleaseTask', 
  'commonEnvironmentSettings', 
  'usesTaskDependencies',
  'onAllTasksComplete', 
  'onTaskFailure', 
  'taskFactory'
];

batchTemplateUtils.propertiesPermittedOnTemplates = __.union([
  'templateMetadata',
  'parameters',
  'metadata'
],
batchTemplateUtils.propertiesReservedForTemplateUse);

/**
 * These properties are reserved for job use 
 * and may not be used on an application template
 */
batchTemplateUtils.propertiesReservedForJobUse = [
  'id', 
  'displayName',
  'priority',
  'constraints', 
  'poolInfo', 
  'applicationTemplateInfo'
];

function resolveTemplatePath(job, workingFolder)
{
  return path.resolve(workingFolder, job.applicationTemplateInfo.filePath);
}

/**
 * Validate a job requesting an application template prior to template expansion.
 * @param {Job} job - the job that needs to be validated
 * @param {string} workingFolder - folder from which the original job was loaded (if any)
 */
batchTemplateUtils.validateJobRequestingApplicationTemplate = function(job, workingFolder, _)
{
  // Rule: If job doesn't request an application template, don't validate further
  if (__.isUndefined(job.applicationTemplateInfo)) {
    return;
  }

  // Rule: Job must specify a template to use
  if (__.isUndefined(job.applicationTemplateInfo.filePath)) {
    throw new Error('No filePath specified for requested application template (define applicationTemplateInfo.filePath and try again).');
  }

  // Rule: Template file must exist
  // (We do this in order to give a good diagnostic in the most common case, knowing that this is 
  // technically a race condition because someone could delete the file between our check here and 
  // reading the file later on. We expect such cases to be rare.)
  try {
    const templateFilePath = resolveTemplatePath(job, workingFolder);
    fs.access(templateFilePath, fs.constants.R_OK, _);
  } catch (error) {
    throw new Error(util.format('Unable to read the template %s (%s).', job.applicationTemplateInfo.filePath, error));      
  }
  
  // Rule: Jobs may not use properties reserved for template use
  const usedButReserved = __.intersection( __.keys(job), batchTemplateUtils.propertiesReservedForTemplateUse);
  if (usedButReserved.length > 0) {
    throw new Error(util.format('Jobs using application templates may not use these properties: %s', usedButReserved));
  }
};

/**
 * Validate an application template prior to use.
 * @param template - the application template requiring validation 
 */
batchTemplateUtils.validateApplicationTemplate = function (template)
{
  // Rule: Templates may not use properties reserved for job use
  const usedButReserved = __.intersection( __.keys(template), batchTemplateUtils.propertiesReservedForJobUse);
  if (usedButReserved.length > 0) {
    throw new Error(util.format('Application templates may not use these job properties: %s', usedButReserved));
  }

  // Rule: Templates may only specify properties permitted
  const unsupported = __.difference( __.keys(template), batchTemplateUtils.propertiesPermittedOnTemplates);
  if (unsupported.length > 0) {
    throw new Error(util.format('Application templates may not use these properties: %s', unsupported));
  }

  // Rule: Every parameter declared on a template must specify one of the supported types
  const supportedTypes = ['int', 'string', 'bool'];
  if (!__.isUndefined( template.parameters)) {
    for (var name in template.parameters) {
      const parameter = template.parameters[name];
      if (__.isUndefined(parameter.type)) {
        throw new Error(util.format('The parameter \'%s\' does not specify a type.', name));
      }
      if (!__.contains(supportedTypes, parameter.type))
      {
        throw new Error(util.format('The parameter \'%s\' specifies unsupported type \'%s\'.', name, parameter.type));
      }
    }
  }
};

/**
 * Validate the partial job generated from an application template prior to merging it with the original job.
 * @param job - partial job requiring validation
 */
batchTemplateUtils.validateGeneratedJob = function(job) {
  // Rule: The job generated by an application template may not use properties reserved for job use
  // (This is a safety to prevent clever abuse of template syntax to specify things that shouldn't be.)
  const usedButReserved = __.intersection( __.keys(job), batchTemplateUtils.propertiesReservedForJobUse);
  if (usedButReserved.length > 0) {
    throw new Error(util.format('Application templates may not specify these properties: %s', usedButReserved));
  }
};

/**
 * Validate the parameters supplied by the job against those defined on the template
 * @param {object} parameters - parameters supplied by the job
 * @param {ojbect} definitions - parameter definitions from the application template 
 */
batchTemplateUtils.validateParameterUsage = function(parameters, definitions) {
  if (__.isUndefined(parameters)) {
    parameters = { };
  }

  if (__.isUndefined(definitions)) {
    definitions = { };
  }

  for (var name in definitions) {
    const definition = definitions[name];
    const parameter = parameters[name];

    // Rule: If the parameter definition has no default value, the template must provide a value
    if (__.isUndefined(definition.defaultValue) && __.isUndefined(parameter)) {
      throw new Error(util.format('A value for parameter \'%s\' must be provided by the job', name));
    }

    // Rule: If the parameter definition specifies 'int', the value provided must be compatible
    if (definition.type === 'int') {
      if (!__.isUndefined(parameter)) {
        // Value given, make sure it's ok'
        if(!__.isNumber(parameter)) { 
          throw new Error(util.format('Value \'%s\' supplied for parameter \'%s\' must be an integer', parameter, name));
        }
      } else {
        // No value provided, check default value
        if(!__.isNumber(definition.defaultValue)) { 
          throw new Error(util.format('Value \'%s\' supplied for parameter \'%s\' must be an integer', parameter, name));
        }        
      }
    }

    // Rule: if the parameter definition specified 'bool', the value provided must be compatible
    if (definition.type === 'bool') {
      if (!__.isBoolean(parameter)) {
        throw new Error(util.format('Value \'%s\' supplied for parameter \'%s\' must be a boolean', parameter, name));              
      }
    }
  }

  // Rule: Only parameters values defined by the template are permitted
  var violations = __.difference( __.keys(parameters), __.keys(definitions));
  if (violations.length > 0) {
    throw new Error(util.format('Provided parameter(s) \'%s\' are not expected by the template.', violations));
  }
};

/**
 * Merge metadata from two different sources
 * @param {object} baseMetadata - a (possibly undefined) set of metadata 
 * @param {object} moreMetadata - more metadata (also possibly undefined)
 */
batchTemplateUtils.mergeMetadata = function(baseMetadata, moreMetadata) {
  var result = [];
  
  if (!__.isUndefined(baseMetadata)) {
    baseMetadata.forEach( function(item) {
      result.push(item);
    });
  }

  if (!__.isUndefined(moreMetadata)) {
    moreMetadata.forEach( function(item) { 
      var clash = __.findWhere(result, { name : item.name});
      if (!__.isUndefined(clash)) {
        throw new Error(util.format('May not have multiple definitions for metadata value \'%s\'.', item.name));        
      }
      result.push(item);
    });
  }

  return result;
};

/**
 * Validate the provided metadata is valid
 */
batchTemplateUtils.validateMetadata = function(metadata) {
  // Rule: The prefix az_batch is reserved for our use and can't be specified on job nor template
  var violation = metadata.find( function(item) { return item.name.startsWith('az_batch:'); });
  if (!__.isUndefined(violation)) {
    throw new Error(util.format('Metadata item \'%s\' cannot be used as the prefix \'az_batch:\' is reserved for batch use.', violation.name));
  }
};

/**
 * Expand an application template reference on a job, returning the modified job
 * @param {object} job - a job definition that may contain an application template reference.
 * @param {string} workingFolder - base folder for evaluation of relative paths (if required).
 * @param {callback} _ - Streamline callback function.
 */
batchTemplateUtils.expandApplicationTemplate = function(job, workingFolder, _) {

  if(__.isUndefined(job.applicationTemplateInfo)) {
    // No application template used; nothing to do
    return job;
  }

  batchTemplateUtils.validateJobRequestingApplicationTemplate(job, workingFolder, _);

  // Retrieve the application template and expand it with the available parameters
  const templateFilePath = resolveTemplatePath(job, workingFolder);
  const templateAsString = fs.readFileSync(templateFilePath).toString();
  var templateAsLoaded;
  try {
    templateAsLoaded = JSON.parse(templateAsString);  
  } catch (error) {
    throw new Error(util.format('Failed to parse JSON loaded from %s (%s)', templateFilePath, error.message));
  }
  

  batchTemplateUtils.validateApplicationTemplate(templateAsLoaded);
  batchTemplateUtils.validateParameterUsage(job.applicationTemplateInfo.parameters, templateAsLoaded.parameters);

  const jobFromTemplate = batchTemplateUtils.parseTemplate(templateAsString, templateAsLoaded, job.applicationTemplateInfo.parameters);
  const metadata = batchTemplateUtils.mergeMetadata( jobFromTemplate.metadata, job.metadata);
  batchTemplateUtils.validateMetadata(metadata);

  metadata.push( { 
    name: 'az_batch:template_filepath',
    value : templateFilePath
  });
  
  // Safety checks that "ambitious" use of templating features haven't allowed someone to bypass the rules
  batchTemplateUtils.validateGeneratedJob(jobFromTemplate);

  // Merge the job as defined by the application template with the original job we were given
  const expandedJob = __.extend(jobFromTemplate, job);
  const cleanedJob = __.omit( expandedJob, 'applicationTemplateInfo', 'templateMetadata', 'parameters');
  cleanedJob.metadata = metadata;
  return cleanedJob;
};

/**
 * Validate the input parameter is valid for specified template. The following check will perform:
 *      Check input fit with parameter type, if yes, convert to correct type
 *      Check input matched with the restriction of parameter
 * Return true if input parameter is valid, otherwise, return false
 * @param {object} cli - The cli class is used to write message to console if provided
 * @param {string} parameterName - The name of the parameter.
 * @param {object} parameterContent - The definition of parameter in the template.
 * @param {object} parameterValues - key/value pair of user provided parameter values, the value may change after calling this function.
 */
batchTemplateUtils.validateParameter = function (cli, parameterName, parameterContent, parameterValues) {
  var parameterValue = parameterValues[parameterName]; 
  if (parameterContent.type === 'int') {
    // Check whether input is integer type or string can convert to integer type
    var x = parseInt(parameterValue);
    if (isNaN(parameterValue) || String(x) !== String(parameterValue)) {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The value \'%s\' of parameter \'%s\' is not an integer', parameterValue, parameterName));
      }
      return false;
    }
    // Update the value to be number type
    parameterValues[parameterName] = parseInt(parameterValue);
  }
  if (parameterContent.type === 'bool') {
    // Check whether input is boolean type or string can convert to boolean type
    if (typeof(parameterValue) !== 'boolean' && parameterValue !== 'true' && parameterValue !== 'false') {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The value \'%s\' of parameter \'%s\' is not a Boolean', parameterValue, parameterName));
      }
      return false;
    }
    if (typeof(parameterValue) !== 'boolean') {
      parameterValues[parameterName] = (parameterValue === 'true');
    }
  }
  if (parameterContent.type === 'string') {
    // Convert to string type
    parameterValues[parameterName] = String(parameterValue);
  }
  parameterValue = parameterValues[parameterName];
  if (!__.isUndefined(parameterContent.allowedValues)) {
    if (!__.contains(parameterContent.allowedValues, parameterValue)) {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The value %s of parameter \'%s\' is not one of the allowed values: %s', parameterValue, parameterName, parameterContent.allowedValues.join(',')));
      }
      return false;
    }
  }
  if (!__.isUndefined(parameterContent.minValue) && parameterContent.type === 'int') {
    if (parameterValue < parameterContent.minValue) {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The value %s of parameter \'%s\' is smaller than minimum value: %s', parameterValue, parameterName, parameterContent.minValue));
      }
      return false;
    }
  }
  if (!__.isUndefined(parameterContent.maxValue) && parameterContent.type === 'int') {
    if (parameterValue > parameterContent.maxValue) {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The value %s of parameter \'%s\' is greater than maximum value: %s', parameterValue, parameterName, parameterContent.maxValue));
      }
      return false;
    }
  }
  if (!__.isUndefined(parameterContent.minLength) && parameterContent.type === 'string') {
    if (parameterValue.length < parameterContent.minLength) {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The length of value %s of parameter \'%s\' is smaller than minimum length: %s', parameterValue, parameterName, parameterContent.minLength));
      }
      return false;
    }
  }
  if (!__.isUndefined(parameterContent.maxLength) && parameterContent.type === 'string') {
    if (parameterValue.length > parameterContent.maxLength) {
      if (!__.isUndefined(cli)) {
        cli.output.warn(util.format('The length of value %s of parameter \'%s\' is greater than maximum length: %s', parameterValue, parameterName, parameterContent.maxLength));
      }
      return false;
    }
  }
  return true;
};

/**
 * Return all required parameter values for the specified template.
 * @param {object} cli                      The cli class is used to write message to console, or get user input
 * @param {object} templateContent          contents of the template as JSON object
 * @param {object} userParameterValues      key/value pair of user provided parameter values
 */
batchTemplateUtils.getTemplateParams = function (cli, templateContent, userParameterValues, _) {
  var paramKeys = {};
  if (templateContent.parameters) {
    Object.keys(templateContent.parameters).forEach_(_, 1, function (_, param) {
      if (__.isUndefined(templateContent.parameters[param].type)) {
        throw new Error(util.format('The parameter \'%s\' does not have type defined', param));
      }

      if (userParameterValues && userParameterValues[param]) {
        paramKeys[param] = userParameterValues[param].value;
      } else {
        paramKeys[param] = templateContent.parameters[param].defaultValue;
      }

      do
      {
        var promptLine;
        if (templateContent.parameters[param].metadata && templateContent.parameters[param].metadata.description) {
          promptLine = util.format('%s (%s):', param, templateContent.parameters[param].metadata.description);
        } else {
          promptLine = param + ': ';
        }
        paramKeys[param] = cli.interaction.promptIfNotGiven(promptLine, paramKeys[param], _);

        if (!batchTemplateUtils.validateParameter(cli, param, templateContent.parameters[param], paramKeys)) {
          paramKeys[param] = undefined;
        }
      } while (__.isUndefined(paramKeys[param]));
    });
  }
  return paramKeys;
};

/**
 * Return JSON object with with the parameters replaced 
 * @param {object} cli                      The cli class is used to write message to console, or get user input
 * @param {string} templateFile             Input template file name
 * @param {string} parameterFile            Template parameter file name (optional)
 */
batchTemplateUtils.expandTemplate = function (cli, templateFile, parameterFile, _) {
  var mergedJobTemplate = {};
  if (templateFile) {
    // parse the template file
    var templateJsonStr = utils.stripBOM(fs.readFileSync(templateFile));
    var templateJsonObj = JSON.parse(templateJsonStr);
    var parameterJsonStr = '';
    var parameterJsonObj = {};

    if (!parameterFile) {
      cli.output.warn('No parameter file supplied.');
    } else {
      // parse the parameter file and fill in the blanks
      parameterJsonStr = utils.stripBOM(fs.readFileSync(parameterFile));
      parameterJsonObj = JSON.parse(parameterJsonStr);
    }

    // Make sure all parameters are filled
    var params = batchTemplateUtils.getTemplateParams(cli, templateJsonObj, parameterJsonObj, _);
    mergedJobTemplate = batchTemplateUtils.parseTemplate(templateJsonStr, templateJsonObj, params);
  }

  return mergedJobTemplate;
};

// Build the installation command line for package reference collection
function getInstallationCmdLine(references) {
  if (__.isUndefined(references) || __.isEmpty(references)) {
    return undefined;
  }

  var isApt = false;
  var isChoco = false;
  var isYum = false;
  var builder = '';

  references.forEach(function (reference) {

    if (__.isUndefined(reference.type) || __.isUndefined(reference.id)) {
      throw new Error('A PackageReference must have a type and id element.');
    }

    if (reference.type === 'aptPackage') {
      if (isChoco || isYum) {
        throw new Error('PackageReferences can only contain a single type of package reference.');
      }

      var aptCmd = '';

      if (reference.version) {
        aptCmd = util.format('=%s', reference.version);
      }
      aptCmd = util.format('apt-get install -y %s', reference.id) + aptCmd;
      if (isApt) {
        builder = builder + ';' + aptCmd; 
      } else {
        builder = aptCmd;
        isApt = true;
      }

      // TODO: deal with repository, keyUrl, sourceLine
      
    } else if (reference.type === 'chocolateyPackage') {
      if (isApt || isYum) {
        throw new Error('PackageReferences can only contain single type of package reference.');
      }

      var chocoCmd = '';

      if (reference.allowEmptyChecksums) {
        chocoCmd = ' --allow-empty-checksums';
      }
      if (reference.version) {
        chocoCmd = util.format(' --version %s', reference.version) + chocoCmd;
      }
      chocoCmd = util.format('choco install %s', reference.id) + chocoCmd;
      if (isChoco) {
        builder = builder + ' & ' + chocoCmd; 
      } else {
        builder = chocoCmd;
        isChoco = true;
      }
    } else if (reference.type === 'yumPackage') {
      if (isApt || isChoco) {
        throw new Error('PackageReferences can only contain single type of package reference.');
      }

      var yumCmd = '';

      if (reference.disableExcludes) {
        yumCmd = util.format(' --disableexcludes=%s', reference.disableExcludes);
      }
      if (reference.version) {
        yumCmd = util.format('-%s', reference.version) + yumCmd;
      }
      yumCmd = util.format('yum -y install %s', reference.id) + yumCmd;
      if (isYum) {
        builder = builder + ';' + yumCmd; 
      } else {
        builder = yumCmd;
        isYum = true;
      }
      
      // TODO: deal with rpmRepository
      // rpm -Uvh <rpmRepository> 

    } else if (reference.type === 'applicationPackage') {
      throw new Error(util.format('ApplicationPackage type for id %s is not supported in this version.', reference.id));
    } else {
      throw new Error(util.format('Unknown PackageReference type %s for id %s.', reference.type, reference.id));
    }
  });  

  var command = '';
  var isWindows = false;
  if (isApt) {
    command = 'apt-get update;' + builder;
  } else if (isChoco) {
    command = 'powershell -NoProfile -ExecutionPolicy unrestricted -Command "(iex ((new-object net.webclient).DownloadString(\'https://chocolatey.org/install.ps1\')))" && SET PATH="%PATH%;%ALLUSERSPROFILE%\\chocolatey\\bin"';
    command += ' && choco feature enable -n=allowGlobalConfirmation & ' + builder;
    isWindows = true;
    // TODO: Do we need to double check with pool agent name
  } else if (isYum) {
    command = builder;
  }
  
  return { cmdLine: command, isWindows: isWindows };
}

/**
 * Parse package reference section in the pool JSON object 
 * @param {object} pool                  Pool object
 */
batchTemplateUtils.parsePoolPackageReferences = function (pool) {
  if (!__.isArray(pool.packageReferences)) {
    throw new Error('PackageReferences of Pool has to be a collection.');
  }

  return getInstallationCmdLine(pool.packageReferences);
};

/**
 * Parse package reference section in the task JSON object
 * @param {object} job                  Job object 
 * @param {object} tasks                Task object collection
 */
batchTemplateUtils.parseTaskPackageReferences = function (job, tasks) {
  if (__.isUndefined(tasks) || __.isEmpty(tasks)) {
    return undefined;
  }

  var packages = [];
  tasks.forEach(function (task) {
    if (task.packageReferences) {
      packages = __.uniq(__.union(packages, task.packageReferences), function(p) {
        if (__.isUndefined(p.type) || __.isUndefined(p.id)) {
          throw new Error('A PackageReference must have a type and id element.');
        }
        return p.id;
      });
      delete task.packageReferences;
    }
  });

  return getInstallationCmdLine(packages);
};

/**
 * Post task expansion to process resource file and output file 
 * @param {object} request                  Job or task object
 */
batchTemplateUtils.postProcessing= function (request, _) {
  //Reform all new resource file references in standard ResourceFiles
  if (util.isArray(request)) {
    var newRequest = [];
    for (var item in request) {
      newRequest.push(batchTemplateUtils.processResourceFiles(request[item], _));
    }
    return newRequest;
  }
  request = batchTemplateUtils.processResourceFiles(request, _);
  return request;
};

module.exports = batchTemplateUtils;
