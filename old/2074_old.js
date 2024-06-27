import Component from '@ember/component';
import { computed } from '@ember/object';

export default Component.extend({});
export default Component.extend({
  classNameBindings: ['layoutClass'],

  layoutName: '',

  layoutClass: computed(function () {
    return this.layoutName.replace('layouts/', 'layout--');
  })
});