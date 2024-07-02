// @flow
import React, { Component, Fragment } from 'react';
import { observer } from 'mobx-react';
import { get } from 'lodash';
import { defineMessages, intlShape, FormattedMessage } from 'react-intl';
import SVGInline from 'react-svg-inline';
import classnames from 'classnames';
import { PopOver } from 'react-polymorph/lib/components/PopOver';
import Wallet from '../../../domains/Wallet';
import StakePool from '../../../domains/StakePool';
import { getColorFromRange, getSaturationColor } from '../../../utils/colors';
import adaIcon from '../../../assets/images/ada-symbol.inline.svg';
import { DECIMAL_PLACES_IN_ADA } from '../../../config/numbersConfig';
import { PoolPopOver } from '../widgets/PoolPopOver';
import styles from './WalletRow.scss';
import popOverThemeOverrides from './WalletRowPopOverOverrides.scss';
import LoadingSpinner from '../../widgets/LoadingSpinner';
import arrow from '../../../assets/images/collapse-arrow.inline.svg';
import hardwareWalletsIcon from '../../../assets/images/hardware-wallet/connect-ic.inline.svg';
import {
  IS_RANKING_DATA_AVAILABLE,
  IS_SATURATION_DATA_AVAILABLE,
} from '../../../config/stakingConfig';
import noDataDashBigImage from '../../../assets/images/no-data-dash-big.inline.svg';

const messages = defineMessages({
  walletAmount: {
    id: 'staking.delegationCenter.walletAmount',
    defaultMessage: '!!!{amount} ADA',
    description:
      'Amount of each wallet for the Delegation center body section.',
  },
  notDelegated: {
    id: 'staking.delegationCenter.notDelegated',
    defaultMessage: '!!!Undelegated',
    description: 'Undelegated label for the Delegation center body section.',
  },
  removeDelegation: {
    id: 'staking.delegationCenter.removeDelegation',
    defaultMessage: '!!!Undelegate',
    description:
      'Remove delegation label for the Delegation center body section.',
  },
  TooltipPoolTickerEpoch: {
    id: 'staking.delegationCenter.stakePoolTooltipTickerEpoch',
    defaultMessage: '!!!From epoch {fromEpoch}',
    description:
      'Delegated stake pool tooltip ticker for the Delegation center body section.',
  },
  TooltipPoolTickerEarningRewards: {
    id: 'staking.delegationCenter.stakePoolTooltipTickerEarningRewards',
    defaultMessage: '!!!Currently earning rewards',
    description:
      'Delegated stake pool tooltip ticker for the Delegation center body section.',
  },
  delegate: {
    id: 'staking.delegationCenter.delegate',
    defaultMessage: '!!!Delegate',
    description: 'Delegate label for the Delegation center body section.',
  },
  redelegate: {
    id: 'staking.delegationCenter.redelegate',
    defaultMessage: '!!!Redelegate',
    description: 'Redelegate label for the Delegation center body section.',
  },
  unknownStakePoolLabel: {
    id: 'staking.delegationCenter.unknownStakePoolLabel',
    defaultMessage: '!!!unknown',
    description:
      'unknown stake pool label for the Delegation center body section.',
  },
  syncingTooltipLabel: {
    id: 'staking.delegationCenter.syncingTooltipLabel',
    defaultMessage: '!!!Syncing {syncingProgress}%',
    description:
      'unknown stake pool label for the Delegation center body section.',
  },
});

type Props = {
  wallet: Wallet,
  delegatedStakePool?: ?StakePool,
  numberOfStakePools: number,
  numberOfRankedStakePools: number,
  onDelegate: Function,
  onUndelegate: Function,
  getStakePoolById: Function,
  nextEpochNumber: ?number,
  futureEpochNumber: ?number,
  onSelect?: Function,
  selectedPoolId?: ?number,
  isListActive?: boolean,
  currentTheme: string,
  onOpenExternalLink: Function,
  showWithSelectButton?: boolean,
  containerClassName: string,
  setListActive?: Function,
  listName?: string,
};

