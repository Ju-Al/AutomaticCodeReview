    .option('--select <fields>', $('select fields to show, call without this parameter to see all fields'))
    .execute(function (nodeName, applicationName, options, _) {
      var progress = cli.interaction.progress($('ShowDeployedServicePackage'));
      
      try {
        var config = serviceFabricUtils.readServiceFabricConfig(progress, _);
        
        var client = new serviceFabricClient('3.0-preview', serviceFabricUtils.createConnectionUrl(config, _), serviceFabricUtils.getClientOptions(config, _));
        applicationName = serviceFabricUtils.parseUrl(applicationName, _);
        var res = client.deployedServicePackages.get(nodeName, applicationName, options, _);
        if (options.select) {
          res = serviceFabricUtils.pick(res, options.select, _);
        }
        serviceFabricUtils.setServiceEnumVal(res);
        
        progress.end();
        
        cli.interaction.formatOutput(res, function (data) {
          log.json(data);
        });
      } catch (e) {
        progress.end();
        throw e;
      }
    });
  
  var serviceManifest = service.category('manifest')
    .description($('Commands to manage your service manifest'));
  
  serviceManifest.command('show [applicationTypeName] [applicationTypeVersion] [serviceManifestName]')
    .description($('Show service manifest'))
    .option('-n --application-type-name <applicationTypeName>', $('the type of the application type'))
    .option('-t --application-type-version <applicationTypeVersion>', $('the version of the application type'))
    .option('-m --service-manifest-name <serviceManifestName>', $('the name of the service manifest'))
    .execute(function (applicationTypeName, applicationTypeVersion, serviceManifestName, options, _) {
      var progress = cli.interaction.progress($('Show service manifest'));
      
      try {
        var config = serviceFabricUtils.readServiceFabricConfig(progress, _);
        
        var client = new serviceFabricClient('3.0-preview', serviceFabricUtils.createConnectionUrl(config, _), serviceFabricUtils.getClientOptions(config, _));
        var res = client.serviceManifests.get(applicationTypeName, applicationTypeVersion, serviceManifestName, options, _);
        if (options.select) {
          res = serviceFabricUtils.pick(res, options.select, _);
        }
        serviceFabricUtils.setServiceEnumVal(res);
        
        progress.end();
        
        cli.interaction.formatOutput(res, function (data) {
          log.json(data);
        });
      } catch (e) {
        progress.end();
        throw e;
      }
    });
  
  var serviceDescription = service.category('description')
    .description($('Commands to manage your service'));
  
  serviceDescription.command('show [serviceName]')
    .description($('Show service description'))
    .option('-n --service-name <serviceName>', $('the name of the service'))
    .option('--select <fields>', $('select fields to show, call without this parameter to see all fields'))
    .execute(function (serviceName, options, _) {
      var progress = cli.interaction.progress($('Show service description'));
      
      try {
        var config = serviceFabricUtils.readServiceFabricConfig(progress, _);
        
        var client = new serviceFabricClient('3.0-preview', serviceFabricUtils.createConnectionUrl(config, _), serviceFabricUtils.getClientOptions(config, _));
        serviceName = serviceFabricUtils.parseUrl(serviceName, _);
        var res = client.serviceDescriptions.get(serviceName, options, _);
        if (options.select) {
          res = serviceFabricUtils.pick(res, options.select, _);
        }
        serviceFabricUtils.setServiceEnumVal(res);
        
        progress.end();
        
        cli.interaction.formatOutput(res, function (data) {
          log.json(data);
        });
      } catch (e) {
        progress.end();
        throw e;
      }
    });
  
  var serviceHealth = service.category('health')
    .description($('Commands to manage your service'));
  
  serviceHealth.command('show [serviceName]')
    .description($('Show service health'))
    .option('-n --service-name <serviceName>', $('the name of the service'))
    .option('--select <fields>', $('select fields to show, call without this parameter to see all fields'))
    .execute(function (serviceName, options, _) {
      var progress = cli.interaction.progress($('Show service description'));
      
      try {
        var config = serviceFabricUtils.readServiceFabricConfig(progress, _);
        
        var client = new serviceFabricClient('3.0-preview', serviceFabricUtils.createConnectionUrl(config, _), serviceFabricUtils.getClientOptions(config, _));
        serviceName = serviceFabricUtils.parseUrl(serviceName, _);
        var res = client.serviceHealths.get(serviceName, options, _);
        if (options.select) {
          res = serviceFabricUtils.pick(res, options.select, _);
        }
        serviceFabricUtils.setServiceEnumVal(res);
        
        progress.end();
        
        cli.interaction.formatOutput(res, function (data) {
          log.json(data);
        });
      } catch (e) {
        progress.end();
        throw e;
      }
    });
  
  serviceHealth.command('send [serviceName] [sourceId] [property] [healthState] [description] [timeToLiveInMilliSeconds] [sequenceNumber] [removeWhenExpired]')
    .description($('Send service health, Example: azure servicefabric service health send --service-name fabric:/app/svc fabric:app --source-id monitor --property pc --health-state Ok --description healthy'))
    .option('-n --service-name <serviceName>', $('the id of the partition'))
    .option('-i --source-id <sourceId>', $('the id of the source'))
    .option('-p --property <property>', $('the property'))
    .option('-e --health-state <healthState>', $('the state of the health, values are [Ok|Warning|Error|Unknown]'))
    .option('-d --description <description>', $('the description'))
    .option('-m --time-to-live-in-milliseconds <timeToLiveInMilliSeconds>', $('the time in milliseconds for live'))
    .option('-q --sequence-number <sequenceNumber>', $('the number of the sequence'))
    .option('-w --remove-when-expired <removeWhenExpired>', $('the boolean of the remove when expired'))
    .execute(function (serviceName, sourceId, property, healthState, description, timeToLiveInMilliseconds, sequenceNumber, removeWhenExpired, options, _) {
      var progress = cli.interaction.progress($('Send service health'));
      
      try {
        var config = serviceFabricUtils.readServiceFabricConfig(progress, _);
        
        var client = new serviceFabricClient('3.0-preview', serviceFabricUtils.createConnectionUrl(config, _), serviceFabricUtils.getClientOptions(config, _));
        serviceName = serviceFabricUtils.parseUrl(serviceName, _);
        var serviceHealthReport = {};
        if (sourceId) serviceHealthReport.sourceId = sourceId;
        if (property) serviceHealthReport.property = property;
        if (healthState) serviceHealthReport.healthState = serviceFabricUtils.getEnumVal('healthState', healthState);
        if (description) serviceHealthReport.description = description;
        if (timeToLiveInMilliseconds) serviceHealthReport.timeToLiveInMilliSeconds = timeToLiveInMilliseconds;
        if (sequenceNumber) serviceHealthReport.sequenceNumber = sequenceNumber;
        if (removeWhenExpired) serviceHealthReport.removeWhenExpired = removeWhenExpired;
        var res = client.serviceHealths.send(serviceName, serviceHealthReport, options, _);
        
        progress.end();
        
        cli.interaction.formatOutput(res, function (data) {
          log.json(data);
        });
      } catch (e) {
        progress.end();
        throw e;
      }
    });
};
