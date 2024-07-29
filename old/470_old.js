
    before(function (done) {
      suiteUtil = new MockedTestUtils(testPrefix, true);
      suiteUtil.setupSuite(done);
    });

    after(function (done) {
      suiteUtil.teardownSuite(done);
    });

    beforeEach(function (done) {
      suiteUtil.setupTest(done);
    });

    afterEach(function (done) {
      suiteUtil.teardownTest(done);
    });

    describe('server cmdlets', function () {
      var oldServiceNames;

      beforeEach(function (done) {
        var cmd = ('node cli.js service list --json').split(' ');
        executeCmd(cmd, function (result) {
          oldServiceNames = JSON.parse(result.text).map(function (service) {
            return service.ServiceName;
          });

          done();
        });
      });

      afterEach(function (done) {
        function deleteUsedServices (serviceNames) {
          if (serviceNames.length > 0) {
            var serviceName = serviceNames.pop();

            var cmd = ('node cli.js service delete ' + serviceName + ' --json').split(' ');
            executeCmd(cmd, function () {
              deleteUsedServices(serviceNames);
            });
          } else {
            done();
          }
        }

        var cmd = ('node cli.js service list --json').split(' ');
        executeCmd(cmd, function (result) {
          var services = JSON.parse(result.text);

          var usedServices = [ ];
          _.each(services, function (service) {
            if (!_.contains(oldServiceNames, service.ServiceName)) {
              usedServices.push(service.ServiceName);
            }
          });

          deleteUsedServices(usedServices);
        });
      });

      describe('Create Cloud Service', function () {
        it('should create a server', function (done) {
          var cloudServiceName = suiteUtil.generateId(createdServicesPrefix, createdServices);

          var cmd = ('node cli.js service create ' + cloudServiceName).split(' ');
          cmd.push('--location');
          cmd.push(location);
          cmd.push('--json');

          executeCmd(cmd, function (result) {
            result.text.should.not.be.null;
            result.exitStatus.should.equal(0);

            var serverName = JSON.parse(result.text).Name;
            serverName.should.not.be.null;
            serverName.should.match(/[0-9a-zA-Z]*/);

            done();
          });
        });
      });

      describe('List and show Cloud Service', function () {
        var cloudServiceName;

        beforeEach(function (done) {
          cloudServiceName = suiteUtil.generateId(createdServicesPrefix, createdServices);

          var cmd = ('node cli.js service create ' + cloudServiceName).split(' ');
          cmd.push('--location');
          cmd.push(location);
          cmd.push('--json');

          executeCmd(cmd, function (result) {
            result.text.should.not.be.null;
            result.exitStatus.should.equal(0);

            serviceName = JSON.parse(result.text).Name;
            serviceName.should.equal(cloudServiceName);

            done();
          });
        });

        it('should show the service', function (done) {
          var cmd = ('node cli.js service show ' + cloudServiceName).split(' ');
          cmd.push('--json');

          executeCmd(cmd, function (result) {
            result.text.should.not.be.null;
            result.exitStatus.should.equal(0);

            var service = JSON.parse(result.text);
            service.Location.should.equal(location);

            done();
          });
        });

        it('should list the service', function (done) {
          var cmd = ('node cli.js service list').split(' ');
          cmd.push('--json');

          executeCmd(cmd, function (result) {
            result.text.should.not.be.null;
            result.exitStatus.should.equal(0);

            var services = JSON.parse(result.text);

            should.exist(services.filter(function (service) {
              return service.Name === cloudServiceName;
            }));

            done();
          });
        });
      });
    });
  });
});