type WalletRowState = {
  highlightedPoolId: boolean,
};
const initialWalletRowState = {
  highlightedPoolId: false,
};
@observer
export default class WalletRow extends Component<Props, WalletRowState> {
  static contextTypes = {
    intl: intlShape.isRequired,
  };

  state = {
    ...initialWalletRowState,
  };

  stakePoolFirstTileRef: { current: null | HTMLDivElement };
  stakePoolAdaSymbolRef: { current: null | HTMLDivElement };

  constructor(props: Props) {
    super(props);

    this.stakePoolFirstTileRef = React.createRef();
    this.stakePoolAdaSymbolRef = React.createRef();
  }

  componentDidUpdate() {
    this.handleFirstTilePopOverStyle();
  }

  handleFirstTilePopOverStyle = () => {
    const {
      wallet: { id },
    } = this.props;
    const existingStyle = document.getElementById(`wallet-row-${id}-style`);
    const { current: firstTileDom } = this.stakePoolFirstTileRef;
    const { current: adaSymbolDom } = this.stakePoolAdaSymbolRef;

    if (!firstTileDom || !adaSymbolDom) {
      if (existingStyle) {
        existingStyle.remove();
      }
      return;
    }

    if (existingStyle) {
      return;
    }

    const firstTileDomRect = firstTileDom.getBoundingClientRect();
    const adaSymbolDomRect = adaSymbolDom.getBoundingClientRect();
    const horizontalDelta =
      firstTileDomRect.width / 2 -
      adaSymbolDomRect.width / 2 -
      (adaSymbolDomRect.left - firstTileDomRect.left);

    const firstTilePopOverStyle = document.createElement('style');
    firstTilePopOverStyle.setAttribute('id', `wallet-row-${id}-style`);
    firstTilePopOverStyle.innerHTML = `.wallet-row-${id} .tippy-arrow { transform: translate(-${horizontalDelta}px, 0); }`;
    document.getElementsByTagName('head')[0].appendChild(firstTilePopOverStyle);
  };

