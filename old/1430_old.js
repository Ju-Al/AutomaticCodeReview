  flipVertical: boolean,
  selectedPoolId: ?number,
};

const initialState = {
  flipHorizontal: false,
  flipVertical: false,
  searchValue: '',
  selectedList: null,
  selectedPoolId: null,
};

export default class DelegationStepsChooseStakePoolDialog extends Component<
  Props,
  State
> {
  static contextTypes = {
    intl: intlShape.isRequired,
  };

  state = {
    ...initialState,
  };

  searchInput: ?HTMLElement = null;

  handleSearch = (searchValue: string) => this.setState({ searchValue });

  handleSelect = (selectedPoolId: number) => {
    this.setState({ selectedPoolId });
  };

  handleSetListActive = (selectedList: string) => {
    this.setState({ selectedList });
  };

  handleDeselectStakePool = () => {
    this.setState({ selectedPoolId: null });
  };

  render() {
    const { intl } = this.context;
    const {
      stepsList,
      stakePoolsDelegatingList,
      stakePoolsList,
      onOpenExternalLink,
      currentTheme,
      onClose,
      onContinue,
      onBack,
    } = this.props;
    const {
      searchValue,
      flipHorizontal,
      flipVertical,
      selectedList,
      selectedPoolId,
    } = this.state;

    const actions = [
      {
        className: 'continueButton',
        label: intl.formatMessage(messages.continueButtonLabel),
        onClick: onContinue,
        primary: true,
        disabled: !selectedPoolId,
      },
    ];

    const selectedPoolBlock = stakePoolId => {
      const selectedPool = find(
        stakePoolsList,
        stakePools => stakePools.id === stakePoolId
      );
      const blockLabel = get(
        selectedPool,
        'slug',
        intl.formatMessage(messages.selectPoolPlaceholder)
      );

      const selectedPoolBlockClasses = classNames([
        styles.selectedPoolBlock,
        selectedPool ? styles.selected : null,
      ]);

      return (
        <div
          role="presentation"
          className={selectedPoolBlockClasses}
          onClick={this.handleDeselectStakePool}
        >
          <div className={styles.label}>{blockLabel}</div>
          <div className={styles.checkmarkWrapper}>
            <SVGInline svg={checkmarkImage} className={styles.checkmarkImage} />
          </div>
        </div>
      );
    };

    const stepsIndicatorLabel = (
      <FormattedMessage
        {...messages.stepIndicatorLabel}
        values={{
          currentStep: 2,
          totalSteps: stepsList.length,
        }}
      />
    );

    return (
      <Dialog
        title={intl.formatMessage(messages.title)}
        subtitle={stepsIndicatorLabel}
        actions={actions}
        closeOnOverlayClick
        onClose={onClose}
        className={styles.delegationStepsChooseStakePoolDialogWrapper}
        closeButton={<DialogCloseButton onClose={onClose} />}
        backButton={<DialogBackButton onBack={onBack} />}
      >
        <div className={styles.delegationStepsIndicatorWrapper}>
          <Stepper
            steps={stepsList}
            activeStep={2}
            skin={StepperSkin}
            labelDisabled
          />
        </div>

        <div className={styles.content}>
          <p className={styles.description}>
            {intl.formatMessage(messages.description)}
          </p>
          <div className={styles.delegatedStakePoolsWrapper}>
            {selectedPoolBlock(selectedPoolId)}

            <div className={styles.delegatedStakePoolsList}>
              <p className={styles.stakePoolsDelegatingListLabel}>
                {intl.formatMessage(messages.delegatedPoolsLabel)}
              </p>
              <StakePoolsList
                listName="stakePoolsDelegatingList"
                flipHorizontal={flipHorizontal}
                flipVertical={flipVertical}
                stakePoolsList={stakePoolsDelegatingList}
                onOpenExternalLink={onOpenExternalLink}
                currentTheme={currentTheme}
                isListActive={selectedList === 'stakePoolsDelegatingList'}
                setListActive={this.handleSetListActive}
                containerClassName="Dialog_content"
                onSelect={this.handleSelect}
                selectedPoolId={selectedPoolId}
                showSelected
                highlightOnHover
              />
            </div>
          </div>

          <div className={styles.searchStakePoolsWrapper}>
            <StakePoolsSearch
              search={searchValue}
              label={intl.formatMessage(messages.searchInputLabel)}
              placeholder={intl.formatMessage(messages.searchInputPlaceholder)}
              onSearch={this.handleSearch}
              registerSearchInput={searchInput => {
                this.searchInput = searchInput;
              }}
            />
          </div>

          <div className={styles.stakePoolsListWrapper}>
            <StakePoolsList
              listName="selectedIndexList"
              flipHorizontal={flipHorizontal}
              flipVertical={flipVertical}
              stakePoolsList={stakePoolsList}
              onOpenExternalLink={onOpenExternalLink}
              currentTheme={currentTheme}
              isListActive={selectedList === 'selectedIndexList'}
              setListActive={this.handleSetListActive}
              onSelect={this.handleSelect}
              selectedPoolId={selectedPoolId}
              containerClassName="Dialog_content"
              showSelected
              highlightOnHover
            />
          </div>
        </div>
      </Dialog>
    );
  }
}
