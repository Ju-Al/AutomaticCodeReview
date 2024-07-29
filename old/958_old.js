                                   {$set: {userId: winningUserId}}, {multi: true});

      // Force all grains to shut down.
      grains.map(function (grain) {
        return backend.shutdownGrain(grain._id, losingUser._id, false);
      }).forEach(function (promise) {
        waitPromise(promise);
      });

      // Transfer grain storage to new owner.
      // Note: We don't parallelize this because it can cause some contention in the Blackrock
      //   back-end.
      grains.forEach(function (grain) {
        return waitPromise(backend.cap().transferGrain(grain.userId, grain._id, winningUserId));
      });

      var result = Meteor.users.update(losingUser._id, {$unset: {devName: 1, expires: 1},
                                                        $set: {"merging.status": "done"}});

    },
  });

  Meteor.users.find({"unmerging.status": "pending"}).observe({
    added: function(sourceUser) {
      console.log("unmerging user " + JSON.stringify(sourceUser));
      var destUserId = sourceUser.unmerging.destinationUserId;
      var servicesSetter = {};
      for (var key in sourceUser.unmerging.sourceServices) {
        servicesSetter["services." + key] = sourceUser.unmerging.sourceServices[key];
      }

      var destUser = Meteor.users.findOne(destUserId);
      if (destUser.merging) {
        Meteor.users.update(destUserId,
                            {$push: {identities: {$each: sourceUser.unmerging.sourceIdentities}},
                             $unset: {merging: 1},
                             $set: servicesSetter});
      }

      var sourceIdentityIds = sourceUser.unmerging.sourceIdentities.map(function (x) { return x.id; });

      var grains = Grains.find({userId: sourceUser._id, identityId: {$in: sourceIdentityIds}}).fetch();

      db.collections.grains.update({userId: sourceUser._id, identityId: {$in: sourceIdentityIds}},
                                   {$set: {userId: destUserId}},
                                   {multi: true});

      // Force all grains to shut down.
      grains.map(function (grain) {
        return backend.shutdownGrain(grain._id, sourceUser._id, false);
      }).forEach(function (promise) {
        waitPromise(promise);
      });

      // Transfer grain storage to new owner.
      // Note: We don't parallelize this because it can cause some contention in the Blackrock
      //   back-end.
      grains.forEach(function (grain) {
        return waitPromise(backend.cap().transferGrain(grain.userId, grain._id, destUserId));
      });

      var result = Meteor.users.update(sourceUser._id, {$unset: {"unmerging": 1}});
    },
  });

}
