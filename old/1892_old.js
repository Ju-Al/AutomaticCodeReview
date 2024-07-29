    return this.validFields.length === this.fields.length;
  }),

  isNotValid: not('isValid'),

  onSubmit() {},

  registerField(field) {
    this.fields.addObject(field);
  },

  unregisterField(field) {
    this.fields.removeObject(field);
  },

  validate() {
    this.fields.invoke('validate');
  },

  init() {
    this._super(...arguments);
    this.fields = [];
  },

  actions: {

    submit() {
      this.validate();
      if (this.isValid) {
        this.onSubmit();
      }
      return false;
    }

  }
});
