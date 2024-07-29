            });
    return signature_with_pubkey;
  }

  iroha::consensus::yac::VoteMessage FakePeer::makeVote(
      const iroha::consensus::yac::YacHash &yac_hash) {
    iroha::consensus::yac::YacHash my_yac_hash = yac_hash;
    my_yac_hash.block_signature = makeSignature(
        shared_model::crypto::Blob(yac_hash.vote_hashes.block_hash));
    return yac_crypto_->getVote(my_yac_hash);
  }

  void FakePeer::sendYacState(const std::vector<iroha::consensus::yac::VoteMessage> &state) {
    yac_transport_->sendState(*real_peer_, state);
  }

  void FakePeer::voteForTheSame(const YacStateMessage &incoming_votes) {
    using iroha::consensus::yac::VoteMessage;
    log_->debug("Got a YAC state message with {} votes.",
                incoming_votes->size());
    if (incoming_votes->size() > 1) {
      // TODO mboldyrev 24/10/2018: rework ignoring states for accepted commits
      log_->debug(
          "Ignoring state with multiple votes, "
          "because it probably refers to an accepted commit.");
      return;
    }
    std::vector<VoteMessage> my_votes;
    my_votes.reserve(incoming_votes->size());
    std::transform(
        incoming_votes->cbegin(),
        incoming_votes->cend(),
        std::back_inserter(my_votes),
        [this](const VoteMessage &incoming_vote) {
          log_->debug(
              "Sending agreement for proposal (Round ({}, {}), hash ({}, {})).",
              incoming_vote.hash.vote_round.block_round,
              incoming_vote.hash.vote_round.reject_round,
              incoming_vote.hash.vote_hashes.proposal_hash,
              incoming_vote.hash.vote_hashes.block_hash);
          return makeVote(incoming_vote.hash);
        });
    sendYacState(my_votes);
  }

}  // namespace integration_framework
