var EventEmitter = require('events');

function warnIfLegacy(eventName) {
  const legacyEvents = ['abi-vanila', 'abi', 'abi-contracts-vanila', 'abi-vanila-deployment'];
  if (legacyEvents.indexOf(eventName) >= 0) {
    console.info(__("this event is deprecated and will be removed in future versions %s", eventName));
  }
}

function log(eventType, eventName) {
  if (['end', 'prefinish', 'error', 'new', 'demo', 'block', 'version'].indexOf(eventName) >= 0) {
    return;
  }
  //console.log(eventType, eventName);
}

EventEmitter.prototype._maxListeners = 200;inversify.decorate(inversify.injectable(), Events);
inversify.decorate(inversify.injectable(), EventEmitter);

module.exports = Events;
