      );
    }

    const {
      id: stakePoolId,
      name: stakePoolName,
      slug: stakePoolSlug,
    } = delegatedStakePool;

    return (
      <UndelegateConfirmationDialog
        walletName={walletName}
        stakePoolName={stakePoolName}
        stakePoolSlug={stakePoolSlug}
        onConfirm={passphrase => {
          actions.wallets.undelegateWallet.trigger({
            walletId,
            stakePoolId,
            passphrase,
          });
        }}
        onCancel={() => {
          actions.dialogs.closeActiveDialog.trigger();
          quitStakePoolRequest.reset();
          actions.wallets.setUndelegateWalletSubmissionSuccess.trigger({
            result: false,
          });
        }}
        onExternalLinkClick={onExternalLinkClick}
        isSubmitting={quitStakePoolRequest.isExecuting}
        error={quitStakePoolRequest.error}
        fees={fees}
      />
    );
  }
}
