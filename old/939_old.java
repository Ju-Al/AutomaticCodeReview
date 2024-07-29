      scheduleNextCheck(ethPeer);
    }
  }

  protected void disconnectPeer(final EthPeer ethPeer) {
    LOG.debug(
        "Disconnecting from peer {} marked invalid by {}",
        ethPeer,
        peerValidator.getClass().getSimpleName());
    ethPeer.disconnect(peerValidator.getDisconnectReason(ethPeer));
  }

  protected void scheduleNextCheck(final EthPeer ethPeer) {
    Duration timeout = peerValidator.nextValidationCheckTimeout(ethPeer);
    ethContext.getScheduler().scheduleFutureTask(() -> checkPeer(ethPeer), timeout);
  }
}
