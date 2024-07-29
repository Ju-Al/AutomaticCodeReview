
    LongStream.range(startBlock, endBlock)
        .forEach(value -> signerMetricResultList.add(generateBlock(value)));

    final JsonRpcRequest request = requestWithParams();

    final JsonRpcSuccessResponse response = (JsonRpcSuccessResponse) method.response(request);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> verify(blockchainQueries).getBlockHeaderByNumber(value));

    verify(blockchainQueries, atLeast(2)).headBlockNumber();

    signerMetricResultList.forEach(
        signerMetricResult ->
            assertThat(response.getResult().toString()).contains(signerMetricResult.toString()));
  }

  @Test
  public void getSignerMetrics() {

    Long startBlock = 1L;
    Long endBlock = 3L;

    List<SignerMetricResult> signerMetricResultList = new java.util.ArrayList<>();

    LongStream.range(startBlock, endBlock)
        .forEach(value -> signerMetricResultList.add(generateBlock(value)));

    final JsonRpcRequest request =
        requestWithParams(String.valueOf(startBlock), String.valueOf(endBlock));

    final JsonRpcSuccessResponse response = (JsonRpcSuccessResponse) method.response(request);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> verify(blockchainQueries).getBlockHeaderByNumber(value));

    verify(blockchainQueries, never()).headBlockNumber();

    signerMetricResultList.forEach(
        signerMetricResult ->
            assertThat(response.getResult().toString()).contains(signerMetricResult.toString()));
  }

  @Test
  public void getSignerMetricsWithLatest() {

    Long startBlock = 1L;
    Long endBlock = 3L;

    List<SignerMetricResult> signerMetricResultList = new java.util.ArrayList<>();

    when(blockchainQueries.headBlockNumber()).thenReturn(endBlock);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> signerMetricResultList.add(generateBlock(value)));

    final JsonRpcRequest request = requestWithParams(String.valueOf(startBlock), "latest");

    final JsonRpcSuccessResponse response = (JsonRpcSuccessResponse) method.response(request);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> verify(blockchainQueries).getBlockHeaderByNumber(value));

    verify(blockchainQueries).headBlockNumber();

    signerMetricResultList.forEach(
        signerMetricResult ->
            assertThat(response.getResult().toString()).contains(signerMetricResult.toString()));
  }

  @Test
  public void getSignerMetricsWithPending() {

    Long startBlock = 1L;
    Long endBlock = 3L;

    List<SignerMetricResult> signerMetricResultList = new java.util.ArrayList<>();

    when(blockchainQueries.headBlockNumber()).thenReturn(endBlock);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> signerMetricResultList.add(generateBlock(value)));

    final JsonRpcRequest request = requestWithParams(String.valueOf(startBlock), "pending");

    final JsonRpcSuccessResponse response = (JsonRpcSuccessResponse) method.response(request);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> verify(blockchainQueries).getBlockHeaderByNumber(value));

    verify(blockchainQueries).headBlockNumber();

    signerMetricResultList.forEach(
        signerMetricResult ->
            assertThat(response.getResult().toString()).contains(signerMetricResult.toString()));
  }

  @Test
  public void getSignerMetricsWithEarliest() {

    Long startBlock = 0L;
    Long endBlock = 3L;

    List<SignerMetricResult> signerMetricResultList = new java.util.ArrayList<>();

    LongStream.range(startBlock, endBlock)
        .forEach(value -> signerMetricResultList.add(generateBlock(value)));

    final JsonRpcRequest request = requestWithParams("earliest", String.valueOf(endBlock));

    final JsonRpcSuccessResponse response = (JsonRpcSuccessResponse) method.response(request);

    LongStream.range(startBlock, endBlock)
        .forEach(value -> verify(blockchainQueries).getBlockHeaderByNumber(value));

    verify(blockchainQueries, never()).headBlockNumber();

    signerMetricResultList.forEach(
        signerMetricResult ->
            assertThat(response.getResult().toString()).contains(signerMetricResult.toString()));
  }

  private JsonRpcRequest requestWithParams(final Object... params) {
    return new JsonRpcRequest(JSON_RPC_VERSION, CLIQUE_METHOD, params);
  }

  private SignerMetricResult generateBlock(final long number) {
    final SECP256K1.KeyPair proposerKeyPair = SECP256K1.KeyPair.generate();
    final Address proposerAddressBlock = Util.publicKeyToAddress(proposerKeyPair.getPublicKey());

    final BlockHeaderTestFixture builder = new BlockHeaderTestFixture();

    builder.number(number);
    builder.gasLimit(5000);
    builder.timestamp(6000 * number);
    builder.mixHash(
        Hash.fromHexString("0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365"));
    builder.ommersHash(Hash.EMPTY_LIST_HASH);
    builder.nonce(0);
    builder.difficulty(UInt256.ONE);
    builder.coinbase(Util.publicKeyToAddress(proposerKeyPair.getPublicKey()));

    final BlockHeader header =
        TestHelpers.createCliqueSignedBlockHeader(
            builder, proposerKeyPair, singletonList(proposerAddressBlock));

    when(blockchainQueries.getBlockHeaderByNumber(number)).thenReturn(Optional.of(header));
    when(cliqueBlockInterface.getProposerOfBlock(header)).thenReturn(proposerAddressBlock);

    SignerMetricResult signerMetricResult = new SignerMetricResult(proposerAddressBlock);
    signerMetricResult.incrementeNbBlock();
    signerMetricResult.setLastProposedBlockNumber(number);

    return signerMetricResult;
  }
}
