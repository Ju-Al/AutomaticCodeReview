    })).isRequired,
    currentLocale: PropTypes.string,
    onSelectLanguage: PropTypes.func.isRequired,
    isSubmitting: PropTypes.bool.isRequired,
    error: PropTypes.instanceOf(LocalizableError),
  };

  static contextTypes = {
    intl: intlShape.isRequired,
  };

  selectLanguage = (values: { locale: string }) => {
    this.props.onSelectLanguage({ locale: values });
  };

  form = new ReactToolboxMobxForm({
    fields: {
      languageId: {
        label: this.context.intl.formatMessage(messages.languageSelectLabel),
        value: this.props.currentLocale,
        bindings: 'ReactToolbox',
      }
    }
  }, {
    options: {
      validateOnChange: false,
    }
  });

  render() {
    const { languages, isSubmitting, error } = this.props;
    const { intl } = this.context;
    const { form } = this;
    const languageId = form.$('languageId');
    const languageOptions = languages.map(language => ({
      value: language.value,
      label: intl.formatMessage(language.label)
    }));
    const componentClassNames = classNames([styles.component, 'general']);
    const languageSelectClassNames = classNames([
      styles.language,
      isSubmitting ? styles.submitLanguageSpinner : null,
    ]);
    return (
      <div className={componentClassNames}>

        <Dropdown
          className={languageSelectClassNames}
          source={languageOptions}
          {...languageId.bind()}
          onChange={this.selectLanguage}
        />

        {error && <p className={styles.error}>{error}</p>}

      </div>
    );
  }

}
