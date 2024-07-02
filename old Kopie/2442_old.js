import Component from '@ember/component';
import { reads } from '@ember/object/computed';

export default Component.extend({
  info: null,
  billingEmail: reads('info.billingEmail'),
  billingEmails: computed('billingEmail', function () {
    return (this.billingEmail || '').split(',');
  }),

  actions: {
    updateValues(values) {
      this.info.set('billingEmail', values.join(','));
    },
  }
});
