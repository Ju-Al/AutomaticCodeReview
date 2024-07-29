                                   sdk_pool_handle, sdk_wallet_steward,
                                   sdk_wallet_trustee, validUpgradeExpForceTrue):
    for node in txnPoolNodeSet[2:]:
        node.upgrader.check_upgrade_possible = lambda a, b, c: None
    for node in txnPoolNodeSet[:2]:
        node.upgrader.check_upgrade_possible = lambda a, b, c: 'some exception'

    with pytest.raises(RequestNackedException, match='some exception'):
        sdk_ensure_upgrade_sent(looper, sdk_pool_handle, sdk_wallet_trustee, validUpgradeExpForceTrue)

    sdk_add_new_nym(looper, sdk_pool_handle, sdk_wallet_steward)