  render() {
    const { intl } = this.context;
    const {
      wallet: {
        name,
        amount,
        isRestoring,
        syncState,
        delegatedStakePoolId,
        isHardwareWallet,
        id,
      },
      delegatedStakePool,
      numberOfRankedStakePools,
      onDelegate,
      onUndelegate,
      nextEpochNumber,
      futureEpochNumber,
      currentTheme,
      onOpenExternalLink,
      showWithSelectButton,
      containerClassName,
    } = this.props;

    // @TODO - remove once quit stake pool delegation is connected with rewards balance
    const isUndelegateBlocked = true;

    const syncingProgress = get(syncState, 'progress.quantity', '');
    const notDelegatedText = intl.formatMessage(messages.notDelegated);
    const removeDelegationText = intl.formatMessage(messages.removeDelegation);
    const delegateText = intl.formatMessage(messages.delegate);
    const redelegateText = intl.formatMessage(messages.redelegate);

    const {
      stakePoolId: nextPendingDelegationStakePoolId,
      stakePool: nextPendingDelegationStakePool,
    } = this.getPendingStakePool(nextEpochNumber || 0);

    const {
      stakePoolId: futurePendingDelegationStakePoolId,
      stakePool: futurePendingDelegationStakePool,
    } = this.getPendingStakePool(
      futureEpochNumber || 0,
      nextPendingDelegationStakePool
    );

    const { highlightedPoolId } = this.state;

    const stakePoolRankingColor = futurePendingDelegationStakePool
      ? getColorFromRange(
          futurePendingDelegationStakePool.ranking,
          numberOfRankedStakePools
        )
      : '';

    const saturationStyles = classnames([
      styles.saturationBar,
      futurePendingDelegationStakePool
        ? styles[
            getSaturationColor(futurePendingDelegationStakePool.saturation)
          ]
        : null,
    ]);

    const futureStakePoolTileStyles = classnames([
      styles.stakePoolTile,
      futurePendingDelegationStakePoolId && futurePendingDelegationStakePool
        ? styles.futureStakePoolTileDelegated
        : styles.futureStakePoolTileUndelegated,
      futurePendingDelegationStakePoolId && !futurePendingDelegationStakePool
        ? styles.futureStakePoolTileUndefined
        : null,
    ]);

    const rightContainerStyles = classnames([
      styles.right,
      isRestoring ? styles.isRestoring : null,
    ]);

    const actionButtonStyles = classnames([
      styles.action,
      highlightedPoolId ? styles.active : null,
    ]);

    return (
      <div className={styles.component}>
        <div className={styles.left}>
          <div className={styles.title}>
            {name}
            {isHardwareWallet && (
              <SVGInline
                svg={hardwareWalletsIcon}
                className={styles.hardwareWalletsIcon}
              />
            )}
          </div>
          <div className={styles.description}>
            {!isRestoring ? (
              <FormattedMessage
                {...messages.walletAmount}
                values={{
                  amount: amount.toFormat(DECIMAL_PLACES_IN_ADA),
                }}
              />
            ) : (
              '-'
            )}
          </div>
        </div>

        <div className={rightContainerStyles}>
          {!isRestoring ? (
            <Fragment>
              {delegatedStakePoolId ? (
                <PopOver
                  themeOverrides={popOverThemeOverrides}
                  className={`wallet-row-${id}`}
                  content={
                    <div className={styles.tooltipLabelWrapper}>
                      <span>
                        {intl.formatMessage(
                          messages.TooltipPoolTickerEarningRewards
                        )}
                      </span>
                    </div>
                  }
                >
                  <div
                    className={styles.stakePoolTile}
                    ref={this.stakePoolFirstTileRef}
                  >
                    <div
                      className={!delegatedStakePool ? styles.unknown : null}
                    >
                      {delegatedStakePool ? (
                        <div className={styles.stakePoolName}>
                          <div
                            className={styles.activeAdaSymbol}
                            ref={this.stakePoolAdaSymbolRef}
                          >
                            <SVGInline svg={adaIcon} />
                          </div>
                          <div className={styles.stakePoolTicker}>
                            {delegatedStakePool.ticker}
                          </div>
                        </div>
                      ) : (
                        <div className={styles.stakePoolUnknown}>
                          {intl.formatMessage(messages.unknownStakePoolLabel)}
                        </div>
                      )}
                    </div>
                  </div>
                </PopOver>
              ) : (
                <div className={styles.stakePoolTile}>
                  <div className={styles.nonDelegatedText}>
                    {notDelegatedText}
                  </div>
                </div>
              )}
              <SVGInline svg={arrow} className={styles.arrow} />
              <div className={styles.stakePoolTile}>
                {nextPendingDelegationStakePoolId ? (
                  <div
                    className={
                      !nextPendingDelegationStakePool ? styles.unknown : null
                    }
                  >
                    {nextPendingDelegationStakePool ? (
                      <div className={styles.stakePoolTicker}>
                        {nextPendingDelegationStakePool.ticker}
                      </div>
                    ) : (
                      <div className={styles.stakePoolUnknown}>
                        {intl.formatMessage(messages.unknownStakePoolLabel)}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className={styles.nonDelegatedText}>
                    {notDelegatedText}
                  </div>
                )}
              </div>
                  <div
                    onMouseEnter={this.handleShowTooltip}
                    onMouseLeave={this.handleHideTooltip}
                  >
                      <PopOver
                        key="stakePoolTooltip"
                        placement="auto"
                        maxWidth={280}
                        appendTo={'parent'}
                        isShowingOnHover={false}
                        isVisible={highlightedPoolId}
                        themeVariables={popOverThemeVariables}
                        allowHTML
                        content={
                          <TooltipPool
                            stakePool={futurePendingDelegationStakePool}
                            isVisible
                            currentTheme={currentTheme}
                            onClick={this.handleHideTooltip}
                            onOpenExternalLink={onOpenExternalLink}
                            top={top}
                            left={left}
                            color={stakePoolRankingColor}
                            showWithSelectButton={showWithSelectButton}
                            containerClassName={containerClassName}
                            numberOfRankedStakePools={numberOfRankedStakePools}
                            isDelegationView
                          />
              <SVGInline svg={arrow} className={styles.arrow} />
              <div className={futureStakePoolTileStyles}>
                {futurePendingDelegationStakePoolId ? (
                  <>
                    {futurePendingDelegationStakePool ? (
                      <PoolPopOver
                        openOnHover
                        color={stakePoolRankingColor}
                        currentTheme={currentTheme}
                        onClose={() =>
                          this.setState({
                            highlightedPoolId: false,
                          })
                        }
                        onOpen={() =>
                          this.setState({
                            highlightedPoolId: true,
                          })
                        }
                        onOpenExternalLink={onOpenExternalLink}
                        openWithDelay={false}
                        stakePool={futurePendingDelegationStakePool}
                        containerClassName={containerClassName}
                        numberOfRankedStakePools={numberOfRankedStakePools}
                        showWithSelectButton={showWithSelectButton}
                      >
                        <div className={styles.stakePoolTicker}>
                          {futurePendingDelegationStakePool.ticker}
                        </div>
                        {IS_RANKING_DATA_AVAILABLE ? (
                          <div
                            className={styles.ranking}
                            style={{ color: stakePoolRankingColor }}
                          >
                            {futurePendingDelegationStakePool.nonMyopicMemberRewards ? (
                              futurePendingDelegationStakePool.ranking
                            ) : (
                              <>
                                {numberOfRankedStakePools + 1}
                                <sup>*</sup>
                              </>
                            )}
                          </div>
                        ) : (
                          <div className={styles.noDataDash}>
                            <SVGInline svg={noDataDashBigImage} />
                          </div>
                        )}
                        {IS_SATURATION_DATA_AVAILABLE && (
                          <div className={saturationStyles}>
                            <span
                              style={{
                                width: `${parseFloat(
                                  futurePendingDelegationStakePool.saturation
                                ).toFixed(2)}%`,
                              }}
                            />
                          </div>
                        )}
                        <div
                          className={styles.stakePoolRankingIndicator}
                          style={{ background: stakePoolRankingColor }}
                        />
                      </PoolPopOver>
                    ) : (
                      <div className={styles.stakePoolUnknown}>
                        {intl.formatMessage(messages.unknownStakePoolLabel)}
                      </div>
                    )}
                  </>
                ) : (
                  <div className={styles.nonDelegatedText}>
                    {notDelegatedText}
                  </div>
                )}
                {futurePendingDelegationStakePoolId && !isUndelegateBlocked && (
                  <div
                    className={actionButtonStyles}
                    role="presentation"
                    onClick={onUndelegate}
                    key="undelegate"
                  >
                    {removeDelegationText}
                  </div>
                )}
                <div
                  className={actionButtonStyles}
                  role="presentation"
                  onClick={onDelegate}
                >
                  {!futurePendingDelegationStakePoolId
                    ? delegateText
                    : redelegateText}
                </div>
              </div>
            </Fragment>
          ) : (
            <PopOver
              content={intl.formatMessage(messages.syncingTooltipLabel, {
                syncingProgress,
              })}
            >
              <LoadingSpinner medium />
            </PopOver>
          )}
        </div>
      </div>
    );
  }

  getPendingStakePool = (
    epochNumber: number,
    fallbackStakePool: ?StakePool
  ) => {
    const {
      wallet: { delegatedStakePoolId, pendingDelegations },
      delegatedStakePool,
      getStakePoolById,
    } = this.props;

    let stakePoolId;
    let stakePool;
    const hasPendingDelegations =
      pendingDelegations && pendingDelegations.length;

    if (hasPendingDelegations) {
      const pendingDelegation = pendingDelegations.filter(
        (item) => get(item, ['changes_at', 'epoch_number'], 0) === epochNumber
      );
      stakePoolId = get(pendingDelegation, '[0].target');
      stakePool = getStakePoolById(stakePoolId);
    }

    if (!stakePool && fallbackStakePool) {
      stakePoolId = fallbackStakePool.id;
      stakePool = fallbackStakePool;
    }

    if (!stakePool && delegatedStakePoolId) {
      stakePoolId = delegatedStakePoolId;
      stakePool = delegatedStakePool;
    }

    return {
      stakePoolId,
      stakePool,
    };
  };
}
