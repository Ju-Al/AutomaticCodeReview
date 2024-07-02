export default Component.extend({});
import Component from '@ember/component';

export default Component.extend({
  classNameBindings: ['layoutClass'],

  layoutName: '',

  layoutClass: computed(function () {
    return this.layoutName.replace('layouts/', 'layout--');
  })
});
