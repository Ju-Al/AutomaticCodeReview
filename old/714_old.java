		implements AsymmetricMeter, SymmetricMeter, OpenemsComponent {

	@Reference
	protected ConfigurationAdmin cm;

	public static final int DEFAULT_UNIT_ID = 1;

	@Reference(policy = ReferencePolicy.STATIC, policyOption = ReferencePolicyOption.GREEDY, cardinality = ReferenceCardinality.MANDATORY)
	protected void setModbus(BridgeModbus modbus) {
		super.setModbus(modbus);
	}

	public enum ChannelId implements io.openems.edge.common.channel.ChannelId {
		;
		private final Doc doc;

		private ChannelId(Doc doc) {
			this.doc = doc;
		}

		@Override
		public Doc doc() {
			return this.doc;
		}
	}

	public FeneconGoodwePvMeter() {
		super(//
				OpenemsComponent.ChannelId.values(), //
				AsymmetricMeter.ChannelId.values(), //
				SymmetricMeter.ChannelId.values(), ChannelId.values() //
		);
	}

	@Activate
	void activate(ComponentContext context, Config config) {
		super.activate(context, config.id(), config.alias(), config.enabled(), config.unit_id(), this.cm, "Modbus",
				config.modbus_id());
	}

	@Deactivate
	protected void deactivate() {
		super.deactivate();
	}

	@Override
	protected ModbusProtocol defineModbusProtocol() {
		return new ModbusProtocol(this, //
				new FC3ReadRegistersTask(36005, Priority.HIGH, //
						m(AsymmetricMeter.ChannelId.ACTIVE_POWER_L1, new SignedWordElement(36005)), //
						m(AsymmetricMeter.ChannelId.ACTIVE_POWER_L2, new SignedWordElement(36006)), //
						m(AsymmetricMeter.ChannelId.ACTIVE_POWER_L3, new SignedWordElement(36007)), //
						m(SymmetricMeter.ChannelId.ACTIVE_POWER, new SignedWordElement(36008)), //
						m(SymmetricMeter.ChannelId.REACTIVE_POWER, new UnsignedWordElement(36009))) //
		);
	}

	@Override
	public MeterType getMeterType() {
		return MeterType.PRODUCTION;
	}

	@Override
	public String debugLog() {
		return "L:" + this.getActivePower().value().asString();
	}

}
